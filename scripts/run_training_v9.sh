#!/usr/bin/env bash
# Kompress v9: C3-pure — Qwen labels only, no generic data dilution
#
# v8 finding: C3 self-distillation works (heretic +0.012, mk_in_ref 1.000)
# but the 200 generic pairs (dolly/openorca/alpaca) may be dragging precision
# down from v2-base's 0.975 ceiling.
#
# v9: train on C3 Qwen-labeled data ONLY. If the generic data was noise,
# we should recover more of v2-base's precision.
#
# Data: 97 C3 Qwen-labeled pairs (no generic merge)
# Base:  chopratejas/kompress-v2-base
# Target: heretic >= 0.965, override_delta = 0
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v9"}

echo "=== 1/4 Convert Qwen labels to token keep/drop ==="
python3 scripts/convert_qwen_to_tokens.py \
    --input data/c3_qwen_labeled.jsonl \
    --output data/c3_train.jsonl \
    --min-keep-ratio 0.03 \
    --max-keep-ratio 0.75

count=$(wc -l < data/c3_train.jsonl)
echo "C3 training pairs: $count"
[ "$count" -gt 20 ] || { echo "ERROR: too few pairs ($count)"; exit 1; }

echo "=== 2/4 Fine-tune from v2-base (C3-only, no generic dilution) ==="
python3 scripts/train_kompress.py \
    --data data/c3_train.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v9-finetuned \
    --epochs 5 \
    --batch-size 8 \
    --lr 1e-5

echo "=== 3/4 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v9-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/4 ONNX export + HuggingFace upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v9-finetuned")
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

os.makedirs("kompress-v9-finetuned/onnx", exist_ok=True)
torch.onnx.export(
    Wrapper(model),
    (dummy["input_ids"], dummy["attention_mask"]),
    "kompress-v9-finetuned/onnx/kompress-fp32.onnx",
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
    folder_path="kompress-v9-finetuned",
    repo_id=repo,
    commit_message="kompress-v9: C3-pure — Qwen2.5-7B labels only, no generic dilution"
)
print(f"Uploaded to {repo}")
PY
fi
echo "=== v9 complete ==="
