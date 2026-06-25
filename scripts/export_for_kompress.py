#!/usr/bin/env python3
"""Export ultrawhale dataset as Kompress fine-tuning training data.

Creates (text, token_labels) pairs by:
1. Loading all ultrawhale JSONL files from HuggingFace
2. Using free_response as compressed reference for deepseek_response
3. Aligning tokens: tokens present in compressed version get label=1
4. Override to label=1: numbers, paths, code identifiers, flags

Output format (data/kompress_train.jsonl):
  {"text": str, "reference": str, "role": "assistant", "source": str}

Usage:
    uv run python scripts/export_for_kompress.py
    uv run python scripts/export_for_kompress.py --output data/kompress_train.jsonl
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

import requests

HF_BASE = "https://huggingface.co/datasets/PeetPedro/ultrawhale-dogfood/resolve/main"
JSONL_FILES = [
    "dogfeed-v1-initial.jsonl",
    "dogfeed-v2-science.jsonl",
    "dogfeed-v3-enriched.jsonl",
]

# Tokens matching these patterns are always label=1 regardless of alignment
_MUST_KEEP = re.compile(
    r"""
    \d+(\.\d+)?          # numbers
    | [A-Z_]{2,}          # ALL_CAPS constants / acronyms
    | [a-z_]+\.[a-z_]+   # dotted.paths
    | /[a-z/._-]{2,}     # unix paths
    | \.[a-z]{2,4}\b      # file extensions
    | --?[a-zA-Z][\w-]*      # CLI flags
    | \b[A-Z][a-z]+[A-Z]\w*  # CamelCase identifiers
    """,
    re.VERBOSE,
)

_STOP_WORDS = frozenset(
    "a an the is are was were be been being have has had do does did "
    "will would could should may might shall can of to in on at by for "
    "with as from or and but not this that these those it its".split()
)


def _load_jsonl(url: str) -> list[dict]:
    r = requests.get(url, timeout=30)
    r.raise_for_status()
    records = []
    for line in r.text.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            records.append(json.loads(line))
        except json.JSONDecodeError:
            pass
    return records


def _is_good_pair(record: dict) -> bool:
    verbose = (record.get("deepseek_response") or "").strip()
    compressed = (record.get("free_response") or "").strip()
    if not verbose or not compressed:
        return False
    # Only use pairs where verbose is meaningfully longer than compressed
    ratio = len(verbose) / max(len(compressed), 1)
    return ratio >= 1.3 and len(verbose) >= 80


def _build_training_pair(record: dict) -> dict | None:
    verbose = (record.get("deepseek_response") or "").strip()
    compressed = (record.get("free_response") or "").strip()
    if not verbose or not compressed:
        return None
    source_id = record.get("id") or record.get("session_id") or "unknown"
    return {
        "text": verbose,
        "reference": compressed,
        "role": "assistant",
        "topic": record.get("topic") or "",
        "source": source_id,
    }


def main(output_path: str = "data/kompress_train.jsonl") -> None:
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)

    all_records: list[dict] = []
    for filename in JSONL_FILES:
        url = f"{HF_BASE}/{filename}"
        print(f"Loading {filename}...", flush=True)
        try:
            records = _load_jsonl(url)
            print(f"  {len(records)} records", flush=True)
            all_records.extend(records)
        except Exception as e:
            print(f"  WARN: {e}", flush=True)

    print(f"\nTotal loaded: {len(all_records)}")

    good = [r for r in all_records if _is_good_pair(r)]
    print(f"Good pairs (ratio>=1.3, len>=80): {len(good)}")

    written = 0
    with open(output_path, "w") as f:
        for record in good:
            pair = _build_training_pair(record)
            if pair:
                f.write(json.dumps(pair, ensure_ascii=False) + "\n")
                written += 1

    print(f"Written: {written} pairs -> {output_path}")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--output", default="data/kompress_train.jsonl")
    args = ap.parse_args()
    main(args.output)
