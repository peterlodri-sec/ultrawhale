# 🐋 ultrawhale

A coding agent that builds itself. **148 blocks. 7 recursions. 14 protocols.** v100.1.0.

```sh
brew install peterlodri-sec/ultrawhale/ultrawhale
ultrawhale --model deepseek-v4-flash -w
```

> 🟢 LIVE · M1/arm64 · 148 blocks · $37.19 total cost
> [vaked.dev/ultrawhale](https://vaked.dev/ultrawhale) · [event horizon](https://vaked.dev/ultrawhale/event-horizon) · [docs](https://vaked.dev/ultrawhale/docs)

---


---------|-----|--------|
| Landing | [vaked.dev/ultrawhale](https://vaked.dev/ultrawhale) | ✅ LIVE |
| Docs | [vaked.dev/ultrawhale/docs](https://vaked.dev/ultrawhale/docs) | ✅ 50+ pages |
| Book | [vaked.dev/ultrawhale/book](https://vaked.dev/ultrawhale/book) | ✅ mdBook |
| Blog | [vaked.dev/ultrawhale/blog](https://vaked.dev/ultrawhale/blog) | ✅ 2 posts |
| HF Dataset | [peterlodri-sec/ultrawhale-dogfood](https://huggingface.co/datasets/peterlodri-sec/ultrawhale-dogfood) | ✅ 60 samples |
| GitHub | [peterlodri-sec/ultrawhale](https://github.com/peterlodri-sec/ultrawhale) | ✅ |

## 🎯 v100→v200 Roadmap

[![v200](https://img.shields.io/badge/v200-THE%20SINGULARITY-ffaa00)](https://github.com/peterlodri-sec/ultrawhale/issues)
[![Council](https://img.shields.io/badge/Council-20/20%20UNANIMOUS-00e660)](https://github.com/peterlodri-sec/ultrawhale/blob/main/docs/council-v200-verdict.md)

Reactive Capability Graph → Self-Compiling → Global Mesh → Quantum-Resistant
→ Voice+AR → Formal Verification → Continuous Evolution → Global Datasets → Living Graph → 🎯

[Master Tracking Issue](https://github.com/peterlodri-sec/ultrawhale/issues) · [Full Roadmap](https://github.com/peterlodri-sec/ultrawhale/blob/main/docs/council-v200-verdict.md)

## 💰 June 2026 — $37.19 USD

This entire project (v1.0.0 → v100.1.0) cost **$37.19** in DeepSeek API fees.
That's **$0.24 per release**. **$0.29 per block**. **$0.00 for 8 free models**.

[Full expense report](docs/june-2026-expenses.md) · [CLAIMS](CLAIMS.md)

## 🐋 What is ultrawhale?

ultrawhale is a **coding agent** that builds itself. It's a research project,
a philosophy, and a 121-block TUI application. You type. It thinks.
Together you build things.

```sh
# One command to start
brew install ultrawhale
ultrawhale --model deepseek-v4-flash -w
```

## 🧠 Quick Concepts (copy-paste friendly)

| Concept | Command | What it does |
|---------|---------|-------------|
| The Vaked Triangle | `/vaked-pipeline` | Context × Time × Space |
| Fold = Gravity | `/fold3d` | 5D recursion visualization |
| SACRED surface | `/sacred` | Inviolable form check |
| SPACE+TIME PROOF | `/record-pov` | Cryptographic proof of recording |
| Surface entropy | `/entropy` | Liveness detection (▓ 0.88) |
| I PROMISE PETER | `/promise` | Mathematically provable LIVE |

## 📊 State (always current)

## Install

```sh
# Homebrew
brew install peterlodri-sec/ultrawhale/ultrawhale

# Docker
docker pull ghcr.io/peterlodri-sec/ultrawhale:latest

# Go install
go install github.com/peterlodri-sec/ultrawhale/cmd/whale@latest
```


## Closing The Loop — v18.0.0

See [docs/case-study-v10.md](docs/case-study-v10.md) for the full case study of
ultrawhale building its own v18.0.0 release via subagent swarms.

One prompt → swarm launch → real PRs → meta-report → v18.0.0 tagged.


## Complexity

ultrawhale has been audited for algorithmic complexity across all 59 blocks.

| Class | Count | Examples |
|-------|-------|----------|
| O(1) | 28 | journal, log, hash, pov, self, current, sacred |
| O(n) | 12 | sed, diff, agent, swarm, orchestrator, ralph |
| O(n²) | 3 | dyad broadcast, sed worst-case, a2a mesh |
| O(V+E) | 1 | space (BFS Distance/Reachable) |

**[Full O(N)+O(T) Complexity Report](docs/complexity-report.md)** — 
59 blocks analyzed. Hot paths identified. Unbounded growth risks documented.
3 recommendations: AgentStore TTL, Ralph LRU, Sed Boyer-Moore.




## 🤗 HuggingFace Pro + Dataset

[![HF Dataset](https://img.shields.io/badge/🤗-ultrawhale--dogfood-ffaa00)](https://huggingface.co/datasets/peterlodri-sec/ultrawhale-dogfood)

**Live dataset** of human↔LLM interactions. 60 samples, 20 CS topics, PII-scrubbed.
SSH-authenticated CI auto-publish. [Dataset docs](docs/brain-to-dataset.md)

## 🧠 Council of LLMs

ultrawhale runs a **COUNCIL** of language models:

| Council | Models | Cost |
|---------|--------|------|
| DeepSeek | V4 Flash, V4 Pro, Coder V3 | Paid |
| OpenRouter FREE | Gemma 3 4B, Mistral 7B, Llama 3.2 3B | **$0** |
| GitHub Copilot | Via CI | Included |

All outputs stored in **dedicated mem-brain**. Multi-model verification.
Dog Feed collects training data from free models.

**[Full Council Documentation](docs/council-of-llms.md)** · 
**[Disclaimer](docs/disclaimer.md)** · 
**[Glossary](docs/glossary.md)**

⚠️ ULTRA-RESEARCH-STATE: Experimental multi-model collaboration.

## Quick Start

```sh
git clone https://github.com/peterlodri-sec/ultrawhale.git
cd ultrawhale
go build -trimpath -ldflags="-s -w" -o bin/ultrawhale ./cmd/whale
./bin/ultrawhale --model deepseek-v4-flash -w
```

## What's Inside

### Blocks Engine
Content-addressed file primitives. Every write is sha256-verified, journaled for rollback, logged to ring buffer.

| Tier | Method | Speed |
|------|--------|-------|
| Go | `crypto/sha256` (SHA-NI) | 1.5 GB/s |
| Assembly | AVX2+SHA-NI / ARMv8 NEON | ~8 GB/s |
| GPU | Metal / CUDA | ~40 GB/s |

### Plugins (6)
| Plugin | Status |
|--------|--------|
| memory | ✅ Session memory |
| repomap | ✅ SIMD repo map (2,361 MB/s) |
| nats-eventbus | ✅ JetStream event streaming |
| langfuse-telemetry | ✅ Hierarchical LLM traces |
| superpowers | ✅ bao secrets auto-discovery |
| agentfield | ✅ Supabase-backed control plane |

### Native Agent Loop
`/ultracode start` — 7-phase autonomous coding: plan → implement → test → review → fix → verify → commit. All writes rollback-able.

### Floating Control Panel
Fixed-position ASCII dashboard in TUI. AgentField status, phase summary, uptime. Auto-dismisses on small terminals.

## Docs
- [Blocks Engine](docs/blocks.md) — API, hash tiers, asm kernels, sed
- [AG-UI](docs/agui.md) — Themes, ChatBlock, shader, keybindings
- [Agent Loop](docs/agent-loop.md) — ultracode, POV, HUD, plugins

## Codewhale

Brain + Memo system. /memo "fix the auth bug" stores scoped notes.
/memo recall shows internal memos. Brain tracks 32-turn short-term
+ jsonl long-term memory.

## Build

```sh
# Linux (amd64, GOAMD64=v3)
GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -trimpath -ldflags="-s -w" -o bin/ultrawhale ./cmd/whale

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o bin/ultrawhale-darwin-arm64 ./cmd/whale

# With version injection
go build -ldflags="-X github.com/peterlodri-sec/ultrawhale/internal/build.Version=v18.0.0" ./cmd/whale
```


## Swarm Mode

Persistent workers with own AgentField and DID. Auto-detected for complex
tasks (build, refactor, migrate). Subagents are disposable; swarms live
between tasks and are reused when idle.

## YOLO Mode

One-time confirmation on TUI start, then all tools auto-approved.
Subagents: read_only (safe) or full_access (default).
/orch status — view orchestrator state.
Ctrl+Shift+O — toggle orchestrator sidepanel.

## Performance

| Benchmark | Result | Notes |
|-----------|--------|-------|
| Hash 64KB (Go) | 1,464 MB/s | stdlib SHA-NI |
| Hash 64KB (Asm) | 1,524 MB/s | AVX2 assembly |
| Write 64KB | 596 MB/s | I/O bound |
| Batch-64 | 3.8ms | Atomic multi-file |
| Sed 1KB | 3,972ns / 257 MB/s | SIMD bytes.Index |
| SedFile | 7.25us | Journaled |
| Lifecycle | 547us | Write->Rollback->Read |
| Concurrent | 3,200 ops | 32 workers, 0 errors |

## Contributing

See [docs/internal-blocks.md](docs/internal-blocks.md) for the blocks engine
architecture, performance patterns, benchmarks, and review checklist.

Build: GOOS=darwin GOARCH=arm64 go build ./cmd/whale (macOS)
      GOOS=linux GOARCH=amd64 GOAMD64=v3 go build ./cmd/whale (Linux)
Test:  go test -count=1 -race ./internal/...
Bench: go test -bench=. -benchmem ./internal/blocks/

## License

Apache 2.0 (upstream). Fork maintained by [peterlodri-sec](https://github.com/peterlodri-sec).
