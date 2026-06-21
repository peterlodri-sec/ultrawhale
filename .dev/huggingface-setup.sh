#!/bin/bash
# Setup HuggingFace dataset publishing
set -e

echo "▸ HuggingFace Dataset Setup"
echo ""
echo "Manual steps (run once):"
echo ""
echo "1. Create account: huggingface.co/join"
echo "2. Login: hf login"
echo "3. Create dataset:"
echo "   hf repo create ultrawhale-dogfood --type dataset"
echo ""
echo "4. Export + push:"
echo "   /dog-feed export"
echo "   cp ~/.ultrawhale/dogfeed/dogfeed-*.jsonl ."
echo "   git add *.jsonl && git commit -m 'dataset v1' && git push"
echo ""
echo "Live at: huggingface.co/datasets/peterlodri-sec/ultrawhale-dogfood"
