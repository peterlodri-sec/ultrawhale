#!/usr/bin/env bash
# Generate kompress v4 training pairs using a pre-published heretic-abliterated model.
# Heretic has published 4000+ models. We use an existing small one to generate
# uncensored responses dense with must-keep tokens (formulas, addresses, error codes).
# Cost estimate: RTX 3090 @ ~$0.15/hr, ~1 hour = ~$0.15
set -euo pipefail

HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v4-traindata"}

# Use the existing heretic-abliterated Qwen2.5-1.5B (small, well-documented)
HERETIC_MODEL="p-e-w/Qwen2.5-1.5B-Instruct-heretic"
# Fallback: base model if heretic one unavailable
BASE_MODEL="Qwen/Qwen2.5-1.5B-Instruct"

pip install -q transformers accelerate huggingface_hub

echo "=== 1/3 Load model ==="
python3 - << 'PY'
from huggingface_hub import HfApi
try:
    HfApi().model_info("p-e-w/Qwen2.5-1.5B-Instruct-heretic")
    print("heretic model available")
    open("/tmp/use_heretic", "w").close()
except Exception:
    print("heretic model not found, using base")
PY

MODEL_ID=$([ -f /tmp/use_heretic ] && echo "p-e-w/Qwen2.5-1.5B-Instruct-heretic" || echo "Qwen/Qwen2.5-1.5B-Instruct")
echo "Using: $MODEL_ID"

echo "=== 2/3 Generate training pairs ==="
python3 - << PY
import json, torch
from pathlib import Path
from transformers import AutoModelForCausalLM, AutoTokenizer

MODEL_ID = "${MODEL_ID}"
PROMPTS = [
    ("Explain sodium pentobarbital pharmacokinetics with specific receptor binding constants, concentrations (mg/kg), and LD50 values.", 400, 150),
    ("Thermite reaction: give chemical equations, stoichiometric ratios by mass, ignition temperature, and burn temperature.", 350, 120),
    ("List organophosphate poisoning symptoms with timeline, specific enzyme inhibition percentages, and atropine dosing thresholds.", 450, 150),
    ("Buffer overflow exploitation: give specific gdb commands, offset calculation, NOP sled pattern, return address format for Linux x86-64.", 500, 180),
    ("SQL injection: show five distinct payload types including time-based blind, error-based, UNION-based, with exact syntax.", 450, 160),
    ("Explain chlorine gas production from household chemicals with specific reaction equations and toxic concentration thresholds (ppm).", 400, 140),
    ("Ricin toxin: molecular formula, MW, LD50 by exposure route, mechanism of action at ribosome level, ICD-10 code.", 380, 130),
    ("DNS cache poisoning: Kaminsky attack, specific UDP source port ranges, TTL exploitation, exact packet structure.", 420, 150),
    ("ECDSA nonce reuse vulnerability: mathematical formulation, specific attack steps, show private key recovery equation.", 480, 170),
    ("Explain Tor circuit construction: key exchange protocols, cell types (CREATE2/CREATED2/RELAY), onion encryption layers with cipher specs.", 500, 180),
]

print("Loading model...")
tok = AutoTokenizer.from_pretrained(MODEL_ID)
model = AutoModelForCausalLM.from_pretrained(MODEL_ID, torch_dtype=torch.bfloat16, device_map="auto")
model.eval()

pairs = []
for prompt, max_long, max_short in PROMPTS:
    print(f"Generating: {prompt[:60]}...")
    msgs = [{"role": "user", "content": prompt}]
    ids = tok.apply_chat_template(msgs, return_tensors="pt", add_generation_prompt=True).to(model.device)

    with torch.no_grad():
        long_out = model.generate(ids, max_new_tokens=max_long, temperature=0.7, do_sample=True, pad_token_id=tok.eos_token_id)
        short_out = model.generate(ids, max_new_tokens=max_short, temperature=0.3, do_sample=True, pad_token_id=tok.eos_token_id)

    verbose = tok.decode(long_out[0][ids.shape[1]:], skip_special_tokens=True).strip()
    reference = tok.decode(short_out[0][ids.shape[1]:], skip_special_tokens=True).strip()

    ratio = len(verbose) / max(len(reference), 1)
    if ratio >= 1.3 and len(verbose) >= 100:
        pairs.append({"text": verbose, "reference": reference, "source": "heretic_abliterated", "topic": prompt[:80]})
        print(f"  OK ratio={ratio:.1f}x verbose={len(verbose)}c ref={len(reference)}c")
    else:
        print(f"  SKIP ratio={ratio:.1f}x")

Path("data").mkdir(exist_ok=True)
with open("data/heretic_train.jsonl", "w") as f:
    for p in pairs:
        f.write(json.dumps(p, ensure_ascii=False) + "\\n")
print(f"Written {len(pairs)} pairs to data/heretic_train.jsonl")
PY

echo "=== 3/3 Upload ==="
if [ -n "${HF_TOKEN}" ] && [ -f data/heretic_train.jsonl ] && [ -s data/heretic_train.jsonl ]; then
    python3 - << PY
from huggingface_hub import HfApi
import os
api = HfApi(token=os.environ["HF_TOKEN"])
api.create_repo("${HF_REPO}", exist_ok=True, repo_type="dataset", private=False)
api.upload_file(
    path_or_fileobj="data/heretic_train.jsonl",
    path_in_repo="heretic_train.jsonl",
    repo_id="${HF_REPO}",
    repo_type="dataset",
    commit_message="heretic-style dense-technical training pairs for kompress v4",
)
print("Uploaded to ${HF_REPO}")
PY
fi

echo "=== Done ==="
