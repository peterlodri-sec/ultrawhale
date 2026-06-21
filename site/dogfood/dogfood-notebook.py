#!/usr/bin/env python3
"""
ultrawhale DogFood — Quick Start Notebook
==========================================
Load, explore, and prepare the DogFood dataset for fine-tuning.

Usage:
    python dogfood-notebook.py dogfeed-20260621.jsonl
"""

import json
import sys
from collections import Counter

def load_dataset(path):
    """Load DogFood JSONL dataset."""
    samples = []
    with open(path) as f:
        for line in f:
            samples.append(json.loads(line.strip()))
    print(f"Loaded {len(samples)} samples from {path}")
    return samples

def stats(samples):
    """Print dataset statistics."""
    models = Counter(s["free_model"] for s in samples)
    dates = Counter(s["timestamp"][:10] for s in samples)
    print(f"\n=== Dataset Stats ===")
    print(f"Samples: {len(samples)}")
    print(f"Models: {dict(models)}")
    print(f"Dates: {dict(dates)}")
    avg_user_len = sum(len(s["user_message"]) for s in samples) / len(samples)
    avg_free_len = sum(len(s["free_response"]) for s in samples) / len(samples)
    print(f"Avg user msg length: {avg_user_len:.0f} chars")
    print(f"Avg free response length: {avg_free_len:.0f} chars")

def to_finetune_format(samples):
    """Convert to HuggingFace chat format."""
    return [{"messages": [
        {"role": "user", "content": s["user_message"]},
        {"role": "assistant", "content": s["free_response"]}
    ]} for s in samples]

if __name__ == "__main__":
    path = sys.argv[1] if len(sys.argv) > 1 else "dogfeed-20260621.jsonl"
    samples = load_dataset(path)
    stats(samples)
    
    # Convert to fine-tuning format
    ft_data = to_finetune_format(samples)
    out = path.replace(".jsonl", "-finetune.json")
    with open(out, "w") as f:
        json.dump(ft_data, f, indent=2)
    print(f"\nFine-tuning data saved to {out}")
    print(f"Ready for: transformers.Trainer, llama.cpp, unsloth, etc.")
