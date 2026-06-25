#!/usr/bin/env bash
# Destroy current Kompress training instance and relaunch on the cheapest
# available RTX 4090 with good network.
#
# Usage:
#   bash scripts/vast_relaunch.sh [OLD_INSTANCE_ID]
#
# If OLD_INSTANCE_ID is omitted, skips the destroy step.
set -euo pipefail

OLD=${1:-}
IMAGE="pytorch/pytorch:2.5.1-cuda12.4-cudnn9-runtime"
HF_REPO=${HF_REPO:-"peterlodri-sec/kompress-v3"}

cat > /tmp/vast_onstart.sh << 'EOF'
#!/bin/bash
git clone --depth=1 https://github.com/peterlodri-sec/ultrawhale.git /workspace/ultrawhale
bash /workspace/ultrawhale/scripts/vast_setup.sh
bash /workspace/ultrawhale/scripts/run_training.sh
EOF

if [ -n "$OLD" ]; then
    echo "Destroying instance $OLD..."
    vastai destroy instance "$OLD"
fi

echo "Searching for cheapest RTX 4090 (inet_up >= 1000 Mbps)..."
OFFER_ID=$(vastai search offers \
    'gpu_name=RTX_4090 num_gpus=1 disk_space>=30 inet_up>=1000' \
    --order dph_total --limit 1 --raw \
    | python3 -c "import sys,json; offers=json.load(sys.stdin); print(offers[0]['id'])")

echo "Best offer: $OFFER_ID"

NEW=$(vastai create instance "$OFFER_ID" \
    --image "$IMAGE" \
    --disk 30 \
    --env "-e HF_TOKEN=${HF_TOKEN:-} -e HF_REPO=$HF_REPO" \
    --onstart /tmp/vast_onstart.sh \
    --raw \
    | python3 -c "import sys,json; r=json.load(sys.stdin); print(r['new_contract'])")

echo "Started instance $NEW"
echo ""
echo "Monitor:  vastai logs $NEW"
echo "Destroy:  vastai destroy instance $NEW"
echo "Relaunch: bash scripts/vast_relaunch.sh $NEW"
