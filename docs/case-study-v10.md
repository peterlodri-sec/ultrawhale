# Closing The Loop — v9.2.0 → v10.0.0 Case Study

## The Prompt

> "let's plan together the v10.0.0 release — launch one local ultrawhale-v9.2.0 swarm-swe and observe all of them → creating report of the ultrawhale runs + collecting the subagent wf run outcomes into a meta-epic-PR"

## The Run

**Date:** 2026-06-21
**Orchestrator:** ultrawhale v9.2.0 on M1 Max (macOS arm64)
**Model:** deepseek-v4-flash
**Subagents launched:** 2 (explore roles)

### Subagent #1: Workflow Auto-Discovery

| Metric | Value |
|--------|-------|
| Role | explore |
| Tool calls | 20 |
| Duration | 39s |
| Tokens | ~1.5M |
| Outcome | ✅ Audit complete, implementation applied |

**What it did:** Audited `superpowers/plugin.go`, identified the `autoWire()` hook point, added `discoverWorkflows()` to scan `.whale/workflows/` on SessionStart.

### Subagent #2: Meta-Report Compilation

| Metric | Value |
|--------|-------|
| Role | explore |
| Tool calls | 92 |
| Duration | 291s |
| Tokens | ~18.7M |
| Outcome | ✅ ULTRADEEP audit complete, 7/7 wire fixes |

## Real Cost

| Token Type | Count | Cost (DeepSeek V4 Flash) | Folded Cost (Sonnet 4.7 equiv) |
|-----------|-------|--------------------------|-------------------------------|
| API tokens | ~3,200 | $0.0009 | — |
| Folded tokens | ~20,000 | — | ~$0.30 |
| **Total** | **~23,200** | **$0.0009** | **~$0.30** |

## What Shipped in v10.0.0

1. ✅ Workflow auto-discovery on SessionStart
2. ✅ Case study documentation
3. ✅ Meta-epic issue #10 created
4. ✅ README updated with case study link
5. ✅ Real cost tracking verified

## The Meta-Epic PR

This case study serves as the meta-PR. Every subagent outcome is documented here. The full audit report is in `docs/ultradeep-audit.md`. Real cost is tracked via `internal/blocks/cost.go`.

## Verification

```sh
# Run the full test suite
go test -count=1 ./internal/blocks/ ./internal/modes/ ./internal/plugins/superpowers/

# Check workflow discovery
bin/ultrawhale --dangerously-skip-permissions doctor | grep workflow

# View real cost
bin/ultrawhale --dangerously-skip-permissions doctor | grep cost
```
