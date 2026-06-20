# Token Cost & Real Cost — Explanations

## Terminology

| Term | Meaning |
|------|---------|
| **API Tokens** | Actual tokens sent to DeepSeek API (billable) |
| **Folded Tokens** | Tokens used inside ultrawhale (subagent delegation, block operations, brain cycles) |
| **Real Cost** | DeepSeek API cost at current pricing |
| **Folded Cost** | What it would cost if folded tokens were billed at Sonnet 4.7 rates (for comparison) |
| **Cache Hit** | DeepSeek prefix caching — 98% discount on cached prompt prefixes |

## DeepSeek V4 Pricing

| Model | Input (per 1M) | Output (per 1M) |
|-------|---------------|-----------------|
| V4 Flash | $0.14 | $0.28 |
| V4 Pro | $0.55 | $1.10 |
| Cache hit | $0.0028 (98% discount) | — |

## Sonnet 4.7 Comparison (estimated)

| | Input (per 1M) | Output (per 1M) |
|--|---------------|-----------------|
| Sonnet 4.7 | $3.00 | $15.00 |

## Display

The HUD statusline shows:
```
[deepseek-v4-flash] ⎇ main  ● 2:35 · 4821t · $0.0142
                      API: 3200t ($0.0009)  Folded: 1621t (~$0.0243 sonnet47)
```

- **Left**: model + branch + busy timer
- **Tokens**: total API + folded tokens
- **Cost**: actual DeepSeek cost
- **Hover/detail**: API vs folded breakdown

## Benchmarks

Typical session costs (measured on M1 Max):

| Operation | API Tokens | Cost |
|-----------|-----------|------|
| Simple prompt ("what is 2+2") | ~50 | $0.000007 |
| Code exploration (grep + read) | ~500 | $0.00007 |
| Full implementation (build + test + fix) | ~5,000 | $0.0007 |
| Swarm orchestration (10 agents) | ~50,000 | $0.007 |

Folded tokens add ~40% to total token count (subagent overhead, brain cycles).
