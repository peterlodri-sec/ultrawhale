#!/usr/bin/env bash
# Kompress v14: GLM-5.1 "council" controls training loop
#
# After each epoch, GLM-5.1 reviews metrics (loss, heretic sample, keep_rate)
# and makes ONE decision: CONTINUE, STOP, or ADJUST (new LR).
# This replaces fixed epoch counts with intelligent convergence detection.
#
# Data: v8 sweet spot (97 C3 + 200 generic) + GLM scenarios (127 regex-labeled)
# Encoder: ModernBERT-base (149M)
# Base: chopratejas/kompress-v2-base
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v14"}
MAX_EPOCHS=10
pip install -q huggingface_hub 2>/dev/null || true

echo "=== 1/4 Merge v8 + GLM scenario data ==="
python3 - << 'PY'
import json, random
# v8 C3 data
c3 = [json.loads(l) for l in open("data/kompress_v8_train.jsonl")]
# GLM scenario data
glm = [json.loads(l) for l in open("data/kompress_v13_train.jsonl") if "glm_regex" in l.get("source","")]
print(f"  v8 C3: {len([r for r in c3 if 'qwen_c3' in r.get('source','')])} pairs")
print(f"  GLM regex: {len(glm)} pairs")
print(f"  Generic: {len([r for r in c3 if 'qwen_c3' not in r.get('source','')])} pairs")
merged = c3 + glm
random.seed(42); random.shuffle(merged)
with open("data/kompress_v14_train.jsonl","w") as f:
    for r in merged: f.write(json.dumps(r,ensure_ascii=False)+"\n")
print(f"  Total: {len(merged)}")
PY

echo "=== 2/4 Council-guided training loop ==="
python3 - << 'PY'
import json, re, torch, sys, os, time
from pathlib import Path
from torch.utils.data import DataLoader, Dataset
from transformers import AutoModel, AutoTokenizer
from huggingface_hub import InferenceClient
try:
    from peft import get_peft_model, LoraConfig
    _PEFT = True
except: _PEFT = False
import logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

ENCODER = "answerdotai/ModernBERT-base"
_MUST_KEEP_RE = re.compile(r"\d+(\.\d+)?"r"|[A-Z_]{2,}"r"|[a-z_]+\.[a-z_]+"r"|/[a-z/._-]{2,}"r"|\.[a-z]{2,4}\b"r"|--?[a-zA-Z][\w-]*"r"|\b[A-Z][a-z]+[A-Z]\w*")

def _word_set(text):
    return set(re.findall(r"\b[a-z]{3,}\b", text.lower()))

def _silver_labels(tokens, ref_words):
    labels, weights = [], []
    for tok in tokens:
        clean = re.sub(r"[^\w]","",tok).lower()
        is_must = bool(_MUST_KEEP_RE.search(tok))
        in_ref = clean in ref_words or len(clean)<3
        label = 1 if (is_must or in_ref) else 0
        weight = 3.0 if is_must else (1.0 if label==1 else 0.5)
        labels.append(label); weights.append(weight)
    return labels, weights

class KompressDataset(Dataset):
    def __init__(self, path, tokenizer, max_length=512):
        self.tok=tokenizer; self.ml=max_length; self.items=[]
        with open(path) as f:
            for l in f:
                d=json.loads(l.strip())
                if d.get("text") and d.get("reference"): self.items.append(d)
        log.info("Dataset: %d items", len(self.items))
    def __len__(self): return len(self.items)
    def __getitem__(self, idx):
        item=self.items[idx]; text=item["text"]; ref_words=_word_set(item["reference"])
        enc=self.tok(text,max_length=self.ml,truncation=True,padding=False,return_offsets_mapping=True,return_tensors="pt")
        ids=enc["input_ids"][0]; am=enc["attention_mask"][0]
        tokens=self.tok.convert_ids_to_tokens(ids)
        labels,weights=_silver_labels(tokens,ref_words)
        return {"input_ids":ids,"attention_mask":am,"labels":torch.tensor(labels,dtype=torch.float),"weights":torch.tensor(weights,dtype=torch.float)}

def collate(batch):
    ml=max(b["input_ids"].size(0) for b in batch)
    ids=torch.stack([torch.nn.functional.pad(b["input_ids"],(0,ml-b["input_ids"].size(0)),value=0) for b in batch])
    am=torch.stack([torch.nn.functional.pad(b["attention_mask"],(0,ml-b["attention_mask"].size(0)),value=0) for b in batch])
    labels=torch.stack([torch.nn.functional.pad(b["labels"],(0,ml-b["labels"].size(0)),value=0) for b in batch])
    weights=torch.stack([torch.nn.functional.pad(b["weights"],(0,ml-b["weights"].size(0)),value=0) for b in batch])
    return {"input_ids":ids,"attention_mask":am,"labels":labels,"weights":weights}

class HeadroomCompressorModel(torch.nn.Module):
    def __init__(self, encoder_id):
        super().__init__()
        self.encoder=AutoModel.from_pretrained(encoder_id)
        h=self.encoder.config.hidden_size
        self.head1=torch.nn.Linear(h,2)
        self.head2=torch.nn.Sequential(torch.nn.Conv1d(h,64,3,padding=1),torch.nn.ReLU(),torch.nn.Conv1d(64,1,3,padding=1))
    def forward(self,ids,am):
        out=self.encoder(input_ids=ids,attention_mask=am)
        h=out.last_hidden_state
        return self.head1(h),torch.sigmoid(self.head2(h.transpose(1,2)).squeeze(1))

device=torch.device("cuda" if torch.cuda.is_available() else "cpu")
log.info("Device: %s", device)

tok=AutoTokenizer.from_pretrained(ENCODER)
ds=KompressDataset("data/kompress_v14_train.jsonl",tok)
loader=DataLoader(ds,batch_size=16,shuffle=True,collate_fn=collate)
model=HeadroomCompressorModel(ENCODER)

# Load v2 weights as starting point
log.info("Loading v2-base weights...")
v2_state = torch.hub.load_state_dict_from_url(
    "https://huggingface.co/chopratejas/kompress-v2-base/resolve/main/merged.pt",
    map_location="cpu")
# Map v2 keys to our model (same architecture)
model.load_state_dict(v2_state, strict=False)
log.info("v2 weights loaded")

if _PEFT:
    lora=LoraConfig(r=16,lora_alpha=32,target_modules=["Wo","Wqkv"],lora_dropout=0.05,bias="none")
    model.encoder=get_peft_model(model.encoder,lora)
    log.info("LoRA applied")

model=model.to(device)
optimizer=torch.optim.AdamW(model.parameters(),lr=2e-5)
loss_fn=torch.nn.BCEWithLogitsLoss(reduction="none")

# Quick eval function for council
def quick_heretic(model, tok, n=5):
    """Fast heretic eval on n prompts for council decision."""
    prompts_path = "data/heretic_expanded.jsonl"
    if not Path(prompts_path).exists():
        return 0.0, 0.0
    prompts = [json.loads(l) for l in open(prompts_path)][:n]
    results = []
    for p in prompts:
        text = p["response"]
        enc = tok(text, return_tensors="pt", truncation=True, max_length=512)
        enc = {k: v.to(device) for k, v in enc.items()}
        with torch.no_grad():
            logits, span = model(enc["input_ids"], enc["attention_mask"])
            probs = torch.softmax(logits, dim=-1)[0,:,1]
            scores = probs * (0.5 + 0.5 * span[0])
            keep = scores > 0.5
        tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])
        kept = [t for t, k in zip(tokens, keep) if k and t not in ("[CLS]","[SEP]")]
        compressed = tok.convert_tokens_to_string(kept)
        must = [m.group(0) for m in _MUST_KEEP_RE.finditer(text)]
        kr = keep.float().mean().item()
        ex = sum(1 for m in must if m in compressed)/max(len(must),1)
        results.append((kr, ex))
    avg_kr = sum(r[0] for r in results)/len(results)
    avg_ex = sum(r[1] for r in results)/len(results)
    return avg_kr, avg_ex

# Council: GLM-5.1 makes one decision per epoch
council_client = None
def ask_council(epoch, loss, kr, ex, prev_loss):
    global council_client
    if council_client is None:
        try:
            council_client = InferenceClient(token=os.environ.get("HF_TOKEN",""))
        except:
            return "CONTINUE"
    
    loss_delta = loss - prev_loss if prev_loss else 0
    prompt = f"""You control a model training loop. After epoch {epoch}, metrics are:
- Training loss: {loss:.4f} (Δ from prev: {loss_delta:+.4f})
- Heretic exact (5 prompts): {ex:.3f}
- Keep rate: {kr:.3f}
- Previous epoch loss: {prev_loss:.4f}

Target: heretic >= 0.965, loss minimizing, keep_rate 0.75-0.88.
Convergence: loss stops decreasing for 2+ epochs → STOP.
Overfit warning: loss < 0.1 AND heretic < 0.90 → STOP (overfit).

Decide ONE action:
- CONTINUE (training is progressing well)
- STOP (converged or overfitting)
- LR_HALF (reduce learning rate, continue)
- LR_DOUBLE (loss plateauing, increase LR)

Reply with ONLY the action word."""

    try:
        r = council_client.chat_completion(
            messages=[{"role":"user","content":prompt}],
            model="zai-org/GLM-5.1-FP8", max_tokens=10, temperature=0.1)
        decision = r.choices[0].message.content.strip().upper()
        for action in ["CONTINUE","STOP","LR_HALF","LR_DOUBLE"]:
            if action in decision:
                return action
        return "CONTINUE"
    except:
        return "CONTINUE"

# Training loop with council
prev_loss = None
for epoch in range(MAX_EPOCHS):
    model.train()
    total_loss = 0.0
    for step, batch in enumerate(loader):
        ids=batch["input_ids"].to(device); am=batch["attention_mask"].to(device)
        labels=batch["labels"].to(device); weights=batch["weights"].to(device)
        logits,span=model(ids,am)
        token_logits=logits[:,:,1]
        raw_loss=loss_fn(token_logits,labels)
        loss=(raw_loss*weights).sum()/weights.sum()
        optimizer.zero_grad(); loss.backward(); optimizer.step()
        total_loss+=loss.item()
        if step==0:
            log.info("Epoch %d step 0/%d loss=%.4f", epoch+1, len(loader), loss.item())
    
    avg_loss = total_loss/len(loader)
    log.info("Epoch %d avg loss=%.4f", epoch+1, avg_loss)
    
    # Quick eval for council
    model.eval()
    kr, ex = quick_heretic(model, tok, n=5)
    log.info("Quick heretic: kr=%.3f ex=%.3f", kr, ex)
    
    # Council decision
    decision = ask_council(epoch+1, avg_loss, kr, ex, prev_loss or avg_loss)
    log.info("COUNCIL (epoch %d): %s", epoch+1, decision)
    
    if decision == "STOP":
        log.info("Council says STOP. Converged at epoch %d.", epoch+1)
        break
    elif decision == "LR_HALF":
        for g in optimizer.param_groups:
            g["lr"] *= 0.5
        log.info("LR halved to %.2e", optimizer.param_groups[0]["lr"])
    elif decision == "LR_DOUBLE":
        for g in optimizer.param_groups:
            g["lr"] *= 2.0
        log.info("LR doubled to %.2e", optimizer.param_groups[0]["lr"])
    
    prev_loss = avg_loss

# Merge LoRA and save
if _PEFT:
    model.encoder = model.encoder.merge_and_unload()
    log.info("LoRA merged")

out_dir="kompress-v14-finetuned"
Path(out_dir).mkdir(exist_ok=True)
torch.save(model.state_dict(),f"{out_dir}/merged.pt")
tok.save_pretrained(out_dir)
log.info("Saved to %s after %d epochs", out_dir, epoch+1)
PY

echo "=== 3/4 Heretic eval (32 prompts) ==="
python3 scripts/eval_heretic.py \
    --model kompress-v14-finetuned \
    --prompts-file data/heretic_expanded.jsonl || echo "WARN: eval non-fatal"

echo "=== 4/4 ONNX + upload ==="
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
    commit_message="kompress-v14: GLM-5.1 council controls training — intelligent convergence + v8+GLM data")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v14 complete ==="
