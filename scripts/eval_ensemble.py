#!/usr/bin/env python3
"""
Voting ensemble: v3, v4, v5 majority vote on token keep/drop.
Anthropic pattern: parallelization → voting.
Each model votes; token kept if >= THRESHOLD models vote keep OR override fires.
"""
import argparse, json, re, torch, sys
sys.path.insert(0, ".")
from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
from transformers import AutoTokenizer

BASE = "answerdotai/ModernBERT-base"
_MK = re.compile(
    r"\b0x[0-9A-Fa-f]+\b|(?<![\w.])\d+(?:\.\d+)?(?![\w.])|[A-Z_]{2,}"
    r"|[a-z_][a-z0-9_]*\.[a-z0-9_]+|/[a-z0-9/._-]{2,}"
    r"|\.[a-z]{2,4}\b|--?[a-zA-Z][\w-]*|\b[A-Z][a-z]+[A-Z]\w*"
)

ENSEMBLE = [
    ("v3", "PeetPedro/kompress-v3"),
    ("v4", "PeetPedro/kompress-v4"),
    ("v5", "PeetPedro/kompress-v5"),
]

def load_models(tok):
    models = []
    for name, mid in ENSEMBLE:
        m = HeadroomCompressorModel(BASE)
        load_v2_weights(m, mid)
        m.eval()
        models.append((name, m))
        print(f"  Loaded {name}")
    return models

def compress_ensemble(text, models, tok, threshold=2, with_override=True):
    """Keep token if >= threshold models vote keep, OR override fires."""
    enc = tok(text, return_tensors="pt", truncation=True, max_length=512)
    tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])
    votes = torch.zeros(len(tokens))

    for name, model in models:
        with torch.no_grad():
            logits, span = model(enc["input_ids"], enc["attention_mask"])
            probs = torch.softmax(logits, dim=-1)[0, :, 1]
            scores = probs * (0.5 + 0.5 * span[0])
            votes += (scores > 0.5).float()

    keep = votes >= threshold

    if with_override:
        for i, t in enumerate(tokens):
            w = tok.convert_tokens_to_string([t]).strip()
            if _MK.search(w):
                keep[i] = True

    kept = [t for t, k in zip(tokens, keep)
            if k and t not in ("[CLS]","[SEP]","<s>","</s>")]
    return tok.convert_tokens_to_string(kept), keep.float().mean().item()

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data/kompress_test.jsonl")
    ap.add_argument("--threshold", type=int, default=2, help="votes needed to keep (default 2/3)")
    ap.add_argument("--n", type=int, default=50)
    args = ap.parse_args()

    tok = AutoTokenizer.from_pretrained(BASE)
    print("Loading ensemble models...")
    models = load_models(tok)

    records = [json.loads(l) for l in open(args.data)][:args.n]
    keep_rates, exact_pcts = [], []

    for r in records:
        compressed, kr = compress_ensemble(r["text"], models, tok, args.threshold)
        must = [m.group(0) for m in _MK.finditer(r["text"])]
        if must:
            exact_pcts.append(sum(1 for m in must if m in compressed) / len(must))
        keep_rates.append(kr)

    print(f"\nEnsemble (threshold={args.threshold}/3, n={len(records)})")
    print(f"  keep_rate:  {sum(keep_rates)/len(keep_rates):.3f}")
    print(f"  exact_pct:  {sum(exact_pcts)/len(exact_pcts):.3f}")

if __name__ == "__main__":
    main()
