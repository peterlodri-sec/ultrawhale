#!/usr/bin/env bash
# Kompress v10: scaled C3 — 1200 Qwen labels + 600 generic (2:1)
#
# v8: 97 C3 + 200 generic → heretic 0.955
# v9: 97 C3 only → heretic 0.921 (overfit)
#
# v10 hypothesis: more Qwen teacher labels + generic diversity at 2:1 ratio
# should push heretic closer to v2-base's 0.975 while keeping mk_in_ref high.
#
# Data: 1200 Qwen-labeled + 600 generic (sampled from kompress_multi_train)
# Base:  chopratejas/kompress-v2-base
# Target: heretic >= 0.965
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v10"}

echo "=== 1/5 Convert Qwen labels to training format ==="
python3 - << 'PY'
import json, random

# Convert Qwen spans to text+reference format
c3 = [json.loads(l) for l in open("data/c3_qwen_labeled.jsonl")]
print(f"  C3 Qwen-labeled: {len(c3)}")

records = []
for r in c3:
    spans = r.get("spans", [])
    # Reconstruct reference from kept spans
    text = r["text"]
    kept_parts = []
    for s in sorted(spans, key=lambda x: x["start"]):
        kept_parts.append(text[s["start"]:s["end"]])
    ref = " ".join(kept_parts)
    
    if len(ref) >= 20:  # filter empty references
        records.append({
            "text": text,
            "reference": ref,
            "role": "tool",
            "source": f"qwen_c3_{r.get('domain','?')}",
            "topic": "compression",
        })

print(f"  Converted: {len(records)} pairs (filtered {len(c3)-len(records)} empty-ref)")

# Load generic and sample to get 2:1 ratio
generic = []
with open("data/kompress_multi_train.jsonl") as f:
    for line in f:
        generic.append(json.loads(line))

random.seed(42)
generic_sample = random.sample(generic, min(len(records)//2, len(generic)))
print(f"  Generic: {len(generic_sample)} (from {len(generic)} available)")

merged = records + generic_sample
random.shuffle(merged)
print(f"  Total: {len(merged)} pairs ({len(records)} C3 + {len(generic_sample)} generic)")

with open("data/kompress_v10_train.jsonl", "w") as f:
    for r in merged:
        f.write(json.dumps(r, ensure_ascii=False) + "\n")
PY

count=$(wc -l < data/kompress_v10_train.jsonl)
echo "Training pairs: $count"
[ "$count" -gt 100 ] || { echo "ERROR: too few pairs ($count)"; exit 1; }

echo "=== 2/5 Fine-tune from v2-base ==="
python3 scripts/train_kompress.py \
    --data data/kompress_v10_train.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v10-finetuned \
    --epochs 3 \
    --batch-size 16 \
    --lr 2e-5

echo "=== 3/5 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v10-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/5 ONNX export ==="
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
torch.onnx.export(
    Wrapper(model),
    (dummy["input_ids"], dummy["attention_mask"]),
    "kompress-v10-finetuned/onnx/kompress-fp32.onnx",
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

echo "=== 5/5 HuggingFace upload ==="
if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api = HfApi(token=os.environ["HF_TOKEN"])
repo = os.environ["HF_REPO"]
api.create_repo(repo, exist_ok=True, private=False)
api.upload_folder(
    folder_path="kompress-v10-finetuned",
    repo_id=repo,
    commit_message="kompress-v10: scaled C3 — 1200 Qwen2.5-7B labels + 600 generic (2:1)"
)
print(f"Uploaded to {repo}")
PY
fi
echo "=== v10 complete ==="
