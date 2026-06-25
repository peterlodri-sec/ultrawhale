#!/usr/bin/env bash
# Kompress v16: v8 data + 10x must-keep weight
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v16"}
echo "=== v16: 10x must-keep weight ==="
python3 scripts/train_kompress_v16.py --data data/kompress_v8_train.jsonl --base-model chopratejas/kompress-v2-base --output kompress-v16-finetuned --epochs 3 --batch-size 16 --lr 2e-5
echo "=== Heretic eval ==="
python3 scripts/eval_heretic.py --model kompress-v16-finetuned --prompts-file data/heretic_expanded.jsonl || echo "WARN"
echo "=== ONNX + upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys,os,torch; sys.path.insert(0,"/workspace/ultrawhale")
from scripts.train_kompress import HeadroomCompressorModel,load_v2_weights
from transformers import AutoTokenizer
m=HeadroomCompressorModel("answerdotai/ModernBERT-base"); load_v2_weights(m,"kompress-v16-finetuned"); m.eval()
t=AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base"); d=t("hello world",return_tensors="pt")
class W(torch.nn.Module):
    def __init__(s,m): super().__init__(); s.m=m
    def forward(s,i,a): l,sp=s.m(i,a); p=torch.softmax(l,dim=-1)[:,:,1]; return p*(0.5+0.5*sp)
os.makedirs("kompress-v16-finetuned/onnx",exist_ok=True)
torch.onnx.export(W(m),(d["input_ids"],d["attention_mask"]),"kompress-v16-finetuned/onnx/kompress-fp32.onnx",input_names=["input_ids","attention_mask"],output_names=["final_scores"],dynamic_axes={"input_ids":{0:"b",1:"s"},"attention_mask":{0:"b",1:"s"},"final_scores":{0:"b",1:"s"}},opset_version=17)
print("ONNX exported")
PY
if [ -n "$HF_TOKEN" ]; then
    HF_REPO="$HF_REPO" python3 - << 'PY'
from huggingface_hub import HfApi; import os
api=HfApi(token=os.environ["HF_TOKEN"]); repo=os.environ["HF_REPO"]
api.create_repo(repo,exist_ok=True,private=False)
api.upload_folder(folder_path="kompress-v16-finetuned",repo_id=repo,commit_message="kompress-v16: v8 data + 10x must-keep weight")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v16 complete ==="
