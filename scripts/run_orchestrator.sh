#!/usr/bin/env bash
# Fine-tune kompress-superpower-orchestrator v2 — DoRA + NEFTune
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-superpower-orchestrator"}

echo "=== Orchestrator v2: DoRA + NEFTune ==="
echo "Data: $(wc -l < data/orchestrator_train.jsonl) pairs"
echo "Base: Qwen/Qwen2.5-7B-Instruct"

pip install -q transformers peft datasets torch accelerate bitsandbytes scipy 2>/dev/null || true

python3 scripts/train_orchestrator.py \
    --data data/orchestrator_train.jsonl \
    --output kompress-superpower-orchestrator \
    --epochs 3 \
    --lr 2e-4 \
    --neftune-alpha 5

echo "=== Upload ==="
if [ -n "$HF_TOKEN" ]; then
    python3 - << 'PY'
from huggingface_hub import HfApi; import os
api=HfApi(token=os.environ["HF_TOKEN"]); repo=os.environ["HF_REPO"]
api.create_repo(repo,exist_ok=True,private=False)
api.upload_folder(folder_path="kompress-superpower-orchestrator",repo_id=repo,
    commit_message="v2: DoRA + NEFTune — SOTA fine-tuning techniques")
print(f"Uploaded to {repo}")
PY
fi
echo "=== Done ==="
