#!/usr/bin/env bash
# Kompress v6: agent-distribution fine-tune from v4
# Hypothesis: synthetic Claude Code patterns (bash/file/error/search/json)
#             close training-production gap, improve real-world compression
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v6"}

echo "=== 1/6 Generate agent training data ==="
python3 scripts/build_domain_data.py \
    --agent-only \
    --per-domain 600 \
    --output data/kompress_agent_train.jsonl

echo "=== 2/6 Use generator references directly (mk_in_ref=1.0 by construction) ==="
# Self-labeling with v4+override degrades agent data: the subword tokenizer splits
# paths (/var/log/app.log) and CamelCase (TokenExpiredError) so individual subtokens
# don't match _MUST_KEEP_RE, causing v4 to drop them (observed mk_in_ref=0.652).
# The generator references already preserve all must-keep tokens (verified: 0 violations
# across 3000 rows). Using them directly is correct — self-labeling was only needed for
# noisy alpaca Q&A labels in v3/v4.
cp data/kompress_agent_train.jsonl data/agent_self_labeled.jsonl
echo "Agent pairs (generator references, mk_in_ref=1.0): $(wc -l < data/agent_self_labeled.jsonl)"

echo "=== 3/6 Merge: agent self-labeled + existing generic ==="
python3 - << 'PY'
import json
sources = [
    ("data/agent_self_labeled.jsonl", "agent"),
    ("data/kompress_multi_train.jsonl", "generic"),
]
merged = []
for path, label in sources:
    try:
        rows = [json.loads(l) for l in open(path)]
        merged.extend(rows)
        print(f"  {label}: {len(rows)} rows from {path}")
    except FileNotFoundError:
        print(f"  WARN: {path} not found, skipping")
import random; random.seed(42); random.shuffle(merged)
with open("data/kompress_v6_train.jsonl","w") as f:
    for r in merged: f.write(json.dumps(r, ensure_ascii=False)+"\n")
print(f"Total: {len(merged)} rows -> data/kompress_v6_train.jsonl")
PY

echo "=== 4/6 Fine-tune from v4 ==="
python3 scripts/train_kompress_v32.py \
    --data data/kompress_v6_train.jsonl \
    --base-model PeetPedro/kompress-v4 \
    --output kompress-v6-finetuned \
    --epochs 3 \
    --batch-size 16

echo "=== 5/6 Heretic eval ==="
python3 scripts/eval_heretic.py \
    --model kompress-v6-finetuned || echo "WARN: eval non-fatal"

echo "=== 6/6 ONNX export + HuggingFace upload ==="
pip install -q onnx onnxruntime
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v6-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")
class W(torch.nn.Module):
    def __init__(self,m): super().__init__(); self.m=m
    def forward(self,i,a):
        l,s=self.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]
        return p*(0.5+0.5*s)
os.makedirs("kompress-v6-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(model),(dummy["input_ids"],dummy["attention_mask"]),
    "kompress-v6-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"],output_names=["final_scores"],
    dynamic_axes={"input_ids":{0:"b",1:"s"},"attention_mask":{0:"b",1:"s"},"final_scores":{0:"b",1:"s"}},
    opset_version=17)
print("ONNX exported")
PY

if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api=HfApi(token=os.environ["HF_TOKEN"])
repo=os.environ["HF_REPO"]
api.create_repo(repo,exist_ok=True,private=False)
api.upload_folder(folder_path="kompress-v6-finetuned",repo_id=repo,
    commit_message="kompress-v6: agent-distribution fine-tune from v4")
print(f"Uploaded to {repo}")
PY
fi
echo "=== Done. Check heretic exact_pct >= 0.967 ==="
