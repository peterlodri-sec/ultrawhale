#!/usr/bin/env bash
# Run on the vast.ai instance to finish eval + export + upload.
# Usage: vastai ssh <instance_id> < scripts/vast_finish.sh
set -euo pipefail
cd /workspace/ultrawhale
git pull --ff-only

echo "=== 3/4 Eval ==="
python3 scripts/eval_kompress.py \
    --model kompress-v3-finetuned \
    --data data/kompress_test.jsonl || echo "WARN: eval non-fatal, continuing"

echo "=== 4/4 ONNX export ==="
pip install -q onnx onnxruntime
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v3-finetuned")
model.eval()

tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")

class Wrapper(torch.nn.Module):
    def __init__(self, m): super().__init__(); self.m = m
    def forward(self, input_ids, attention_mask):
        logits, span = self.m(input_ids, attention_mask)
        prob = torch.softmax(logits, dim=-1)[:, :, 1]
        return prob * (0.5 + 0.5 * span)

os.makedirs("kompress-v3-finetuned/onnx", exist_ok=True)
torch.onnx.export(
    Wrapper(model),
    (dummy["input_ids"], dummy["attention_mask"]),
    "kompress-v3-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids", "attention_mask"],
    output_names=["final_scores"],
    dynamic_axes={"input_ids": {0:"batch",1:"seq"}, "attention_mask": {0:"batch",1:"seq"}, "final_scores": {0:"batch",1:"seq"}},
    opset_version=17,
)
print("ONNX exported")
PY

echo "=== Upload to HuggingFace ==="
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"peterlodri-sec/kompress-v3"}
if [ -n "$HF_TOKEN" ]; then
    python3 - << PY
from huggingface_hub import HfApi
api = HfApi()
api.create_repo("${HF_REPO}", exist_ok=True, private=False)
api.upload_folder(
    folder_path="kompress-v3-finetuned",
    repo_id="${HF_REPO}",
    commit_message="kompress-v3: fine-tuned on ultrawhale, loss=0.0812",
)
print("Uploaded to huggingface.co/${HF_REPO}")
PY
else
    echo "HF_TOKEN not set — model at kompress-v3-finetuned/"
fi

echo "=== Done ==="
