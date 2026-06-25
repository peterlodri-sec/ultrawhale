#!/usr/bin/env python3
"""Evaluate a Kompress checkpoint against the ultrawhale test split.

Metrics:
  keep_rate    — fraction of tokens kept (lower = more compression)
  sem_sim      — cosine similarity of original vs compressed embeddings (bge-small)
  exact_pct    — % of must-keep tokens (numbers, identifiers) that survive

Usage:
    python eval_kompress.py --model kompress-v3-finetuned --data data/kompress_test.jsonl
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

import torch
from transformers import AutoTokenizer

BASE_ENCODER = "answerdotai/ModernBERT-base"

_MUST_KEEP_RE = re.compile(
    r"\d+(\.\d+)?|[A-Z_]{2,}|[a-z_]+\.[a-z_]+|/[a-z/._-]{2,}|\.[a-z]{2,4}\b|--?[a-zA-Z][\w-]*|\b[A-Z][a-z]+[A-Z]\w*"
)


def _compress_text(text: str, model, tokenizer, threshold: float = 0.5) -> tuple[str, float]:
    enc = tokenizer(text, return_tensors="pt", truncation=True, max_length=512)
    with torch.no_grad():
        logits, span = model(enc["input_ids"], enc["attention_mask"])
        token_probs = torch.softmax(logits, dim=-1)[0, :, 1]  # [L]
        span_scores = span[0]
        borderline = (token_probs > 0.3) & (token_probs <= threshold)
        keep = (token_probs > threshold) | (borderline & (span_scores > 0.5))

    tokens = tokenizer.convert_ids_to_tokens(enc["input_ids"][0])
    kept = [t for t, k in zip(tokens, keep) if k and t not in ("[CLS]", "[SEP]", "<s>", "</s>")]
    compressed = tokenizer.convert_tokens_to_string(kept)
    keep_rate = keep.float().mean().item()
    return compressed, keep_rate


def _exact_preserved(original: str, compressed: str) -> float:
    must_keep = [m.group(0) for m in _MUST_KEEP_RE.finditer(original)]
    if not must_keep:
        return 1.0
    preserved = sum(1 for t in must_keep if t in compressed)
    return preserved / len(must_keep)


def main(model_dir: str, data_path: str) -> None:
    from train_kompress import HeadroomCompressorModel, load_v2_weights

    tokenizer = AutoTokenizer.from_pretrained(BASE_ENCODER)
    model = HeadroomCompressorModel(BASE_ENCODER)
    try:
        load_v2_weights(model, model_dir)
    except Exception as e:
        print(f"Could not load weights: {e}", file=sys.stderr)
        sys.exit(1)
    model.eval()

    keep_rates, exact_pcts = [], []
    with open(data_path) as f:
        for line in f:
            d = json.loads(line.strip())
            text = d.get("text", "")
            if not text:
                continue
            compressed, kr = _compress_text(text, model, tokenizer)
            ep = _exact_preserved(text, compressed)
            keep_rates.append(kr)
            exact_pcts.append(ep)

    print(f"Samples evaluated: {len(keep_rates)}")
    print(f"Avg keep_rate:     {sum(keep_rates)/len(keep_rates):.4f}  (lower = more compression)")
    print(f"Avg exact_pct:     {sum(exact_pcts)/len(exact_pcts):.4f}  (must-keep token survival)")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", default="kompress-v3-finetuned")
    ap.add_argument("--data", default="data/kompress_test.jsonl")
    args = ap.parse_args()
    main(args.model, args.data)
