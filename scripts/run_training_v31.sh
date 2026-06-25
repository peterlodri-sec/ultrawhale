#!/usr/bin/env bash
# Full Kompress v3.1 training run — executes on vast.ai instance
# Changes from run_training.sh:
#   - Merges kompress_train_split + kompress_multi_train + domain_train into v31_train
#   - Uses train_kompress_v31.py (must_keep_weight=6.0, hard label override)
#   - Fine-tunes from PeetPedro/kompress-v3 (not v2-base)
#   - Uploads to PeetPedro/kompress-v31
set -euo pipefail

cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v31"}

echo "=== 1/5 Build domain training data ==="
python3 scripts/build_domain_data.py --output data/domain_train.jsonl
python3 -c "
import json
n = sum(1 for _ in open('data/domain_train.jsonl'))
print(f'Domain pairs: {n}')
"

echo "=== 2/5 Merge training splits ==="
python3 - << 'PY'
import json, pathlib

files = [
    "data/kompress_train_split.jsonl",
    "data/kompress_multi_train.jsonl",
    "data/domain_train.jsonl",
]

pathlib.Path("data").mkdir(exist_ok=True)
out_path = "data/v31_train.jsonl"
total = 0
with open(out_path, "w") as out:
    for path in files:
        p = pathlib.Path(path)
        if not p.exists():
            print(f"WARN: {path} not found, skipping")
            continue
        count = 0
        for line in open(path):
            line = line.strip()
            if line:
                out.write(line + "\n")
                count += 1
        print(f"  {path}: {count} rows")
        total += count

print(f"Merged {total} total rows -> {out_path}")
PY

echo "=== 3/5 Fine-tune v3.1 ==="
python3 scripts/train_kompress_v31.py \
    --data data/v31_train.jsonl \
    --base-model PeetPedro/kompress-v3 \
    --output kompress-v31-finetuned \
    --epochs 3 \
    --batch-size 16

echo "=== 4/5 Eval ==="
python3 scripts/eval_kompress.py \
    --model kompress-v31-finetuned \
    --data data/kompress_test.jsonl || echo "WARN: eval failed, continuing to export"

echo "=== 5/5 Export ONNX and upload to HF ==="
pip install -q onnx onnxruntime

python3 - << 'PY'
import sys, os
sys.path.insert(0, "/workspace/ultrawhale")

import torch
from scripts.train_kompress_v31 import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v31-finetuned")
model.eval()

tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")

dummy = tok("hello world", return_tensors="pt")
ids = dummy["input_ids"]
mask = dummy["attention_mask"]

class _Wrapper(torch.nn.Module):
    def __init__(self, m): super().__init__(); self.m = m
    def forward(self, input_ids, attention_mask):
        logits, span = self.m(input_ids, attention_mask)
        token_probs = torch.softmax(logits, dim=-1)[:, :, 1]
        final_scores = token_probs * (0.5 + 0.5 * span)
        return final_scores

wrapped = _Wrapper(model)

os.makedirs("kompress-v31-finetuned/onnx", exist_ok=True)
torch.onnx.export(
    wrapped, (ids, mask),
    "kompress-v31-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids", "attention_mask"],
    output_names=["final_scores"],
    dynamic_axes={
        "input_ids": {0: "batch", 1: "seq"},
        "attention_mask": {0: "batch", 1: "seq"},
        "final_scores": {0: "batch", 1: "seq"},
    },
    opset_version=17,
)
print("ONNX exported to kompress-v31-finetuned/onnx/kompress-fp32.onnx")
PY

if [ -n "$HF_TOKEN" ]; then
    echo "Uploading to HF: $HF_REPO"
    python3 - << PY
from huggingface_hub import HfApi
api = HfApi()
api.create_repo("${HF_REPO}", exist_ok=True, private=False)
api.upload_folder(
    folder_path="kompress-v31-finetuned",
    repo_id="${HF_REPO}",
    commit_message="kompress-v31: must_keep_weight=6.0, hard label override, domain data",
)
print("Uploaded to huggingface.co/${HF_REPO}")
PY
else
    echo "HF_TOKEN not set, skipping upload. Model in: kompress-v31-finetuned/"
fi

echo "=== Done. GPU time used: check vast.ai billing ==="
