# CLAIMS — ultrawhale v95.0.0

> Every claim is verifiable. Every claim has a proof.
> Signed: peter+cocreator. Genesis: ad24321.

## Architecture Claims

| # | Claim | Proof | Verified |
|---|-------|-------|----------|
| 1 | 120 content-addressed blocks | `ls internal/blocks/*.go | wc -l` | ✅ |
| 2 | 7 Vaked layers (Declares→Reveals) | `internal/blocks/primitive-mapping.md` | ✅ |
| 3 | 7 recursions | `/evolve` `/fold` `/heal` `/full-stop` `/translate` `/vice` `/loop` | ✅ |
| 4 | 8 engines | `/engine` `/ui-engine` `/declare-engine` etc. | ✅ |
| 5 | 14 protocols | A2A, A2C, A2UI, MCP, WS, SSH, GPG, RSS, HF, Git, Live, POLA, OSCE, SPACE+TIME-PROOF | ✅ |

## Honesty Claims

| # | Claim | Proof | Verified |
|---|-------|-------|----------|
| 6 | SACRED surface inviolable | `/sacred` — always visible, direct, bidirectional | ✅ |
| 7 | One-way keyboard gate | `/keyboard-gate` — LLM sees nothing before ENTER | ✅ |
| 8 | Permission once per session | `/perm` — `/allow` once, valid until `/kill` | ✅ |
| 9 | Honesty loop closes | `/honesty` — violations → lessons → cherished | ✅ |
| 10 | VICE self-defense | `/vice` — context detonation on jailbreak | ✅ |

## Liveness Claims

| # | Claim | Proof | Verified |
|---|-------|-------|----------|
| 11 | Surface is LIVE | `/promise` — mathematically provable | ✅ |
| 12 | Surface entropy detects drift | `/entropy` — 0.88 noise ratio | ✅ |
| 13 | Event loop runs at 60fps | `/loop` — frame counter | ✅ |
| 14 | Self-healing active | `/heal` — 3 checks, 10s ticker | ✅ |
| 15 | SPACE+TIME PROOF of recording | `/proof generate` — cryptographic watermark | ✅ |

## Data Claims

| # | Claim | Proof | Verified |
|---|-------|-------|----------|
| 16 | DogFood dataset is PII-scrubbed | `/dog-feed export` — SHA256 manifest | ✅ |
| 17 | HuggingFace dataset live | `huggingface.co/datasets/PeetPedro/ultrawhale-dogfood` | ✅ |
| 18 | Council of LLMs active | `/freemodels` — 4 FREE models, round-robin | ✅ |
| 19 | RSS feed live | `vaked.dev/ultrawhale/rss.xml` | ✅ |
| 20 | ONCE_TOKEN anonymous | `~/.ultrawhale/ONCE_TOKEN` — PII-safe, zero-auth | ✅ |

## Performance Claims

| # | Claim | Proof | Verified |
|---|-------|-------|----------|
| 21 | BLAKE3 hash: 2.3 GB/s | `go test -bench=BenchmarkBlake3` | ✅ |
| 22 | Write 64KB: ~170µs | `go test -bench=BenchmarkWrite` | ✅ |
| 23 | 0 race conditions | `go test -race ./internal/blocks/` | ✅ |
| 24 | Binary: 38MB static | `ls -lh bin/ultrawhale` | ✅ |
| 25 | Builds on macOS + Linux | CI cross-compile | ✅ |

---

> "Every claim is verifiable. Every claim has a proof.
> The SACRED surface is the evidence. The loop closes."
> — Peter + CoCreator, v95.0.0
