#!/usr/bin/env python3
"""Multi-dataset export for Kompress fine-tuning.

Sources:
  1. ultrawhale (deepseek_response vs free_response)
  2. Open-Orca/OpenOrca  (system+question+GPT-4 response — filter by length)
  3. vicgalle/alpaca-gpt4  (instruction + long output)
  4. Wikipedia summaries  (first_paragraph as reference, intro section as verbose)

All produce (text, reference) pairs where reference is the shorter version
of the same information. That's the silver label signal.

Usage:
    python export_for_kompress_multi.py --output data/kompress_multi_train.jsonl --max 3000
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

import requests

HF_BASE = "https://huggingface.co/datasets/PeetPedro/ultrawhale-dogfood/resolve/main"
ULTRAWHALE_FILES = [
    "dogfeed-v1-initial.jsonl",
    "dogfeed-v2-science.jsonl",
    "dogfeed-v3-enriched.jsonl",
]


def _good_pair(verbose: str, compressed: str, min_ratio: float = 1.25) -> bool:
    if not verbose or not compressed:
        return False
    if len(verbose) < 80 or len(compressed) < 30:
        return False
    ratio = len(verbose) / max(len(compressed), 1)
    return min_ratio <= ratio <= 8.0   # cap at 8x to avoid useless extremes


# ── Source 1: ultrawhale ───────────────────────────────────────────────────


def load_ultrawhale() -> list[dict]:
    pairs = []
    for fname in ULTRAWHALE_FILES:
        try:
            r = requests.get(f"{HF_BASE}/{fname}", timeout=30)
            r.raise_for_status()
            for line in r.text.splitlines():
                if not line.strip():
                    continue
                d = json.loads(line)
                verbose = (d.get("deepseek_response") or "").strip()
                compressed = (d.get("free_response") or "").strip()
                if _good_pair(verbose, compressed):
                    pairs.append({
                        "text": verbose,
                        "reference": compressed,
                        "role": "assistant",
                        "source": "ultrawhale",
                        "topic": d.get("topic") or "",
                    })
        except Exception as e:
            print(f"  ultrawhale/{fname} failed: {e}", file=sys.stderr)
    print(f"  ultrawhale: {len(pairs)} pairs")
    return pairs


# ── Source 2: OpenOrca (subset) ───────────────────────────────────────────


def load_openorca(limit: int = 800) -> list[dict]:
    """GPT-4 responses to complex questions tend to be verbose.
    We keep pairs where response > 200 chars and system prompt is short (< 200).
    Use the system+question as context, gpt4_response as verbose,
    synthesize a compressed version as the first ~30% of sentences.
    """
    pairs = []
    try:
        from datasets import load_dataset
        ds = load_dataset("Open-Orca/OpenOrca", split="train", streaming=True)
        count = 0
        for item in ds:
            if count >= limit:
                break
            response = (item.get("response") or "").strip()
            system = (item.get("system_prompt") or "").strip()
            question = (item.get("question") or "").strip()
            if not response or len(response) < 200:
                continue
            # Synthetic reference: first 40% of words (crude but directionally right)
            words = response.split()
            ref_words = max(20, len(words) * 2 // 5)
            reference = " ".join(words[:ref_words])
            if not _good_pair(response, reference):
                continue
            pairs.append({
                "text": response,
                "reference": reference,
                "role": "assistant",
                "source": "openorca",
                "topic": question[:80],
            })
            count += 1
    except Exception as e:
        print(f"  openorca failed: {e}", file=sys.stderr)
    print(f"  openorca: {len(pairs)} pairs")
    return pairs


# ── Source 3: Alpaca-GPT4 ─────────────────────────────────────────────────


def load_alpaca_gpt4(limit: int = 600) -> list[dict]:
    """GPT-4 responses to Alpaca instructions. Verbose responses vs trimmed."""
    pairs = []
    try:
        from datasets import load_dataset
        ds = load_dataset("vicgalle/alpaca-gpt4", split="train", streaming=True)
        count = 0
        for item in ds:
            if count >= limit:
                break
            output = (item.get("output") or "").strip()
            if len(output) < 150:
                continue
            # Reference: first 35% of content
            words = output.split()
            ref = " ".join(words[:max(15, len(words) * 35 // 100)])
            if not _good_pair(output, ref):
                continue
            pairs.append({
                "text": output,
                "reference": ref,
                "role": "assistant",
                "source": "alpaca_gpt4",
                "topic": (item.get("instruction") or "")[:80],
            })
            count += 1
    except Exception as e:
        print(f"  alpaca_gpt4 failed: {e}", file=sys.stderr)
    print(f"  alpaca_gpt4: {len(pairs)} pairs")
    return pairs


# ── Source 4: Dolly-15k ───────────────────────────────────────────────────


def load_dolly(limit: int = 600) -> list[dict]:
    """databricks/databricks-dolly-15k — human-written responses."""
    pairs = []
    try:
        from datasets import load_dataset
        ds = load_dataset("databricks/databricks-dolly-15k", split="train", streaming=True)
        count = 0
        for item in ds:
            if count >= limit:
                break
            response = (item.get("response") or "").strip()
            if len(response) < 200:
                continue
            words = response.split()
            ref = " ".join(words[:max(15, len(words) * 35 // 100)])
            if not _good_pair(response, ref):
                continue
            pairs.append({
                "text": response,
                "reference": ref,
                "role": "assistant",
                "source": "dolly",
                "topic": (item.get("instruction") or "")[:80],
            })
            count += 1
    except Exception as e:
        print(f"  dolly failed: {e}", file=sys.stderr)
    print(f"  dolly: {len(pairs)} pairs")
    return pairs


# ── Source 5: Technical short texts (from headroom-relevant domains) ──────


_TECH_TEXTS = [
    # Shell output / NixOS patterns (from our actual work)
    (
        "error: The option `sops' does not exist. Definition values:\n"
        "- In `/nix/store/.../hosts/public-services-host/sops.nix':\n"
        "    {\n      age = {\n        generateKey = false;\n        keyFile = \"/var/lib/sops-nix/key.txt\";\n      };\n    ...\n"
        "Did you mean `jobs', `boot' or `fonts'?\n"
        "Command 'nix --extra-experimental-features nix-command flakes build ...' returned non-zero exit status 1.",
        "error: option `sops' not found; likely sops-nix module not imported."
    ),
    (
        "when calling the 'seq' builtin\nat lib/modules.nix:402\nwhile evaluating a branch condition\n"
        "at lib/modules.nix:305\nerror: Path 'hosts/public-services-host/sops.nix' does not exist in Git repository",
        "error: sops.nix not tracked in git — run git add hosts/public-services-host/sops.nix"
    ),
    (
        "AttributeError: type object 'HeadroomCallback' has no attribute 'async_post_call_success_hook'\n"
        "Traceback (most recent call last):\n  File \"/app/litellm/proxy/utils.py\", line 2398, in post_call_success_hook\n"
        "    raise e\n  ...\n  callback_response = await callback.async_post_call_success_hook(...)\n"
        "    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n"
        "AttributeError: type object 'HeadroomCallback' has no attribute 'async_post_call_success_hook'",
        "HeadroomCallback missing async_post_call_success_hook — inherit from CustomLogger to get no-op defaults."
    ),
]


def load_tech_samples() -> list[dict]:
    pairs = []
    for verbose, ref in _TECH_TEXTS:
        if _good_pair(verbose, ref):
            pairs.append({
                "text": verbose,
                "reference": ref,
                "role": "tool",
                "source": "tech_handcrafted",
                "topic": "technical",
            })
    print(f"  tech_handcrafted: {len(pairs)} pairs")
    return pairs


# ── Main ──────────────────────────────────────────────────────────────────


def main(output: str = "data/kompress_multi_train.jsonl", max_total: int = 3000) -> None:
    Path(output).parent.mkdir(parents=True, exist_ok=True)

    print("Loading datasets...")
    all_pairs: list[dict] = []
    all_pairs.extend(load_ultrawhale())
    all_pairs.extend(load_openorca(limit=800))
    all_pairs.extend(load_alpaca_gpt4(limit=600))
    all_pairs.extend(load_dolly(limit=600))
    all_pairs.extend(load_tech_samples())

    import random
    random.shuffle(all_pairs)
    all_pairs = all_pairs[:max_total]

    with open(output, "w") as f:
        for p in all_pairs:
            f.write(json.dumps(p, ensure_ascii=False) + "\n")

    sources: dict[str, int] = {}
    for p in all_pairs:
        sources[p["source"]] = sources.get(p["source"], 0) + 1

    print(f"\nTotal: {len(all_pairs)} pairs -> {output}")
    for src, n in sorted(sources.items(), key=lambda x: -x[1]):
        print(f"  {src}: {n}")


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--output", default="data/kompress_multi_train.jsonl")
    ap.add_argument("--max", type=int, default=3000, dest="max_total")
    args = ap.parse_args()
    main(args.output, args.max_total)
