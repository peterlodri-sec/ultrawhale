#!/usr/bin/env bash
# Kompress v9: C3-pure — Qwen labels only, no generic data dilution
#
# v8: heretic 0.955 with C3+generic. Hypothesis: generic data (dolly/openorca)
# drags precision down from v2-base's 0.975. v9 trains on C3 Qwen-labeled
# data ONLY to test this.
#
# Data: 97 C3 Qwen-labeled pairs in text+reference format
# Base:  chopratejas/kompress-v2-base
# Target: heretic >= 0.965, override_delta = 0
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v9"}

echo "=== 1/4 C3-pure training data (97 Qwen-labeled pairs) ==="
wc -l data/kompress_v9_train.jsonl

echo "=== 2/4 Fine-tune from v2-base (C3-only) ==="
python3 scripts/train_kompress.py \
    --data data/kompress_v9_train.jsonl \
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
    commit_message="kompress-v9: C3-pure — Qwen2.5-7B labels only, 97 pairs, no generic dilution"
)
print(f"Uploaded to {repo}")
PY
fi
echo "=== v9 complete ==="
