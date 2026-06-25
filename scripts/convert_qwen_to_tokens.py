#!/usr/bin/env python3
"""Convert Qwen2.5-7B character-span labels to ModernBERT token keep/drop labels.

Usage:
    python3 scripts/convert_qwen_to_tokens.py \
        --input data/c3_qwen_labeled.jsonl \
        --output data/c3_train.jsonl
"""
from __future__ import annotations

import argparse, json, random, sys
from pathlib import Path
from transformers import AutoTokenizer

TOKENIZER_ID = "answerdotai/ModernBERT-base"


def spans_to_token_labels(text: str, spans: list[dict], tokenizer) -> tuple[list[str], list[int]]:
    """Convert char-level spans to token-level keep/drop labels.

    Returns (tokens, labels) where labels: 1=must-keep, 0=can-drop.
    Labels align with tokenizer output after stripping special tokens.
    """
    encoding = tokenizer(text, truncation=True, max_length=512, padding=False)
    tokens = tokenizer.convert_ids_to_tokens(encoding["input_ids"])
    offsets = encoding.encodings[0].offsets  # list of (char_start, char_end) per token

    # Build a char-level mask: which chars are inside any must-keep span
    char_keep = [False] * len(text)
    for span in spans:
        for c in range(span["start"], span["end"]):
            char_keep[c] = True

    # Per token: keep if ANY of its characters fall in a must-keep span
    labels = []
    valid_tokens = []
    for i, (tok, (cs, ce)) in enumerate(zip(tokens, offsets)):
        if tok in ("[CLS]", "[SEP]", "<s>", "</s>", "<pad>"):
            continue
        if cs == ce == 0:
            # Special token with no offset
            continue
        # Check if any char in [cs, ce) overlaps a must-keep span
        overlap = any(char_keep[c] for c in range(cs, min(ce, len(text))))
        valid_tokens.append(tok)
        labels.append(1 if overlap else 0)

    return valid_tokens, labels


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--input", default="data/c3_qwen_labeled.jsonl")
    ap.add_argument("--output", default="data/c3_train.jsonl")
    ap.add_argument("--min-keep-ratio", type=float, default=0.05,
                    help="Min fraction of tokens labeled keep (filter empty)")
    ap.add_argument("--max-keep-ratio", type=float, default=0.80,
                    help="Max fraction of tokens labeled keep (filter all-keep)")
    args = ap.parse_args()

    inp = Path(__file__).parent.parent / args.input
    out = Path(__file__).parent.parent / args.output

    if not inp.exists():
        print(f"ERROR: {inp} not found. Run Qwen labeling first.")
        sys.exit(1)

    rows = [json.loads(l) for l in open(inp)]

    print(f"Loading tokenizer: {TOKENIZER_ID}")
    tok = AutoTokenizer.from_pretrained(TOKENIZER_ID)

    records = []
    skipped_empty = 0
    skipped_all = 0
    total_tokens = 0
    total_kept = 0

    for row in rows:
        text = row["text"]
        spans = row.get("spans", [])
        if not spans:
            skipped_empty += 1
            continue

        tokens, labels = spans_to_token_labels(text, spans, tok)

        keep_ratio = sum(labels) / max(len(labels), 1)
        if keep_ratio < args.min_keep_ratio:
            skipped_empty += 1
            continue
        if keep_ratio > args.max_keep_ratio:
            skipped_all += 1
            continue

        # Only keep tokens that have content (strip CLS/SEP noise)
        # The labels align 1:1 with the filtered token list
        records.append({
            "text": text,
            "tokens": tokens,
            "labels": labels,
            "domain": row.get("domain", "unknown"),
            "source": "qwen_c3_labeled",
        })
        total_tokens += len(labels)
        total_kept += sum(labels)

    rng = random.Random(42)
    rng.shuffle(records)

    with open(out, "w") as f:
        for r in records:
            f.write(json.dumps(r, ensure_ascii=False) + "\n")

    keep_pct = (total_kept / max(total_tokens, 1)) * 100
    print(f"Converted {len(records)} pairs → {out}")
    print(f"  Skipped: {skipped_empty} too-few-keep, {skipped_all} too-many-keep")
    print(f"  Avg token keep rate: {keep_pct:.1f}%")
    print(f"  Total tokens: {total_tokens}, kept: {total_kept}")


if __name__ == "__main__":
    main()
