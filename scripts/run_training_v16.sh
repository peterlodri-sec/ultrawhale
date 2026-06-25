#!/usr/bin/env bash
# Kompress v16: v8 sweet spot + 10x must-keep weight
# Hypothesis: v8's 3x must-keep weight is too low. At 10x, the model
# prioritizes critical tokens more, potentially pushing heretic closer to 0.975.
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v16"}

echo "=== v16: v8 data (97 C3 + 200 generic) + 10x must-keep weight ==="

python3 - << 'PY'
import json, re, torch, logging, sys
from pathlib import Path
from torch.utils.data import DataLoader, Dataset
from transformers import AutoModel, AutoTokenizer
try: from peft import get_peft_model, LoraConfig; _PEFT=True
except: _PEFT=False

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)
ENCODER="answerdotai/ModernBERT-base"
_MR=re.compile(r"\d+(\.\d+)?"r"|[A-Z_]{2,}"r"|[a-z_]+\.[a-z_]+"r"|/[a-z/._-]{2,}"r"|\.[a-z]{2,4}\b"r"|--?[a-zA-Z][\w-]*"r"|\b[A-Z][a-z]+[A-Z]\w*")
def _ws(t): return set(re.findall(r"\b[a-z]{3,}\b",t.lower()))
def _sl(toks,ref):
    l,w=[],[]
    for t in toks:
        c=re.sub(r"[^\w]","",t).lower(); im=bool(_MR.search(t)); ir=c in ref or len(c)<3
        lb=1 if(im or ir)else 0; wt=10.0 if im else(1.0 if lb==1 else 0.5)
        l.append(lb); w.append(wt)
    return l,w
class KD(Dataset):
    def __init__(s,p,tk,ml=512): s.tk=tk; s.ml=ml; s.it=[]
        with open(p)as f:
            for l in f:
                d=json.loads(l.strip())
                if d.get("text")and d.get("reference"): s.it.append(d)
        log.info("Dataset: %d items",len(s.it))
    def __len__(s): return len(s.it)
    def __getitem__(s,i):
        it=s.it[i]; tx=it["text"]; rw=_ws(it["reference"])
        e=s.tk(tx,max_length=s.ml,truncation=True,padding=False,return_offsets_mapping=True,return_tensors="pt")
        ids=e["input_ids"][0]; am=e["attention_mask"][0]; toks=s.tk.convert_ids_to_tokens(ids)
        lb,wt=_sl(toks,rw)
        return{"input_ids":ids,"attention_mask":am,"labels":torch.tensor(lb,dtype=torch.float),"weights":torch.tensor(wt,dtype=torch.float)}
def _col(b):
    ml=max(x["input_ids"].size(0)for x in b)
    ids=torch.stack([torch.nn.functional.pad(x["input_ids"],(0,ml-x["input_ids"].size(0)),value=0)for x in b])
    am=torch.stack([torch.nn.functional.pad(x["attention_mask"],(0,ml-x["attention_mask"].size(0)),value=0)for x in b])
    lb=torch.stack([torch.nn.functional.pad(x["labels"],(0,ml-x["labels"].size(0)),value=0)for x in b])
    wt=torch.stack([torch.nn.functional.pad(x["weights"],(0,ml-x["weights"].size(0)),value=0)for x in b])
    return{"input_ids":ids,"attention_mask":am,"labels":lb,"weights":wt}
class HCM(torch.nn.Module):
    def __init__(s,eid): super().__init__(); s.enc=AutoModel.from_pretrained(eid); h=s.enc.config.hidden_size; s.h1=torch.nn.Linear(h,2); s.h2=torch.nn.Sequential(torch.nn.Conv1d(h,64,3,padding=1),torch.nn.ReLU(),torch.nn.Conv1d(64,1,3,padding=1))
    def forward(s,i,a): o=s.enc(input_ids=i,attention_mask=a); h=o.last_hidden_state; return s.h1(h),torch.sigmoid(s.h2(h.transpose(1,2)).squeeze(1))

device=torch.device("cuda"if torch.cuda.is_available()else"cpu"); log.info("Device: %s",device)
tok=AutoTokenizer.from_pretrained(ENCODER); ds=KD("data/kompress_v8_train.jsonl",tok)
loader=DataLoader(ds,batch_size=16,shuffle=True,collate_fn=_col)
model=HCM(ENCODER)
if _PEFT:
    lc=LoraConfig(r=16,lora_alpha=32,target_modules=["Wo","Wqkv"],lora_dropout=0.05,bias="none")
    model.enc=get_peft_model(model.enc,lc); log.info("LoRA applied")
model=model.to(device); opt=torch.optim.AdamW(model.parameters(),lr=2e-5); lfn=torch.nn.BCEWithLogitsLoss(reduction="none")
for ep in range(3):
    tl=0.0
    for st,b in enumerate(loader):
        ids=b["input_ids"].to(device); am=b["attention_mask"].to(device)
        lb=b["labels"].to(device); wt=b["weights"].to(device)
        lg,sp=model(ids,am); tl_g=lg[:,:,1]; rl=lfn(tl_g,lb)
        loss=(rl*wt).sum()/wt.sum(); opt.zero_grad(); loss.backward(); opt.step(); tl+=loss.item()
        if st==0: log.info("Epoch %d step 0/%d loss=%.4f",ep+1,len(loader),loss.item())
    al=tl/len(loader); log.info("Epoch %d avg loss=%.4f",ep+1,al)

if _PEFT: model.enc=model.enc.merge_and_unload(); log.info("LoRA merged")
od="kompress-v16-finetuned"; Path(od).mkdir(exist_ok=True)
# Save in v2 format
st=model.state_dict(); ek={k:v for k,v in st.items()if k.startswith("enc.")}
v2e={k[4:]:v for k,v in ek.items()}
v2s={"encoder_state_dict":v2e,"token_head_state_dict":{"weight":st["h1.weight"],"bias":st["h1.bias"]},"span_conv_state_dict":{"0.weight":st["h2.0.weight"],"0.bias":st["h2.0.bias"],"2.weight":st["h2.2.weight"],"2.bias":st["h2.2.bias"]}}
torch.save(v2s,f"{od}/merged.pt"); tok.save_pretrained(od); log.info("Saved to %s",od)
PY

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
