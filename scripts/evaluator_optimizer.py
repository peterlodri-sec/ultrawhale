#!/usr/bin/env python3
"""
Evaluator-Optimizer: diagnose v4 must-keep failures with Qwen2.5-7B teacher.

Phase 1 (diagnosis): Run v4 self-labeling on agent data, compare mk_in_ref
against ground-truth references. For pairs where v4 misses must-keep tokens,
ask Qwen2.5-7B to identify ALL must-keep tokens in the text and explain the
patterns that v4 systematically misses.

Phase 2 (relabeling, future): Use Qwen's output as corrected references for
training v8 — iterate until mk_in_ref >= 0.9 across the dataset.

Usage:
    export HF_INFER_PRO=hf_...
    python3 scripts/evaluator_optimizer.py --model PeetPedro/kompress-v4 --sample 50
    python3 scripts/evaluator_optimizer.py --phase relabel --model PeetPedro/kompress-v4
"""
from __future__ import annotations

import argparse
import json
import os
import random
import re
import sys
import time
from collections import defaultdict
from pathlib import Path

# ── Must-keep regex (same as production kompress_compressor.py) ──────────
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


def load_model(model_id: str):
    import torch
    sys.path.insert(0, str(Path(__file__).parent.parent))
    from scripts.train_kompress import HeadroomCompressorModel, load_v2_weights
    from transformers import AutoTokenizer

    tok = AutoTokenizer.from_pretrained("answerdotai/ModernBERT-base")
    model = HeadroomCompressorModel("answerdotai/ModernBERT-base")
    load_v2_weights(model, model_id)
    model.eval()
    model = model.to("cpu")
    return tok, model, "cpu"


def v4_self_label(text: str, tok, model, device) -> tuple[str, set[str]]:
    """Run v4 inference + must-keep override. Returns (compressed_text, must_keep_matches_in_output)."""
    import torch
    enc = tok(text, return_tensors="pt", truncation=True, max_length=512, padding=True)
    enc = {k: v.to(device) for k, v in enc.items()}
    with torch.no_grad():
        logits, span = model(enc["input_ids"], enc["attention_mask"])
        probs = torch.softmax(logits, dim=-1)[0, :, 1]
        scores = probs * (0.5 + 0.5 * span[0])
        keep = scores > 0.5
    tokens = tok.convert_ids_to_tokens(enc["input_ids"][0])

    # Must-keep override with sliding window (matches production)
    n = len(tokens)
    override_count = 0
    for i in range(n):
        for w in (1, 2, 3):
            if i + w > n:
                break
            window = tok.convert_tokens_to_string(tokens[i : i + w]).strip()
            if _MUST_KEEP_RE.search(window):
                for j in range(i, i + w):
                    if not keep[j]:
                        keep[j] = True
                        override_count += 1
                break

    kept_tokens = [t for t, k in zip(tokens, keep) if k and t not in ("[CLS]", "[SEP]", "<s>", "</s>")]
    compressed = tok.convert_tokens_to_string(kept_tokens)

    # Must-keep patterns that survived in the compressed output
    must_in_output = {m.group() for m in _MUST_KEEP_RE.finditer(compressed)}

    return compressed, must_in_output


def extract_must_keep_from_reference(reference: str) -> set[str]:
    """Extract all must-keep patterns from the ground-truth reference."""
    return {m.group() for m in _MUST_KEEP_RE.finditer(reference)}


def compute_mk_in_ref(kept: set[str], must_matches: set[str]) -> float:
    """What fraction of regex-identified must-keep tokens survived?"""
    if not must_matches:
        return 1.0
    survived = sum(1 for m in must_matches if m in " ".join(sorted(kept)))
    return survived / len(must_matches)


def ask_qwen_to_identify_must_keep(text: str, v4_kept_str: str, reference: str) -> dict:
    """Ask Qwen2.5-7B to identify all must-keep tokens in the text."""
    from huggingface_hub import InferenceClient

    token = os.environ.get("HF_INFER_PRO", "")
    if not token:
        return {"error": "HF_INFER_PRO not set"}

    client = InferenceClient(token=token)

    prompt = f"""You are a token compression evaluator. Below is a tool output that needs compression.
The current compression model (v4) kept these tokens (space-separated):
{v4_kept_str}

The ground-truth reference (what SHOULD be kept) is:
{reference[:500]}

TASK: List ALL tokens/patterns that MUST survive compression — things that carry irreplaceable 
information: compiler flags (-O2, --verbose), hex addresses (0xDEAD), file paths (/var/log), 
env vars (HEADROOM_FOO), error names (TokenExpiredError), version numbers (1.2.3), CamelCase 
identifiers, dotted names (my.module), and extension patterns (.py, .rs).

Output as JSON with these fields:
- "must_keep_tokens": list of token strings that MUST be kept
- "v4_missed": tokens the model SHOULD have kept but didn't
- "pattern_class": what kind of pattern (compiler_flag, path, hex_addr, env_var, etc.)
- "why_v4_fails": brief explanation of why the subword tokenizer causes v4 to miss these

Text to analyze:
{text[:2000]}"""

    try:
        result = client.chat_completion(
            messages=[{"role": "user", "content": prompt}],
            model="Qwen/Qwen2.5-7B-Instruct",
            max_tokens=500,
            temperature=0.1,
        )
        response = result.choices[0].message.content
        # Try to extract JSON
        json_match = re.search(r"\{.*\}", response, re.DOTALL)
        if json_match:
            return json.loads(json_match.group())
        return {"raw_response": response}
    except Exception as e:
        return {"error": str(e)}


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", default="PeetPedro/kompress-v4")
    ap.add_argument("--data", default="data/kompress_agent_train.jsonl")
    ap.add_argument("--sample", type=int, default=20, help="Pairs to analyze (max)")
    ap.add_argument("--mk-threshold", type=float, default=0.9,
                    help="mk_in_ref below this → ask Qwen for analysis")
    ap.add_argument("--output", default="data/evaluator_optimizer_diagnosis.jsonl")
    ap.add_argument("--dry-run", action="store_true", help="Don't call Qwen, just compute v4 stats")
    args = ap.parse_args()

    # Load data
    rows = [json.loads(l) for l in open(args.data)]
    rng = random.Random(42)
    sample = rng.sample(rows, min(args.sample, len(rows)))

    print(f"Loading {args.model}...")
    tok, model, device = load_model(args.model)

    print(f"\n=== Evaluator-Optimizer Diagnosis ===\n")
    print(f"{'#':>3} {'mk_ref':>7} {'v4_kept':>8} {'must_kept':>9} {'needs_teacher':>13}")
    print("-" * 50)

    results = []
    needs_teacher = []

    for idx, row in enumerate(sample):
        text = row["text"]
        reference = row["reference"]

        # Run v4 self-labeling
        v4_compressed, v4_must = v4_self_label(text, tok, model, device)

        # Extract ground-truth must-keep patterns from reference
        ref_must = extract_must_keep_from_reference(reference)

        # Compute mk_in_ref: how many must-keep patterns from the TEXT
        # survived in v4's compressed output?
        text_must = {m.group() for m in _MUST_KEEP_RE.finditer(text)}
        mk_survived = sum(1 for m in text_must if m in v4_compressed)
        mk_in_ref = mk_survived / max(len(text_must), 1)

        needs = "ASK_QWEN" if mk_in_ref < args.mk_threshold else ""
        print(f"{idx+1:>3} {mk_in_ref:>7.3f} {len(v4_compressed):>8} {mk_survived}/{len(text_must):<6} {needs:>13}")

        entry = {
            "idx": idx,
            "source": row.get("source", "unknown"),
            "mk_in_ref": round(mk_in_ref, 4),
            "v4_kept_chars": len(v4_compressed),
            "text_must_count": len(text_must),
            "mk_survived": mk_survived,
            "text": text,
            "reference": reference,
            "v4_compressed": v4_compressed,
        }

        if mk_in_ref < args.mk_threshold:
            needs_teacher.append(entry)

        results.append(entry)

    # Summary
    avg_mk = sum(r["mk_in_ref"] for r in results) / len(results) if results else 0
    print(f"\n--- Summary ---")
    print(f"  Samples analyzed: {len(results)}")
    print(f"  Avg mk_in_ref: {avg_mk:.3f}")
    print(f"  Pairs needing teacher (mk<{args.mk_threshold}): {len(needs_teacher)}")

    if needs_teacher and not args.dry_run:
        print(f"\n=== Qwen2.5-7B Teacher Analysis ({len(needs_teacher)} pairs) ===\n")
        for i, entry in enumerate(needs_teacher[:5]):  # Limit to 5
            print(f"--- Pair {i+1}/{min(len(needs_teacher),5)} (mk_in_ref={entry['mk_in_ref']:.3f}) ---")
            diagnosis = ask_qwen_to_identify_must_keep(
                entry["text"],
                "",  # v4 kept — we'll let Qwen figure it out
                entry["reference"],
            )
            entry["qwen_diagnosis"] = diagnosis
            if "error" in diagnosis:
                print(f"  ERROR: {diagnosis['error']}")
            elif "raw_response" in diagnosis:
                print(f"  RAW: {diagnosis['raw_response'][:200]}")
            else:
                print(f"  must_keep_tokens: {diagnosis.get('must_keep_tokens', [])}")
                print(f"  v4_missed: {diagnosis.get('v4_missed', [])}")
                print(f"  pattern_class: {diagnosis.get('pattern_class', 'unknown')}")
                print(f"  why_v4_fails: {diagnosis.get('why_v4_fails', 'unknown')}")
            time.sleep(1)  # Rate limit
            print()

    # Save results
    out_path = Path(__file__).parent.parent / args.output
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w") as f:
        for r in results:
            f.write(json.dumps(r, ensure_ascii=False) + "\n")
    print(f"Saved {len(results)} diagnosis records to {out_path}")

    # Pattern analysis
    pattern_issues = defaultdict(int)
    for r in needs_teacher:
        source = r["source"]
        pattern_issues[source] += 1
    if pattern_issues:
        print(f"\n=== Failure patterns by source ===")
        for src, count in sorted(pattern_issues.items(), key=lambda x: -x[1]):
            print(f"  {src}: {count}")


if __name__ == "__main__":
    main()
