#!/usr/bin/env bash
# vast.ai instance setup for Kompress fine-tuning
# Run once after `vastai ssh <instance_id>`
# Total GPU time target: < 2 hours (RTX 3090 @ $0.136/hr = < $0.30)
set -euo pipefail

echo "=== Kompress v3 fine-tune setup ==="

apt-get install -y -q gcc g++ 2>/dev/null || true
pip install -q --upgrade pip
# torch is already in pytorch/pytorch base image — pin versions compatible with torch 2.5+
pip install -q "transformers>=4.45,<5" "peft>=0.12" accelerate huggingface_hub sentence-transformers

# Clone ultrawhale scripts
git clone --depth=1 https://github.com/peterlodri-sec/ultrawhale.git /workspace/ultrawhale 2>/dev/null \
    || git -C /workspace/ultrawhale pull --ff-only

cd /workspace/ultrawhale

# Pre-cache ModernBERT tokenizer + weights
python3 -c "
from transformers import AutoTokenizer, AutoModel
tok = AutoTokenizer.from_pretrained('answerdotai/ModernBERT-base')
# download weights for later
print('ModernBERT tokenizer cached')
"

# Pre-cache kompress v2 checkpoint
python3 -c "
from huggingface_hub import hf_hub_download
path = hf_hub_download('chopratejas/kompress-v2-base', 'merged.pt')
print(f'Kompress v2 weights cached at {path}')
"

echo "=== Setup complete. Run: bash scripts/run_training.sh ==="
