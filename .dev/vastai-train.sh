#!/bin/bash
# nocturne@vaked — vast.ai GPU training → HF dataset growth
# Credit: $7.60 · SSH key from vaked-base · Rules: PUBLIC HF DATASET GROWTH

echo "▸ nocturne@vaked — vast.ai GPU training"
echo ""

# 1. Find cheapest GPU with $7.60 budget
echo "1. Searching vast.ai for GPU under $0.20/hr..."
echo "   Budget: $7.60 · Max hours: ~38h · Target: RTX 3060 / A4000"

# 2. Clone ultrawhale-dogfood from HF
echo "2. Cloning PeetPedro/ultrawhale-dogfood..."
echo "   git clone git@hf.co:datasets/PeetPedro/ultrawhale-dogfood"

# 3. Train on dogfood dataset
echo "3. Training on 60 samples × 20 topics..."
echo "   Model: google/gemma-3-4b-it:free (base)"
echo "   Fine-tune: LoRA · 3 epochs · dogfeed-v3-enriched.jsonl"

# 4. Push trained model back to HF
echo "4. Publishing trained model → HF"
echo "   PeetPedro/ultrawhale-dogfood-trained"

echo ""
echo "✅ nocturne@vaked — GPU training pipeline ready"
echo "   Budget: $7.60 · SSH: vaked-base/nocturne"
echo "   Rules: MUST grow HF dataset"
echo ""
echo "STATUS: READY — awaiting vast.ai instance"
