# Kompress Fine-Tuning Loop — State File

*Updated: 2026-06-25 (session 2). This file is the spine of the training loop.
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
- [x] Domain routing — IMPLEMENTED: BUILD_OUTPUT/SOURCE_CODE bias=0.50 (2x compression), SEARCH bias=0.70 (1.45x). headroom PR #1418 (OPEN, ready to merge — no merge perms)
- [x] Must-keep override in production — IMPLEMENTED in headroom kompress_compressor.py (session 2). Sliding-window regex safety net catches compiler flags, hex addrs, paths, env vars, CamelCase. All existing tests pass. Achieves 96.2% mk_in_ref on 50-sample agent data eval.
- [x] Evaluator-optimizer diagnosis — COMPLETED (session 2). Qwen2.5-7B teacher identifies v4 failure patterns: multi-dot versions (3.28.33), = -separated values (timeout=123), rare error names. script: evaluator_optimizer.py
- [ ] C3 self-distillation — use real headroom proxy logs as training data **(NEXT)**

## Benchmark status

**32-prompt expanded heretic (Qwen2.5-7B generated responses, 2026-06-25):**

| Version | keep_rate | exact_pct (32) | exact_pct (8) | override_delta |
|---|---|---|---|---|
| v4 | 0.854 | 0.943 | 0.967 | 0.000 |
| v6 | 0.746 | 0.942 | 0.962 | 0.000 |
| v7 | 0.782 | 0.944 | 0.956 | +0.002 |

**Standard eval going forward:** `python3 scripts/eval_heretic.py --model X --prompts-file data/heretic_expanded.jsonl`
Target: exact_pct > 0.940, override_delta = 0.000

## What the loop has learned

1. Q&A test set is the wrong benchmark (labels are noisy, ceiling at 0.882)
2. Heretic adversarial eval is the right benchmark (dense must-keep tokens)
3. Label quality (mk_in_ref) is the bottleneck, not model capacity
4. One self-labeling iteration (v3→v4) was sufficient to internalize override behavior
5. Second iteration (v4→v5) added noise — convergence criterion met
6. Agent-pattern training (v6) increases keep_rate (0.823→0.854): model more conservative
7. Self-labeling degrades on agent data: single-subtoken check misses CamelCase/paths/flags (mk_in_ref 0.652)
8. Sliding-window fix works mechanically but produces more conservative model (v7)
9. Agent-distribution direction is a dead end. More agent training → more conservative → worse adversarial accuracy
10. **Session 2 discovery**: Must-keep override in production (not training) achieves 96.2% mk_in_ref on agent data with v4. Post-inference regex safety net is the right approach — don't train the model to do what regex can do perfectly.
11. Qwen2.5-7B teacher confirms remaining gaps are edge cases: multi-dot versions, = -separated values

## Next run decision

**STOP fine-tuning agent-distribution direction.** v6 and v7 both increased keep_rate and reduced heretic precision.

**v4 + must-keep override = production recommendation.** heretic 0.967, override_delta=0, keep_rate=0.823, agent mk_in_ref=0.962 (with override).

**Next:** C3 self-distillation — collect real headroom proxy logs with log_full_messages=true, label with Qwen2.5-7B, train v8 on real traffic.

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

## Session 2 — 2026-06-25

### Actions completed

1. **LOOP_STATE.md refreshed** with latest status
2. **Domain routing PR #1418** — attempted merge (no perms on headroomlabs-ai/headroom). PR is MERGEABLE with all checks passing. Needs repo owner to merge.
3. **Must-keep override in production** (`headroom/headroom/transforms/kompress_compressor.py`):
   - Added `_MUST_KEEP_RE` regex (compiler flags, hex addrs, paths, env vars, CamelCase, dot-names, extensions)
   - Added sliding-window (1-3 words) safety net after ML inference
   - Added `KompressConfig.enable_must_keep_override: bool = True` (default on)
   - All 81 existing tests pass (28 kompress + 53 content_router)
   - Verified: 11/11 must-keep patterns survive on a bash_output example where model previously dropped `Cargo.toml` → `C argo . tom l`
4. **Evaluator-optimizer diagnosis** (`scripts/evaluator_optimizer.py`):
   - Self-labeling eval of v4 + must-keep override on 50 agent data samples
   - **Result: 96.2% avg mk_in_ref** (vs 0.652 without override)
   - Qwen2.5-7B teacher analyzed 4 failing pairs: identified multi-dot versions, = -separated values, and rare error names as remaining gaps
   - Data saved: `data/evaluator_optimizer_diagnosis.jsonl`
5. **HF_INFER_PRO validated** — Qwen2.5-7B-Instruct works via huggingface_hub InferenceClient

### Key finding

The must-keep override in production code (post-inference regex safety net) is the CORRECT approach for compiler flags/paths/hex/enums. Training the model to handle these (v7 sliding-window training) made it more conservative AND regressed heretic precision. Regex in production gives us the precision of v4 (0.967 heretic) PLUS the coverage of v7 (agent must-keep patterns).

### Headroom changes needing merge

File: `headroom/headroom/transforms/kompress_compressor.py`
Changes:
- `import re` added
- `_MUST_KEEP_RE` module-level regex pattern
- `KompressConfig.enable_must_keep_override: bool = True`
- Sliding-window must-keep safety net in `compress()` method

---

## Session close — 2026-06-25 (session 1)

**PRs opened today:**
- headroom #1400 — kompress must-keep override (approved by JerrettDavis, governance pass)
- headroom #1403 — RTK scope-mixing regression tests
- headroom #1408 — knowledge-worker memory audit integration proposal

**Budget used:** ~$1.00 of $7.92. Remaining: ~$6.92.
