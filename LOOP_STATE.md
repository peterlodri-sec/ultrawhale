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
| v6 | 1.000* | 0.962 | 0.000 | agent-distribution — keep_rate↑ 0.854, heretic holds > 0.96 |
| v7 | ~0.87* | 0.949 / 0.956+ov | +0.007 | sliding-window fix — override returned, SSL regressed (0.789→0.684) |

*domain_train.jsonl and kompress_agent_train.jsonl both have mk_in_ref=1.0 by construction
**self-labeling skipped for agent data: v4 subword tokenizer drops paths/CamelCase/flags (fixed in v7)

## Open hypotheses

- [x] Voting ensemble — NEGATIVE: v3 noisy votes degrade ensemble to 0.931 vs v4 alone 0.961
- [x] v6 proxy eval — Mode A: v4=9.5% avg compression, v6=4.2% (more conservative keep_rate↑)
- [x] Sliding-window self-labeling fix — test: TokenExpiredError, /var/log, --verbose all pass with 3-token window (run_training_v7.sh)
- [ ] Domain routing — lower threshold for code/logs, higher for prose
- [ ] Evaluator-optimizer on self-labeling — iterate relabeling until mk_in_ref >= 0.9
- [ ] C3 self-distillation — use real headroom proxy logs as training data

## What the loop has learned

1. Q&A test set is the wrong benchmark (labels are noisy, ceiling at 0.882)
2. Heretic adversarial eval is the right benchmark (dense must-keep tokens)
3. Label quality (mk_in_ref) is the bottleneck, not model capacity
4. One self-labeling iteration (v3→v4) was sufficient to internalize override behavior
5. Second iteration (v4→v5) added noise — convergence criterion met
6. Agent-pattern training (v6) increases keep_rate (0.823→0.854): model more conservative, less aggressive compression
7. Self-labeling degrades on agent data: single-subtoken check misses CamelCase/paths/flags (mk_in_ref 0.652)
8. Sliding-window fix (1/2/3-subtoken windows): TokenExpiredError, /var/log/app.log, --verbose all force-kept correctly
9. Agent-distribution direction is a dead end: each iteration (v6, v7) increases keep_rate and decreases heretic precision. More agent training → more conservative → worse adversarial accuracy. Stop here, pivot to domain routing.

## Next run decision

**STOP fine-tuning agent-distribution direction.** v6 and v7 both increased keep_rate and reduced heretic precision. The trend is clear: more agent data → more conservative → worse adversarial accuracy.

**v4 remains production recommendation.** heretic 0.967, override_delta=0, keep_rate=0.823.

**Next (free, no GPU):** domain routing — per-domain thresholds in content_router.py. Lower threshold for code/logs (already dense must-keep), higher for prose. No training needed.

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
| v6 (agent-dist train) | $0.20 | $1.20 |
| **Total** | | **~$1.20** |

Remaining budget: ~$5.70

---

## Session close — 2026-06-25

**PRs opened today:**
- headroom #1400 — kompress must-keep override (approved by JerrettDavis, governance pass)
- headroom #1403 — RTK scope-mixing regression tests (documents the savings underreporting bug)
- headroom #1408 — knowledge-worker memory audit integration proposal

**RTK bug found:** `b70fccbe` changed default scope global → project avg_savings_pct (18.5%) underreports session savings (62.6%). Fix PR incoming from Discord.

**knowledge-worker:** provenance-backed memory audit layer proposed for headroom learn subsystem. Integration surface documented in #1408.

**Budget used today:** ~$1.00 of $7.92 available. Remaining: ~$6.92.

**Next session should read:**
1. This file
2. `.agents/skills/kompress-finetune/SKILL.md` — training decision rules
3. `headroom PR #1400` — approved, merge pending
4. `LOOP_STATE.md` open hypotheses — domain routing (free), C3 self-distillation (needs headroom logging mode)
