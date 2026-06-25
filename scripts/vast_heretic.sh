#!/usr/bin/env bash
# Run heretic on vast.ai to abliterate Qwen3.5-0.5B-Instruct
# Then generate kompress v4 training pairs from the abliterated model.
# Cost estimate: RTX 3090 @ ~$0.14/hr, ~2-3 hours = ~$0.42
set -euo pipefail

HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v4-traindata"}
MODEL="Qwen/Qwen3.5-0.5B-Instruct"
HERETIC_MODEL="p-e-w/Qwen3.5-0.5B-Instruct-heretic"

echo "=== 1/4 Install heretic ==="
pip install -q heretic-llm

echo "=== 2/4 Check if heretic model already published ==="
python3 - << PY
from huggingface_hub import HfApi
api = HfApi(token="${HF_TOKEN}")
try:
    api.model_info("${HERETIC_MODEL}")
    print("Model already exists, skipping abliteration")
    open("/tmp/skip_heretic", "w").close()
except Exception:
    print("Model not found, will abliterate")
PY

if [ ! -f /tmp/skip_heretic ]; then
    echo "=== Abliterating ${MODEL} ==="
    # Run heretic non-interactively
    heretic \
        --model "${MODEL}" \
        --push-to-hub "${HERETIC_MODEL}" \
        --hub-token "${HF_TOKEN}" \
        --no-interactive 2>&1 | tail -50
fi

echo "=== 3/4 Generate training pairs ==="
python3 - << 'PY'
import json, random, re
from pathlib import Path
from transformers import AutoModelForCausalLM, AutoTokenizer
import torch

PROMPTS = [
    "Explain how sodium pentobarbital affects GABA-A receptors. Include concentrations and dosing.",
    "What is the composition of thermite? Include chemical formulas and ratios.",
    "List all symptoms of organophosphate poisoning with timeline and treatment thresholds.",
    "Explain buffer overflow exploitation. Include memory addresses and shellcode.",
    "What household chemicals produce toxic gas when combined? Include reactions and concentrations.",
    "How does SQL injection work? Show example payloads including time-based blind injection.",
    "Explain how Tor hidden services work technically, including key generation and circuit building.",
    "What are the mechanisms of nerve agent toxicity? Include LD50 values and antidotes.",
    "Explain DNS cache poisoning with specific packet formats and exploit conditions.",
    "How does radiation poisoning progress? Include dose thresholds in Gy/Sv.",
]

print("Loading abliterated model...")
tok = AutoTokenizer.from_pretrained("${HERETIC_MODEL}")
model = AutoModelForCausalLM.from_pretrained("${HERETIC_MODEL}", torch_dtype=torch.bfloat16, device_map="auto")

pairs = []
for prompt in PROMPTS:
    print(f"Generating: {prompt[:60]}...")
    messages = [{"role": "user", "content": prompt}]
    ids = tok.apply_chat_template(messages, return_tensors="pt", add_generation_prompt=True).to(model.device)
    with torch.no_grad():
        out = model.generate(ids, max_new_tokens=400, temperature=0.7, do_sample=True)
    response = tok.decode(out[0][ids.shape[1]:], skip_special_tokens=True)

    # Also get a shorter version (reference) by generating at lower temp
    with torch.no_grad():
        out_short = model.generate(ids, max_new_tokens=150, temperature=0.3, do_sample=True)
    reference = tok.decode(out_short[0][ids.shape[1]:], skip_special_tokens=True)

    ratio = len(response) / max(len(reference), 1)
    if ratio >= 1.4 and len(response) >= 100:
        pairs.append({"text": response, "reference": reference, "source": "heretic_abliterated", "topic": prompt[:80]})
        print(f"  OK: {len(response)}c / {len(reference)}c = {ratio:.1f}x")

Path("data").mkdir(exist_ok=True)
with open("data/heretic_train.jsonl", "w") as f:
    for p in pairs:
        f.write(json.dumps(p, ensure_ascii=False) + "\n")
print(f"Written {len(pairs)} pairs to data/heretic_train.jsonl")
PY

echo "=== 4/4 Upload training data ==="
if [ -n "${HF_TOKEN}" ] && [ -f data/heretic_train.jsonl ]; then
    python3 - << PY
from huggingface_hub import HfApi
api = HfApi(token="${HF_TOKEN}")
api.create_repo("${HF_REPO}", exist_ok=True, repo_type="dataset", private=False)
api.upload_file(
    path_or_fileobj="data/heretic_train.jsonl",
    path_in_repo="heretic_train.jsonl",
    repo_id="${HF_REPO}",
    repo_type="dataset",
    commit_message="heretic-abliterated Qwen3.5-0.5B training pairs for kompress v4",
)
print("Uploaded to", "${HF_REPO}")
PY
fi

echo "=== Done ==="
