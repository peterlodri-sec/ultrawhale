#!/usr/bin/env bash
# Kompress v7: sliding-window self-labeling fixes agent-data tokenizer gap
#
# v6 finding: compress_with_override checks individual subtokens.
# TokenExpiredError → Token+Expired+Error, none match CamelCase regex.
# /var/log/app.log → /+var+/+log+..., none match path regex.
# Fix: decode 1/2/3-token windows; if combined string matches, keep all.
# Expected: mk_in_ref recovers from 0.652 to >=0.85 on agent data.
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v7"}

echo "=== 1/6 Generate agent training data ==="
python3 scripts/build_domain_data.py \
    --agent-only \
    --per-domain 600 \
    --output data/kompress_agent_train.jsonl

echo "=== 2/6 Self-label agent data with v6 + sliding-window override ==="
python3 - << 'PY'
import json, re, torch, sys, pathlib
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

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
load_v2_weights(model, "PeetPedro/kompress-v6")
model.eval()
device = "cuda" if torch.cuda.is_available() else "cpu"
model = model.to(device)
print(f"Self-labeling on {device} with sliding-window override")

def compress_with_sliding_window_override(text):
    enc = tok(text, return_tensors="pt", truncation=True, max_length=512, padding=True)
    enc = {k: v.to(device) for k, v in enc.items()}
    with torch.no_grad():
        logits, span = model(enc["input_ids"], enc["attention_mask"])
        probs = torch.softmax(logits, dim=-1)[0, :, 1]
        scores = probs * (0.5 + 0.5 * span[0])
        keep = scores > 0.5
    tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])
    n = len(tokens)
    # Sliding-window: check 1, 2, 3-token windows for must-keep patterns.
    # Fixes: TokenExpiredError (3 subtokens), /var/log/app.log (multi-subtoken),
    #        --verbose (2 subtokens: -- + verbose).
    for i in range(n):
        for width in (1, 2, 3):
            if i + width > n:
                break
            w = tok.convert_tokens_to_string(tokens[i:i+width]).strip()
            if _MUST_KEEP_RE.search(w):
                for j in range(i, i + width):
                    keep[j] = True
                break
    kept = [t for t, k in zip(tokens, keep) if k
            and t not in ("[CLS]","[SEP]","<s>","</s>")]
    return tok.convert_tokens_to_string(kept)

records = [json.loads(l) for l in open("data/kompress_agent_train.jsonl")]
print(f"Self-labeling {len(records)} agent records...")
out, skipped = [], 0
for i, r in enumerate(records):
    if i % 500 == 0: print(f"  {i}/{len(records)}")
    new_ref = compress_with_sliding_window_override(r["text"])
    ratio = len(r["text"]) / max(len(new_ref), 1)
    if ratio >= 1.2 and len(new_ref) >= 30:
        out.append({
            "text": r["text"], "reference": new_ref,
            "role": r.get("role", "tool"),
            "source": "self_labeled_v6_sw_" + r["source"],
            "topic": r.get("topic", ""),
        })
    else:
        skipped += 1

import random; random.seed(42)
samp = random.sample(out, min(100, len(out)))
mk_t, mk_r = 0, 0
for r in samp:
    must = [m.group(0) for m in _MUST_KEEP_RE.finditer(r["text"])]
    mk_t += len(must); mk_r += sum(1 for m in must if m in r["reference"])
mk_ratio = mk_r / max(mk_t, 1)
print(f"mk_in_ref: {mk_ratio:.3f} (target >= 0.85, v6-no-fix was 0.652)")
if mk_ratio < 0.85:
    print(f"WARN: mk_in_ref {mk_ratio:.3f} < 0.85 — sliding-window fix insufficient, using generator refs")
    import shutil
    shutil.copy("data/kompress_agent_train.jsonl", "data/agent_self_labeled.jsonl")
    print("Fell back to generator references (mk_in_ref=1.0)")
else:
    print(f"Sliding-window fix works: mk_in_ref {mk_ratio:.3f}")
    with open("data/agent_self_labeled.jsonl","w") as f:
        for r in out: f.write(json.dumps(r, ensure_ascii=False)+"\n")
    print(f"Written {len(out)} self-labeled pairs (skipped {skipped})")
PY

count=$(wc -l < data/agent_self_labeled.jsonl)
echo "Agent pairs: $count"
[ "$count" -gt 0 ] || { echo "ERROR: no agent pairs"; exit 1; }

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
with open("data/kompress_v7_train.jsonl","w") as f:
    for r in merged: f.write(json.dumps(r, ensure_ascii=False)+"\n")
print(f"Total: {len(merged)} rows -> data/kompress_v7_train.jsonl")
PY

echo "=== 4/6 Fine-tune from v6 ==="
python3 scripts/train_kompress_v32.py \
    --data data/kompress_v7_train.jsonl \
    --base-model PeetPedro/kompress-v6 \
    --output kompress-v7-finetuned \
    --epochs 3 \
    --batch-size 16

echo "=== 5/6 Heretic eval ==="
python3 scripts/eval_heretic.py \
    --model kompress-v7-finetuned || echo "WARN: eval non-fatal"

echo "=== 6/6 ONNX export + HuggingFace upload ==="
pip install -q onnx onnxruntime
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v7-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")
class W(torch.nn.Module):
    def __init__(self,m): super().__init__(); self.m=m
    def forward(self,i,a):
        l,s=self.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]
        return p*(0.5+0.5*s)
os.makedirs("kompress-v7-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(model),(dummy["input_ids"],dummy["attention_mask"]),
    "kompress-v7-finetuned/onnx/kompress-fp32.onnx",
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
api.upload_folder(folder_path="kompress-v7-finetuned",repo_id=repo,
    commit_message="kompress-v7: sliding-window self-labeling fix, fine-tuned from v6")
print(f"Uploaded to {repo}")
PY
fi
echo "=== Done. Check heretic exact_pct vs v6 (0.962) and mk_in_ref vs target 0.85 ==="
