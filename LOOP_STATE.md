# Kompress Fine-Tuning Loop — State File

*Updated: 2026-06-25. This file is the spine of the training loop.
The agent forgets. The repo does not. (Addy Osmani)*

## Convergence status

| Version | mk_in_ref | Heretic exact | Override delta | Status |
|---------|-----------|---------------|----------------|--------|
| v2-base | — | 0.975 | — | baseline, keep_rate=0.897 |
| v3 | 0.720 | 0.942 | +0.027 | ✓ shipped |
| v3.1 | ~0.85 | 0.925 | +0.002 | ✓ shipped |
| v3.2 | ~0.85 | 0.929 | +0.002 | ✓ shipped |
| v3.3 | 1.00* | 0.942 | — | domain-only, overfit |
| v4 | 0.823 | 0.967 | 0.000 | ✓ **BREAKTHROUGH** — override redundant |
| v5 | ~0.86 | 0.961 | 0.000 | loop converged, slight regression |

*domain_train.jsonl has mk_in_ref=1.0 by construction

## Open hypotheses

- [ ] Voting ensemble (v3+v4+v5, threshold=2/3) — may beat any single model
- [ ] Domain routing — lower threshold for code/logs, higher for prose
- [ ] Evaluator-optimizer on self-labeling — iterate relabeling until mk_in_ref >= 0.9
- [ ] C3 self-distillation — use real headroom proxy logs as training data

## What the loop has learned

1. Q&A test set is the wrong benchmark (labels are noisy, ceiling at 0.882)
2. Heretic adversarial eval is the right benchmark (dense must-keep tokens)
3. Label quality (mk_in_ref) is the bottleneck, not model capacity
4. One self-labeling iteration (v3→v4) was sufficient to internalize override behavior
5. Second iteration (v4→v5) added noise — convergence criterion met

## Next run decision

**Recommended:** voting ensemble eval (free, CPU) + domain routing experiment ($0.20)

**Skip:** more self-labeling iterations (converged), larger models (not the bottleneck)

## Budget tracking

| Run | Cost | Cumulative |
|-----|------|------------|
| v3 training | $0.09 | $0.09 |
| v3.1 | $0.20 | $0.29 |
| v3.2 | $0.20 | $0.49 |
| v3.3 | $0.15 | $0.64 |
| v4 (self-label+train) | $0.15 | $0.79 |
| v5 (self-label+train) | $0.15 | $0.94 |
| heretic data gen | $0.06 | $1.00 |
| **Total** | | **~$1.00** |

Remaining budget: ~$5.90
