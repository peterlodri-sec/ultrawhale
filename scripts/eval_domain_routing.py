#!/usr/bin/env python3
"""Measure per-domain compression rates to find optimal bias thresholds.

Runs kompress-v4 on synthetic agent data across each domain at multiple
bias values. Outputs: tokens_removed, mk_survival_rate per (domain, bias).

Usage:
    python3 scripts/eval_domain_routing.py --model PeetPedro/kompress-v4
"""
from __future__ import annotations

import argparse
import json
import random
import re
import sys
from collections import defaultdict
from pathlib import Path

_MUST_KEEP_RE = re.compile(
    r"\b0x[0-9A-Fa-f]+\b"
    r"|(?<![\w.])\d+(?:\.\d+)?(?![\w.])"
    r"|[A-Z_]{2,}"
    r"|[a-z_][a-z0-9_]*\.[a-z0-9_]+"
    r"|/[a-z0-9/._-]{2,}"
    r"|\.[a-z]{2,4}\b"
    r"|--?[a-zA-Z][\w-]*"
    r"|\b[A-Z][a-z]+[A-Z]\w*"
)

BIAS_VALUES = [0.5, 0.7, 0.85, 1.0, 1.2, 1.5]
SAMPLES_PER_DOMAIN = 30


def load_model(model_id: str):
    import torch
    sys.path.insert(0, str(Path(__file__).parent.parent))
    from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
    from transformers import AutoTokenizer

    tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
    model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
    load_v2_weights(model, model_id)
    model.eval()
    device = "cpu"
    model = model.to(device)
    return tok, model, device


def compress(text: str, tok, model, device, bias: float) -> str:
    import torch
    enc = tok(text, return_tensors="pt", truncation=True, max_length=512, padding=True)
    enc = {k: v.to(device) for k, v in enc.items()}
    with torch.no_grad():
        logits, span = model(enc["input_ids"], enc["attention_mask"])
        probs = torch.softmax(logits, dim=-1)[0, :, 1]
        scores = probs * (0.5 + 0.5 * span[0]) * bias
        keep = scores > 0.5
    tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])
    # Must-keep override with sliding window
    n = len(tokens)
    for i in range(n):
        for w in (1, 2, 3):
            if i + w > n:
                break
            window = tok.convert_tokens_to_string(tokens[i:i+w]).strip()
            if _MUST_KEEP_RE.search(window):
                for j in range(i, i + w):
                    keep[j] = True
                break
    kept = [t for t, k in zip(tokens, keep) if k and t not in ("[CLS]","[SEP]","<s>","</s>")]
    return tok.convert_tokens_to_string(kept)


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", default="PeetPedro/kompress-v4")
    ap.add_argument("--data", default="data/kompress_agent_train.jsonl")
    ap.add_argument("--samples", type=int, default=SAMPLES_PER_DOMAIN)
    args = ap.parse_args()

    print(f"Loading {args.model}...")
    tok, model, device = load_model(args.model)

    rows = [json.loads(l) for l in open(args.data)]
    by_source: dict[str, list] = defaultdict(list)
    for r in rows:
        source = r.get("source", "unknown").replace("self_labeled_v4_", "").replace("self_labeled_v6_sw_", "")
        by_source[source].append(r)

    rng = random.Random(42)
    print(f"\nDomain routing eval — {args.model}")
    print(f"{'Domain':<25} {'Bias':>6} {'Ratio':>7} {'MK%':>6} {'Tokens↓':>8}")
    print("-" * 60)

    results: dict[str, dict] = {}
    for source, domain_rows in sorted(by_source.items()):
        sample = rng.sample(domain_rows, min(args.samples, len(domain_rows)))
        results[source] = {}
        for bias in BIAS_VALUES:
            ratios, mk_surv = [], []
            for r in sample:
                text = r["text"]
                compressed = compress(text, tok, model, device, bias)
                ratio = len(text) / max(len(compressed), 1)
                ratios.append(ratio)
                must = [m.group(0) for m in _MUST_KEEP_RE.finditer(text)]
                survived = sum(1 for m in must if m in compressed)
                mk_surv.append(survived / max(len(must), 1))
            avg_ratio = sum(ratios) / len(ratios)
            avg_mk = sum(mk_surv) / len(mk_surv)
            avg_tokens_removed = sum(len(r["text"]) - len(compress(r["text"], tok, model, device, bias))
                                     for r in sample[:5]) / 5  # estimate
            results[source][bias] = (avg_ratio, avg_mk)
            print(f"  {source:<23} {bias:>6.2f} {avg_ratio:>7.2f}x {avg_mk*100:>5.1f}% {'↑' if bias < 1.0 else ''}")

    print("\n=== RECOMMENDED BIAS PER DOMAIN ===")
    print("Criteria: keep mk_survival >= 0.95, maximize ratio")
    for source, bias_results in results.items():
        best_bias = max(
            (b for b, (ratio, mk) in bias_results.items() if mk >= 0.95),
            key=lambda b: bias_results[b][0],
            default=1.0,
        )
        ratio, mk = bias_results[best_bias]
        print(f"  {source:<25} bias={best_bias:.2f}  ratio={ratio:.2f}x  mk={mk*100:.1f}%")


if __name__ == "__main__":
    main()
