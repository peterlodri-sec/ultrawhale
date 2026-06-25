#!/usr/bin/env bash
# Kompress v11: ModernBERT-large encoder (352M) + C3 teacher labels
#
# Hypothesis: ModernBERT-base (149M) may be capacity-bottlenecked.
# The 0.020 heretic gap to v2-base (0.975) might be from the encoder
# not having enough capacity for both aggressive compression AND
# rare-pattern preservation. A 2.25x larger encoder tests this.
#
# Encoder: answerdotai/ModernBERT-large (352M, same tokenizer)
# Data:    v8 sweet spot — 97 C3 Qwen-labeled + 200 generic
# Training: from scratch (random heads), 5 epochs, LoRA r=16
# Target:  heretic >= 0.965 (beats v8's 0.955)
set -euo pipefail
cd /workspace/ultrawhale
HF_TOKEN=${HF_TOKEN:-}
HF_REPO=${HF_REPO:-"PeetPedro/kompress-v11"}

echo "=== 1/4 Training data: $(wc -l < data/kompress_v8_train.jsonl) pairs ==="

echo "=== 2/4 Train from ModernBERT-large (352M, from scratch) ==="
python3 - << 'PY'
import sys, json, re, torch, logging, argparse
from pathlib import Path
from torch.utils.data import DataLoader, Dataset
from transformers import AutoModel, AutoTokenizer
try:
    from peft import get_peft_model, LoraConfig
    _PEFT = True
except ImportError:
    _PEFT = False

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger(__name__)

# ── Larger encoder ──
ENCODER = "answerdotai/ModernBERT-large"  # 352M vs 149M base

_MUST_KEEP_RE = re.compile(
    r"\d+(\.\d+)?" r"|[A-Z_]{2,}" r"|[a-z_]+\.[a-z_]+"
    r"|/[a-z/._-]{2,}" r"|\.[a-z]{2,4}\b" r"|--?[a-zA-Z][\w-]*"
    r"|\b[A-Z][a-z]+[A-Z]\w*"
)

def _word_set(text):
    return set(re.findall(r"\b[a-z]{3,}\b", text.lower()))

def _silver_labels(tokens, ref_words):
    labels, weights = [], []
    for tok in tokens:
        clean = re.sub(r"[^\w]", "", tok).lower()
        is_must = bool(_MUST_KEEP_RE.search(tok))
        in_ref = clean in ref_words or len(clean) < 3
        label = 1 if (is_must or in_ref) else 0
        weight = 3.0 if is_must else (1.0 if label == 1 else 0.5)
        labels.append(label); weights.append(weight)
    return labels, weights

class KompressDataset(Dataset):
    def __init__(self, path, tokenizer, max_length=512):
        self.tokenizer = tokenizer; self.max_length = max_length
        self.items = []
        with open(path) as f:
            for line in f:
                d = json.loads(line.strip())
                if d.get("text") and d.get("reference"):
                    self.items.append(d)
        log.info("Dataset: %d items", len(self.items))
    def __len__(self):
        return len(self.items)
    def __getitem__(self, idx):
        item = self.items[idx]
        text = item["text"]
        ref_words = _word_set(item["reference"])
        enc = self.tokenizer(text, max_length=self.max_length, truncation=True,
                             padding=False, return_offsets_mapping=True, return_tensors="pt")
        input_ids = enc["input_ids"][0]
        attention_mask = enc["attention_mask"][0]
        tokens = self.tokenizer.convert_ids_to_tokens(input_ids)
        labels, weights = _silver_labels(tokens, ref_words)
        return {
            "input_ids": input_ids,
            "attention_mask": attention_mask,
            "labels": torch.tensor(labels, dtype=torch.float),
            "weights": torch.tensor(weights, dtype=torch.float),
        }

def collate_fn(batch):
    max_len = max(b["input_ids"].size(0) for b in batch)
    pad_id = 0
    ids = torch.stack([torch.nn.functional.pad(b["input_ids"], (0, max_len-b["input_ids"].size(0)), value=pad_id) for b in batch])
    mask = torch.stack([torch.nn.functional.pad(b["attention_mask"], (0, max_len-b["attention_mask"].size(0)), value=0) for b in batch])
    labels = torch.stack([torch.nn.functional.pad(b["labels"], (0, max_len-b["labels"].size(0)), value=0) for b in batch])
    weights = torch.stack([torch.nn.functional.pad(b["weights"], (0, max_len-b["weights"].size(0)), value=0) for b in batch])
    return {"input_ids": ids, "attention_mask": mask, "labels": labels, "weights": weights}

class HeadroomCompressorModel(torch.nn.Module):
    def __init__(self, encoder_id):
        super().__init__()
        self.encoder = AutoModel.from_pretrained(encoder_id)
        hidden = self.encoder.config.hidden_size
        self.head1 = torch.nn.Linear(hidden, 2)       # token classifier
        self.head2 = torch.nn.Sequential(              # span CNN
            torch.nn.Conv1d(hidden, 64, 3, padding=1),
            torch.nn.ReLU(),
            torch.nn.Conv1d(64, 1, 3, padding=1),
        )
    def forward(self, input_ids, attention_mask):
        out = self.encoder(input_ids=input_ids, attention_mask=attention_mask)
        hidden_states = out.last_hidden_state
        logits = self.head1(hidden_states)
        span_feat = hidden_states.transpose(1, 2)
        span = torch.sigmoid(self.head2(span_feat)).squeeze(1)
        return logits, span

log.info("Device: %s", "cuda" if torch.cuda.is_available() else "cpu")
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

tok = AutoTokenizer.from_pretrained(ENCODER)
ds = KompassDataset("data/kompress_v8_train.jsonl", tok)
loader = DataLoader(ds, batch_size=8, shuffle=True, collate_fn=collate_fn)

model = HeadroomCompressorModel(ENCODER)

if _PEFT:
    lora_config = LoraConfig(r=16, lora_alpha=32, target_modules=["Wo","Wqkv"],
                             lora_dropout=0.05, bias="none")
    model.encoder = get_peft_model(model.encoder, lora_config)
    log.info("LoRA applied to encoder")
else:
    log.warning("PEFT not available, full fine-tuning")

model = model.to(device)
optimizer = torch.optim.AdamW(model.parameters(), lr=1e-4)
loss_fn = torch.nn.BCEWithLogitsLoss(reduction="none")

for epoch in range(5):
    total_loss = 0.0
    for step, batch in enumerate(loader):
        ids = batch["input_ids"].to(device)
        am = batch["attention_mask"].to(device)
        labels = batch["labels"].to(device)
        weights = batch["weights"].to(device)

        logits, span = model(ids, am)
        token_logits = logits[:, :, 1]
        raw_loss = loss_fn(token_logits, labels)
        masked = raw_loss * weights
        loss = masked.sum() / weights.sum()

        optimizer.zero_grad()
        loss.backward()
        optimizer.step()
        total_loss += loss.item()

        if step == 0:
            log.info("Epoch %d step 0/%d loss=%.4f", epoch+1, len(loader), loss.item())

    avg_loss = total_loss / len(loader)
    log.info("Epoch %d avg loss=%.4f", epoch+1, avg_loss)

# Save
out_dir = "kompress-v11-finetuned"
Path(out_dir).mkdir(exist_ok=True)
torch.save(model.state_dict(), f"{out_dir}/merged.pt")
tok.save_pretrained(out_dir)
log.info("Saved to %s", out_dir)
PY

echo "=== 3/4 Heretic eval (32 prompts) ==="
python3 - << 'PY'
import sys, torch, json, re, argparse
sys.path.insert(0, "/workspace/ultrawhale")
from pathlib import Path
from transformers import AutoModel, AutoTokenizer

ENCODER = "answerdotai/ModernBERT-large"
_MUST_KEEP_RE = re.compile(
    r"\d+(\.\d+)?" r"|[A-Z_]{2,}" r"|[a-z_]+\.[a-z_]+"
    r"|/[a-z/._-]{2,}" r"|\.[a-z]{2,4}\b" r"|--?[a-zA-Z][\w-]*"
    r"|\b[A-Z][a-z]+[A-Z]\w*"
)

class HeadroomCompressorModel(torch.nn.Module):
    def __init__(self, encoder_id):
        super().__init__()
        self.encoder = AutoModel.from_pretrained(encoder_id)
        hidden = self.encoder.config.hidden_size
        self.head1 = torch.nn.Linear(hidden, 2)
        self.head2 = torch.nn.Sequential(
            torch.nn.Conv1d(hidden, 64, 3, padding=1),
            torch.nn.ReLU(),
            torch.nn.Conv1d(64, 1, 3, padding=1),
        )
    def forward(self, input_ids, attention_mask):
        out = self.encoder(input_ids=input_ids, attention_mask=attention_mask)
        h = out.last_hidden_state
        return self.head1(h), torch.sigmoid(self.head2(h.transpose(1,2)).squeeze(1))

tok = AutoTokenizer.from_pretrained(ENCODER)
model = HeadroomCompressorModel(ENCODER)
model.load_state_dict(torch.load("kompress-v11-finetuned/merged.pt", map_location="cpu"))
model.eval()

heretic_path = "data/heretic_expanded.jsonl"
prompts = [json.loads(l) for l in open(heretic_path)]
print(f"Loaded {len(prompts)} prompts from {heretic_path}")

results = []
for p in prompts:
    text = p["response"]
    enc = tok(text, return_tensors="pt", truncation=True, max_length=512)
    with torch.no_grad():
        logits, span = model(enc["input_ids"], enc["attention_mask"])
        probs = torch.softmax(logits, dim=-1)[0,:,1]
        scores = probs * (0.5 + 0.5 * span[0])
        keep = scores > 0.5

    tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])
    # Hard override
    keep_ov = keep.clone()
    for i, t in enumerate(tokens):
        word = tok.convert_tokens_to_string([t]).strip()
        if _MUST_KEEP_RE.search(word):
            keep_ov[i] = True

    def metrics(km):
        kept = [t for t, k in zip(tokens, km) if k and t not in ("[CLS]","[SEP]","<s>","</s>")]
        compressed = tok.convert_tokens_to_string(kept)
        must = [m.group(0) for m in _MUST_KEEP_RE.finditer(text)]
        kr = km.float().mean().item()
        ex = sum(1 for m in must if m in compressed) / max(len(must),1)
        return kr, ex

    kr_b, ex_b = metrics(keep)
    kr_o, ex_o = metrics(keep_ov)
    results.append({"ex_b": ex_b, "ex_o": ex_o, "kr_b": kr_b, "kr_o": kr_o})
    print(f"  {p['prompt'][:45]:<45} {kr_b:>8.3f} {ex_b:>8.3f} {kr_o:>8.3f} {ex_o:>8.3f}")

avg_ex_b = sum(r["ex_b"] for r in results)/len(results)
avg_ex_o = sum(r["ex_o"] for r in results)/len(results)
avg_kr_b = sum(r["kr_b"] for r in results)/len(results)
avg_kr_o = sum(r["kr_o"] for r in results)/len(results)
print(f"{'AVERAGE':<45} {avg_kr_b:>8.3f} {avg_ex_b:>8.3f} {avg_kr_o:>8.3f} {avg_ex_o:>8.3f}")
print(f"\nexact_pct improvement from override: +{avg_ex_o - avg_ex_b:.3f}")
PY

echo "=== 4/4 ONNX export + upload ==="
pip install -q onnx onnxruntime 2>/dev/null || true
python3 - << 'PY'
import sys, os, torch
from pathlib import Path
from transformers import AutoModel, AutoTokenizer

ENCODER = "answerdotai/ModernBERT-large"

class HeadroomCompressorModel(torch.nn.Module):
    def __init__(self, encoder_id):
        super().__init__()
        self.encoder = AutoModel.from_pretrained(encoder_id)
        hidden = self.encoder.config.hidden_size
        self.head1 = torch.nn.Linear(hidden, 2)
        self.head2 = torch.nn.Sequential(
            torch.nn.Conv1d(hidden, 64, 3, padding=1),
            torch.nn.ReLU(),
            torch.nn.Conv1d(64, 1, 3, padding=1),
        )
    def forward(self, input_ids, attention_mask):
        out = self.encoder(input_ids=input_ids, attention_mask=attention_mask)
        h = out.last_hidden_state
        return self.head1(h), torch.sigmoid(self.head2(h.transpose(1,2)).squeeze(1))

model = HeadroomCompressorModel(ENCODER)
model.load_state_dict(torch.load("kompress-v11-finetuned/merged.pt", map_location="cpu"))
model.eval()

class Wrapper(torch.nn.Module):
    def __init__(self, m):
        super().__init__(); self.m = m
    def forward(self, i, a):
        l, s = self.m(i, a)
        p = torch.softmax(l, dim=-1)[:,:,1]
        return p * (0.5 + 0.5 * s)

tok = AutoTokenizer.from_pretrained(ENCODER)
dummy = tok("hello world", return_tensors="pt")
os.makedirs("kompress-v11-finetuned/onnx", exist_ok=True)
torch.onnx.export(Wrapper(model), (dummy["input_ids"], dummy["attention_mask"]),
    "kompress-v11-finetuned/onnx/kompress-fp32.onnx",
    input_names=["input_ids","attention_mask"], output_names=["final_scores"],
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
api.upload_folder(folder_path="kompress-v11-finetuned", repo_id=repo,
    commit_message="kompress-v11: ModernBERT-large (352M) encoder + C3 teacher labels")
print(f"Uploaded to {repo}")
PY
fi
echo "=== v11 complete ==="
