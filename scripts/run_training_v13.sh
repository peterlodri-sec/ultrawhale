#!/usr/bin/env bash
# Kompress v13: GLM-generated agent scenarios + regex-as-teacher labels
#
# Data: GLM-5.1 simulates realistic multi-turn coding sessions across
# Rust/Python/TypeScript/Go. Labels: _MUST_KEEP_RE — deterministic,
# perfect must-keep identification (the same regex in production).
#
# Hypothesis: realistic data distribution + perfectly consistent labels
# should beat both Qwen teachers (biased) and synthetic data (unrealistic).
#
# Encoder: ModernBERT-base (149M) — proven sweet spot
# Base: chopratejas/kompress-v2-base
# Target: heretic >= 0.965
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v13"}

echo "=== 1/4 Convert GLM scenarios to training format ==="
python3 - << 'PY'
import json, re, random

_MUST_KEEP_RE = re.compile(
    r"\b0x[0-9A-Fa-f]+\b"
    r"|(?<![\w.])\d+(?:\.\d+)?(?![\w.])"
    r"|[A-Z_]{2,}"
    r"|[a-z_][a-z0-9_]*\.[a-z0-9_]+"
    r"|/[a-z0-9/._-]{2,}"
    r"|\.[a-z]{2,4}\b"
    r"|--?[a-zA-Z][\w-]*"
    r"|\b[A-Z][a-z]+[A-Z]\w*"
)

rows = [json.loads(l) for l in open("data/glm_scenarios.jsonl")]
print(f"  GLM turns: {len(rows)}")

# Use regex as the labeling "teacher" — extract must-keep tokens from output
records = []
for r in rows:
    text = r["output"]
    matches = [m.group() for m in _MUST_KEEP_RE.finditer(text)]
    if len(matches) >= 3 and len(text) >= 100:
        ref = " ".join(matches)
        records.append({
            "text": text, "reference": ref, "role": "tool",
            "source": f"glm_regex_{r.get('language','?')}", "topic": "compression",
        })

print(f"  Regex-labeled: {len(records)} pairs")

# Mix with generic for diversity (33% ratio)
generic = [json.loads(l) for l in open("data/kompress_multi_train.jsonl")]
random.seed(42)
generic_sample = random.sample(generic, min(len(records)*2, len(generic)))

merged = records + generic_sample
random.shuffle(merged)
with open("data/kompress_v13_train.jsonl", "w") as f:
    for r in merged:
        f.write(json.dumps(r, ensure_ascii=False) + "\n")
print(f"  Total: {len(merged)} ({len(records)} GLM + {len(generic_sample)} generic, {len(records)/len(merged)*100:.0f}% GLM)")
PY

echo "=== 2/4 Fine-tune from v2-base ==="
python3 scripts/train_kompress.py \
    --data data/kompress_v13_train.jsonl \
    --base-model chopratejas/kompress-v2-base \
    --output kompress-v13-finetuned \
    --epochs 3 \
    --batch-size 16 \
    --lr 2e-5

echo "=== 3/4 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v13-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/4 ONNX + upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v13-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")
class W(torch.nn.Module):
    def __init__(self,m): super().__init__(); self.m=m
    def forward(self,i,a): l,s=self.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]; return p*(0.5+0.5*s)
os.makedirs("kompress-v13-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(model),(dummy["input_ids"],dummy["attention_mask"]),
    "kompress-v13-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"],output_names=["final_scores"],
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
api.upload_folder(folder_path="kompress-v13-finetuned", repo_id=repo,
    commit_message="kompress-v13: GLM agent scenarios + regex-as-teacher labels")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v13 complete ==="
