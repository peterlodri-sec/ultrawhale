---
name: kompress-finetune
description: Fine-tune the Kompress token compression model. Use when training a new version, evaluating compression quality, or running the self-labeling loop. Trigger on: "train kompress", "fine-tune compression", "self-label", "eval heretic", "voting ensemble".
---

# Kompress Fine-Tuning Skill

Fine-tune `chopratejas/kompress-v2-base` (ModernBERT token classifier, 149M params).
Published models: `PeetPedro/kompress-v{3,31,32,33,4,5}`.

## The self-labeling loop (key insight)

Use previous model + hard override to compress training texts → use compressed output as new reference → retrain. One iteration sufficient; second adds noise.

```
v3 (mk_in_ref=0.72, heretic=0.942, override needed)
→ self-label → v4 (mk_in_ref=0.823, heretic=0.967, override=no-op)  ← use this
→ self-label → v5 (mk_in_ref=0.86, heretic=0.961)  ← converged
```

## Benchmarks (critical)

Q&A test set (`kompress_test.jsonl`) has NOISY labels — ceiling ~0.88, ignore plateau.
Heretic eval (`eval_heretic.py`) is ground truth — use this to decide convergence.

Convergence: override_delta = 0 AND heretic exact_pct > 0.96.

## Decision rules (correctable loop invariant)

- Heretic plateaus 3 runs → change approach, not more training
- override_delta = 0 → model internalized behavior, stop self-labeling
- mk_in_ref plateaus 2 iterations → loop converged

## State

See `LOOP_STATE.md` for full progression and budget.
