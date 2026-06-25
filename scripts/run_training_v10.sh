#!/usr/bin/env bash
# Kompress v10: scaled C3 — 285 Qwen labels + 600 generic (1:2 ratio)
# Pre-built training data in kompress_v10_train.jsonl (text+reference format)
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v10"}

echo "=== 1/4 Training data: $(wc -l < data/kompress_v10_train.jsonl) pairs ==="

echo "=== 2/4 Fine-tune from v2-base ==="
python3 scripts/train_kompress.py \
    --data data/kompress_v10_train.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v10-finetuned \
    --epochs 3 \
    --batch-size 16 \
    --lr 2e-5

echo "=== 3/4 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v10-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/4 ONNX export + upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v10-finetuned")
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

os.makedirs("kompress-v10-finetuned/onnx", exist_ok=True)
torch.onnx.export(Wrapper(model), (dummy["input_ids"], dummy["attention_mask"]),
    "kompress-v10-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"], output_names=["final_scores"],
    dynamic_axes={"input_ids":{0:"b",1:"s"},"attention_mask":{0:"b",1:"s"},"final_scores":{0:"b",1:"s"}},
    opset_version=17)
print("ONNX exported")
PY

if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api = HfApi(token=os.environ["HF_TOKEN"])
repo = os.environ["HF_REPO"]
api.create_repo(repo, exist_ok=True, private=False)
api.upload_folder(folder_path="kompress-v10-finetuned", repo_id=repo,
    commit_message="kompress-v10: scaled C3 — 285 Qwen2.5-7B labels + 600 generic")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v10 complete ==="
