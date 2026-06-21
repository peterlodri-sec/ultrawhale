#!/bin/bash
set -e
# vast.ai one-shot GPU training — SSH key injected via --onstart, not via attach

GPU=${1:-RTX_3060}
BUDGET=${2:-0.20}
DISK=${3:-20}

echo "▸ nocturne@vaked — vast.ai one-shot GPU training"
echo "  GPU: $GPU · Budget: $BUDGET/hr · Disk: ${DISK}GB"

# Get SSH public key
SSH_KEY=$(cat ~/.ssh/id_ed25519.pub 2>/dev/null || cat ~/.ssh/id_rsa.pub 2>/dev/null)
if [ -z "$SSH_KEY" ]; then
    echo "❌ No SSH key found at ~/.ssh/id_ed25519.pub"
    exit 1
fi

# Find cheapest GPU
echo "  Searching..."
OFFER=$(vastai search offers "gpu_name=$GPU" "dph<$BUDGET" --limit 1 --raw 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(d[0]['id'])" 2>/dev/null)
if [ -z "$OFFER" ]; then
    echo "❌ No GPU found: $GPU under $BUDGET/hr"
    exit 1
fi

# Create instance with SSH key injected via --onstart
echo "  Creating instance with onstart SSH injection..."
RESULT=$(vastai create instance "$OFFER" \
    --image pytorch \
    --disk "$DISK" \
    --ssh \
    --onstart "mkdir -p ~/.ssh && echo '$SSH_KEY' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys" \
    --raw 2>&1)

INSTANCE_ID=$(echo "$RESULT" | python3 -c "import sys,json; print(json.load(sys.stdin)['new_contract'])" 2>/dev/null)
echo "  Instance: $INSTANCE_ID"

# Wait for provisioning
echo "  Waiting for instance to be ready..."
for i in $(seq 1 30); do
    sleep 10
    STATUS=$(vastai show instance "$INSTANCE_ID" --raw 2>/dev/null | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('actual_status',''))" 2>/dev/null)
    if [ "$STATUS" = "running" ]; then
        echo "  🟢 Instance running"
        break
    fi
    echo "  ... waiting ($STATUS)"
done

# Connect and train
SSH_URL=$(vastai ssh-url "$INSTANCE_ID" 2>/dev/null)
echo "  SSH: $SSH_URL"
echo ""
echo "  Next: SSH in and run:"
echo "    apt update && apt install -y git python3-pip"
echo "    pip install transformers datasets torch peft"
echo "    git clone https://github.com/peterlodri-sec/ultrawhale.git"
echo "    cd ultrawhale && bash .dev/vastai-train.sh"
