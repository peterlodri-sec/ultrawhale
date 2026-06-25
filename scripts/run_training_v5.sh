#!/usr/bin/env bash
# Kompress v4: self-labeled references via v3+override, then fine-tune
# Hypothesis: mk_in_ref=1.0 labels → exact_pct escapes 0.882 ceiling?
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v5"}

echo "=== 1/5 Self-label references using v3 + override ==="
python3 - << 'PY'
import json, re, torch, sys
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
from huggingface_hub import hf_hub_download
import pathlib

_MUST_KEEP_RE = re.compile(
    r"\b0x[0-9A-Fa-f]+\b"
    r"|(?<![\w.])\d+(?:\.\d+)?(?![\w.])"
    r"|[A-Z_]{2,}"
    r"|[a-z_][a-z0-9_]*\.[a-z0-9_]+"
    r"|/[a-z0-9/._-]{2,}"
    r"|\.[a-z]{2,4}\b"
    r"|--?[a-z][\w-]*"
    r"|\b[A-Z][a-z]+[A-Z]\w*"
)

BASE = "answerdotai/ModernBERT-base"
tok = AutoTokenizer.from_pretrained(BASE)
model = HeadroomCompressorModel(BASE)
load_v2_weights(model, "PeetPedro/kompress-v4")
model.eval()
device = "cuda" if torch.cuda.is_available() else "cpu"
model = model.to(device)
print(f"Device: {device}")

def compress_with_override(text):
    enc = tok(text, return_tensors="pt", truncation=True, max_length=512,
               padding=True)
    enc = {k: v.to(device) for k, v in enc.items()}
    with torch.no_grad():
        logits, span = model(enc["input_ids"], enc["attention_mask"])
        probs = torch.softmax(logits, dim=-1)[0, :, 1]
        scores = probs * (0.5 + 0.5 * span[0])
        keep = scores > 0.5
    tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])
    for i, t in enumerate(tokens):
        w = tok.convert_tokens_to_string([t]).strip()
        if _MUST_KEEP_RE.search(w):
            keep[i] = True
    kept = [t for t, k in zip(tokens, keep) if k
            and t not in ("[CLS]","[SEP]","<s>","</s>")]
    return tok.convert_tokens_to_string(kept)

records = [json.loads(l) for l in open("data/kompress_train_split.jsonl")]
print(f"Self-labeling {len(records)} records on {device}...")
out, skipped = [], 0
for i, r in enumerate(records):
    if i % 200 == 0: print(f"  {i}/{len(records)}")
    text = r["text"]
    new_ref = compress_with_override(text)
    ratio = len(text) / max(len(new_ref), 1)
    if ratio >= 1.2 and len(new_ref) >= 30:
        out.append({
            "text": text, "reference": new_ref,
            "role": r.get("role","assistant"),
            "source": "self_labeled_v4_override",
            "topic": r.get("topic",""),
        })
    else:
        skipped += 1

# mk_in_ref check
import random; random.seed(42)
samp = random.sample(out, min(100, len(out)))
mk_t, mk_r = 0, 0
for r in samp:
    must = [m.group(0) for m in _MUST_KEEP_RE.finditer(r["text"])]
    mk_t += len(must); mk_r += sum(1 for m in must if m in r["reference"])
print(f"mk_in_ref: {mk_r/max(mk_t,1):.3f} over sample (target ~1.0)")
print(f"Written {len(out)} pairs (skipped {skipped})")

pathlib.Path("data").mkdir(exist_ok=True)
with open("data/self_labeled_train.jsonl","w") as f:
    for r in out: f.write(json.dumps(r, ensure_ascii=False)+"\n")
PY

echo "=== 2/5 Merge with domain data ==="
python3 - << 'PY'
import json, pathlib
splits = ["data/self_labeled_train.jsonl", "data/domain_train.jsonl"]
merged = []
for s in splits:
    try:
        rows = [json.loads(l) for l in open(s)]
        merged.extend(rows)
        print(f"  {s}: {len(rows)} rows")
    except FileNotFoundError:
        print(f"  WARN: {s} not found, skipping")
with open("data/v5_train.jsonl","w") as f:
    for r in merged: f.write(json.dumps(r, ensure_ascii=False)+"\n")
print(f"Merged {len(merged)} total -> data/v5_train.jsonl")
PY

echo "=== 3/5 Fine-tune v4 ==="
python3 scripts/train_kompress_v32.py \
    --data data/v5_train.jsonl \
    --base-model PeetPedro/kompress-v4 \
    --output kompress-v5-finetuned \
    --epochs 3 \
    --batch-size 16

echo "=== 4/5 Eval ==="
python3 scripts/eval_kompress.py \
    --model kompress-v5-finetuned \
    --data data/kompress_test.jsonl || echo "WARN: eval non-fatal"

echo "=== 5/5 ONNX + upload ==="
pip install -q onnx onnxruntime
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v5-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")
class W(torch.nn.Module):
    def __init__(self,m): super().__init__(); self.m=m
    def forward(self,i,a):
        l,s=self.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]
        return p*(0.5+0.5*s)
os.makedirs("kompress-v5-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(model),(dummy["input_ids"],dummy["attention_mask"]),
    "kompress-v5-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"],output_names=["final_scores"],
    dynamic_axes={"input_ids":{0:"b",1:"s"},"attention_mask":{0:"b",1:"s"},"final_scores":{0:"b",1:"s"}},
    opset_version=17)
print("ONNX exported")
PY
if [ -n "$HF_TOKEN" ]; then
    python3 - << PY
from huggingface_hub import HfApi; import os
api=HfApi(token=os.environ["HF_TOKEN"])
api.create_repo("${HF_REPO}",exist_ok=True,private=False)
api.upload_folder(folder_path="kompress-v5-finetuned",repo_id="${HF_REPO}",
    commit_message="kompress-v5: self-labeled references via v3+override")
print("Uploaded to ${HF_REPO}")
PY
fi
echo "=== Done ==="
