#!/usr/bin/env bash
# Kompress v14: GLM-5.1 council reviews training results — ship or retrain
#
# Standard v8 training (3 epochs, proven approach) + GLM council that
# reviews loss, heretic, keep_rate and makes ONE decision: SHIP or RETRAIN.
# If RETRAIN, increases epochs and continues. Simple, works with existing pipeline.
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v14"}
MAX_ROUNDS=3
EPOCHS=3

echo "=== Council-guided training (max $MAX_ROUNDS rounds, $EPOCHS epochs each) ==="

for round in $(seq 1 $MAX_ROUNDS); do
    echo ""
    echo "--- Round $round/$MAX_ROUNDS ---"
    
    echo "=== Training ($EPOCHS epochs) ==="
    python3 scripts/train_kompress.py \
        --data data/kompress_v14_train.jsonl \
        --base-model chopratejas/kompress-v2-base \
        --output kompress-v14-finetuned \
        --epochs $EPOCHS \
        --batch-size 16 \
        --lr 2e-5
    
    echo "=== Quick eval for council ==="
    EVAL_OUT=$(python3 scripts/eval_heretic.py \
        --model kompress-v14-finetuned \
        --prompts-file data/heretic_expanded.jsonl 2>&1)
    
    HERETIC=$(echo "$EVAL_OUT" | grep "AVERAGE" | awk '{print $3}')
    OVERRIDE=$(echo "$EVAL_OUT" | grep "exact_pct improvement" | grep -oP '[\d.]+')
    echo "  Heretic: $HERETIC, override_delta: +$OVERRIDE"
    
    echo "=== Council decision (GLM-5.1) ==="
    export COUNCIL_HERETIC="$HERETIC"
    export COUNCIL_OVERRIDE="$OVERRIDE"
    export COUNCIL_ROUND="$round"
    DECISION=$(python3 - << 'PY'
import os
from huggingface_hub import InferenceClient
h = float(os.environ.get("COUNCIL_HERETIC","0"))
o = float(os.environ.get("COUNCIL_OVERRIDE","0"))
r = int(os.environ.get("COUNCIL_ROUND","1"))

prompt = f"""Kompress model review after training round {r}/3.
heretic_exact={h:.3f} override_delta={o:.3f}
Target: heretic>=0.960 override=0. Best so far: v8=0.955 v2=0.975.

RULES: Round 1-2 ALWAYS retrain unless heretic>=0.965. Round 3 ship if heretic>=0.950.
Reply ONE word: SHIP or RETRAIN"""
try:
    client = InferenceClient(token=os.environ.get("HF_TOKEN",""))
    r = client.chat_completion(messages=[{"role":"user","content":prompt}],
        model="zai-org/GLM-5.1-FP8", max_tokens=5, temperature=0.1)
    d = r.choices[0].message.content.strip().upper()
    print("SHIP" if "SHIP" in d else "RETRAIN")
except Exception as e:
    print(f"RETRAIN")  # default to retrain on error
PY
    ) 2>&1
    
    echo "  Council says: $DECISION"
    
    if [ "$DECISION" = "SHIP" ]; then
        echo "Council approved! Shipping model."
        break
    fi
    
    echo "Council says retrain. Increasing epochs..."
    EPOCHS=$((EPOCHS + 2))
done

echo ""
echo "=== Final heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v14-finetuned \
    --prompts-file data/heretic_expanded.jsonl

echo "=== ONNX export + upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
sys.path.insert(0, "/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer
model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
load_v2_weights(model, "kompress-v14-finetuned")
model.eval()
tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
dummy = tok("hello world", return_tensors="pt")
class W(torch.nn.Module):
    def __init__(self,m): super().__init__(); self.m=m
    def forward(self,i,a): l,s=self.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]; return p*(0.5+0.5*s)
os.makedirs("kompress-v14-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(model),(dummy["input_ids"],dummy["attention_mask"]),
    "kompress-v14-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"],output_names=["final_scores"],
    dynamic_axes={"input_ids":{0:"b",1:"s"},"attention_mask":{0:"b",1:"s"},"final_scores":{0:"b",1:"s"}},
    opset_version=17)
print("ONNX exported")
PY

if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api = HfApi(token=os.environ["HF_TOKEN"])
repo = os.environ["HF_REPO"]
api.create_repo(repo, exist_ok=True, private=False)
api.upload_folder(folder_path="kompress-v14-finetuned", repo_id=repo,
    commit_message="kompress-v14: GLM-5.1 council reviewed — ship/retrain decisions")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v14 complete ==="
