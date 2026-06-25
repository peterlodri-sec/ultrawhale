#!/usr/bin/env bash
# Kompress v12: Qwen3-Coder-Next teacher labels (stronger than Qwen2.5-7B)
#
# Hypothesis: Qwen3-Coder-Next is purpose-built for coding agents — it should
# identify must-keep tokens in code-heavy tool outputs more accurately than
# the general-purpose Qwen2.5-7B. Better labels → higher heretic precision.
#
# Teacher: Qwen/Qwen3-Coder-Next (latest coding agent model)
# Data: ~300 Qwen3-Coder labeled + 600 generic (33% C3 ratio, v8 sweet spot)
# Encoder: ModernBERT-base (149M) — we know large encoder doesn't help
# Base: chopratejas/kompress-v2-base
# Target: heretic >= 0.965
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v12"}

echo "=== 1/5 Convert Qwen3-Coder labels to training format ==="
python3 - << 'PY'
import json, random

c3 = [json.loads(l) for l in open("data/c3_qwen3coder_labeled.jsonl")]
good = [r for r in c3 if len(r.get("spans", [])) > 0]
print(f"  Qwen3-Coder labeled: {len(c3)}, with-spans: {len(good)}")

records = []
for r in good:
    text = r["text"]
    spans = sorted(r["spans"], key=lambda x: x["start"])
    parts = [text[s["start"]:s["end"]] for s in spans]
    ref = " ".join(parts)
    if len(ref) >= 20:
        records.append({
            "text": text, "reference": ref, "role": "tool",
            "source": f"qwen3coder_c3_{r.get('domain','?')}", "topic": "compression",
        })

print(f"  Converted: {len(records)} pairs")

# 33% C3 ratio (v8 sweet spot): need 2x generic
generic = [json.loads(l) for l in open("data/kompress_multi_train.jsonl")]
random.seed(42)
generic_sample = random.sample(generic, min(len(records)*2, len(generic)))
print(f"  Generic: {len(generic_sample)}")

merged = records + generic_sample
random.shuffle(merged)
with open("data/kompress_v12_train.jsonl", "w") as f:
    for r in merged:
        f.write(json.dumps(r, ensure_ascii=False) + "\n")
print(f"  Total: {len(merged)}")
PY

echo "=== 2/5 Fine-tune from v2-base ==="
python3 scripts/train_kompress.py \
    --data data/kompress_v12_train.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v12-finetuned \
    --epochs 3 \
    --batch-size 16 \
    --lr 2e-5

echo "=== 3/5 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v12-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/5 ONNX export ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v12-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")
class W(torch.nn.Module):
    def __init__(self,m): super().__init__(); self.m=m
    def forward(self,i,a): l,s=self.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]; return p*(0.5+0.5*s)
os.makedirs("kompress-v12-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(model),(dummy["input_ids"],dummy["attention_mask"]),
    "kompress-v12-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"],output_names=["final_scores"],
    dynamic_axes={"input_ids":{0:"b",1:"s"},"attention_mask":{0:"b",1:"s"},"final_scores":{0:"b",1:"s"}},
    opset_version=17)
print("ONNX exported")
PY

echo "=== 5/5 HuggingFace upload ==="
if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api = HfApi(token=os.environ["HF_TOKEN"])
repo = os.environ["HF_REPO"]
api.create_repo(repo, exist_ok=True, private=False)
api.upload_folder(folder_path="kompress-v12-finetuned", repo_id=repo,
    commit_message="kompress-v12: Qwen3-Coder-Next teacher labels (stronger than Qwen2.5-7B)")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v12 complete ==="
