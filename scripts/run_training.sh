#!/usr/bin/env bash
# Full Kompress v3 training run — executes on vast.ai instance
# Expected runtime: ~30 min on RTX 3090
set -euo pipefail

cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"peterlodri-sec/kompress-v3"}

echo "=== 1/4 Training data ==="
# Use pre-exported data committed to the repo (avoids HF download issues on some instances)
if [ -f "data/kompress_train_split.jsonl" ] && [ -f "data/kompress_test.jsonl" ]; then
    echo "Using committed data files"
    python3 -c "
import json
train = sum(1 for _ in open('data/kompress_train_split.jsonl'))
test  = sum(1 for _ in open('data/kompress_test.jsonl'))
print(f'Train: {train}, Test: {test}')
"
else
    echo "Falling back to HF download..."
    python3 scripts/export_for_kompress.py --output data/kompress_train.jsonl
    python3 - << 'PY'
import json, random, pathlib
records = [json.loads(l) for l in open("data/kompress_train.jsonl")]
random.shuffle(records)
split = int(len(records) * 0.9)
pathlib.Path("data").mkdir(exist_ok=True)
with open("data/kompress_train_split.jsonl","w") as f:
    for r in records[:split]: f.write(json.dumps(r)+"\n")
with open("data/kompress_test.jsonl","w") as f:
    for r in records[split:]: f.write(json.dumps(r)+"\n")
print(f"Train: {split}, Test: {len(records)-split}")
PY
fi

echo "=== 2/4 Fine-tune ==="
python3 scripts/train_kompress.py \
    --data data/kompress_train_split.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v3-finetuned \
    --epochs 3 \
    --batch-size 16

echo "=== 3/4 Eval ==="
python3 scripts/eval_kompress.py \
    --model kompress-v3-finetuned \
    --data data/kompress_test.jsonl

echo "=== 4/4 Export ONNX and upload to HF ==="
# Re-use headroom's export script pointed at our checkpoint
pip install -q onnx onnxruntime

python3 - << 'PY'
import sys, os
sys.path.insert(0, "/workspace/ultrawhale")

# Minimal ONNX export matching headroom's expected format
import torch
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v3-finetuned")
model.eval()

tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")

# Dummy input for tracing
dummy = tok("hello world", return_tensors="pt")
ids = dummy["input_ids"]
mask = dummy["attention_mask"]

class _Wrapper(torch.nn.Module):
    def __init__(self, m): super().__init__(); self.m = m
    def forward(self, input_ids, attention_mask):
        logits, span = self.m(input_ids, attention_mask)
        token_probs = torch.softmax(logits, dim=-1)[:, :, 1]
        span_scores = span
        final_scores = token_probs * (0.5 + 0.5 * span_scores)
        return final_scores

wrapped = _Wrapper(model)

import os; os.makedirs("kompress-v3-finetuned/onnx", exist_ok=True)
torch.onnx.export(
    wrapped, (ids, mask),
    "kompress-v3-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids", "attention_mask"],
    output_names=["final_scores"],
    dynamic_axes={
        "input_ids": {0: "batch", 1: "seq"},
        "attention_mask": {0: "batch", 1: "seq"},
        "final_scores": {0: "batch", 1: "seq"},
    },
    opset_version=17,
)
print("ONNX exported to kompress-v3-finetuned/onnx/kompress-fp32.onnx")
PY

if [ -n "$HF_TOKEN" ]; then
    echo "Uploading to HF: $HF_REPO"
    python3 - << PY
from huggingface_hub import HfApi
api = HfApi()
api.create_repo("${HF_REPO}", exist_ok=True, private=False)
api.upload_folder(
    folder_path="kompress-v3-finetuned",
    repo_id="${HF_REPO}",
    commit_message="kompress-v3: LoRA fine-tuned on ultrawhale + headroom-aware labels",
)
print("Uploaded to huggingface.co/datasets/${HF_REPO}")
PY
else
    echo "HF_TOKEN not set, skipping upload. Model in: kompress-v3-finetuned/"
fi

echo "=== Done. GPU time used: check vast.ai billing ==="
