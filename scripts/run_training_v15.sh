#!/usr/bin/env bash
# Kompress v8: C3 + GLM scenarios — 983 pairs (largest yet) — train on everything bagel: 97 C3 Qwen + 286 GLM regex + 600 generic real tool outputs
#
# Hypothesis: v4 lost 0.008 heretic precision (0.975→0.967) from self-labeling noise.
# Using a stronger teacher (Qwen2.5-7B) to label must-keep spans on real tool outputs
# should recover v2-base precision while keeping override_delta=0.
#
# Data: 100-120 domain_train_2k texts labeled by Qwen2.5-7B via HF_INFER_PRO
# Base:  chopratejas/kompress-v2-base (the precision ceiling)
# Target: heretic >= 0.970, override_delta = 0, agent mk_in_ref >= 0.95
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v15"}

echo "=== 1/4 Convert Qwen labels to token keep/drop ==="
python3 scripts/convert_qwen_to_tokens.py \
    --input data/c3_qwen_labeled.jsonl \
    --output data/c3_train.jsonl \
    --min-keep-ratio 0.03 \
    --max-keep-ratio 0.75

count=$(wc -l < data/c3_train.jsonl)
echo "Training pairs: $count"
[ "$count" -gt 20 ] || { echo "ERROR: too few pairs ($count), need >20"; exit 1; }

echo "=== 2/4 Training data ready (pre-merged C3 + generic) ==="
python3 - << 'PY'
import json
# Pre-built merged file from repo (C3 Qwen-labeled + generic multi-train)
with open("data/kompress_v15_train.jsonl") as f:
    rows = [json.loads(l) for l in f]
from collections import Counter
sources = Counter(r.get("source", "?") for r in rows)
print(f"  Total: {len(rows)} rows from {len(sources)} sources")
for s, c in sources.most_common():
    print(f"    {s}: {c}")
PY

echo "=== 3/4 Fine-tune from v2-base (recover precision, don't inherit v4 noise) ==="
python3 scripts/train_kompress.py \
    --data data/kompress_v15_train.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v15-finetuned \
    --epochs 3 \
    --batch-size 16 \
    --lr 2e-5

echo "=== 4/4 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v15-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/4 ONNX export + HuggingFace upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v15-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")

class Wrapper(torch.nn.Module):
    def __init__(self, m):
        super().__init__()
        self.m = m
    def forward(self, input_ids, attention_mask):
        logits, span = self.m(input_ids, attention_mask)
        probs = torch.softmax(logits, dim=-1)[:, :, 1]
        return probs * (0.5 + 0.5 * span)

os.makedirs("kompress-v15-finetuned/onnx", exist_ok=True)
torch.onnx.export(
    Wrapper(model),
    (dummy["input_ids"], dummy["attention_mask"]),
    "kompress-v15-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids", "attention_mask"],
    output_names=["final_scores"],
    dynamic_axes={
        "input_ids": {0: "batch", 1: "seq"},
        "attention_mask": {0: "batch", 1: "seq"},
        "final_scores": {0: "batch", 1: "seq"},
    },
    opset_version=17,
)
print("ONNX exported")
PY

if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api = HfApi(token=os.environ["HF_TOKEN"])
repo = os.environ["HF_REPO"]
api.create_repo(repo, exist_ok=True, private=False)
api.upload_folder(
    folder_path="kompress-v15-finetuned",
    repo_id=repo,
    commit_message="kompress-v15: C3 + GLM scenarios — 983 pairs (largest yet) — Qwen2.5-7B teacher labels on real tool outputs"
)
print(f"Uploaded to {repo}")
PY
fi
echo "=== v8 complete ==="
