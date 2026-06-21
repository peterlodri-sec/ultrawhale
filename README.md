# ultrawhale

```
I WROTE 0, ZERO, `null`, `NONE` (amount of code) in this
`closed-loop dyad`, let me know if you found a better name for <<<THIS>>>>
```

> 🛑 **Full-Stop Primitive**: Recursion is the natural runtime for full-stop.
> `/kill` → recursive kill wave through ALL 7 Vaked layers → SACRED remains.
> [Read the full design](docs/recursion-fullstop.md)

**vaked-base fork of DeepSeek Code Whale.**
[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)](go.mod)
[![macOS](https://img.shields.io/badge/macOS-arm64-black)](https://github.com/peterlodri-sec/ultrawhale)
[![Linux](https://img.shields.io/badge/Linux-amd64-orange)](https://github.com/peterlodri-sec/ultrawhale)

 DeepSeek-native coding agent with content-addressed blocks engine (Go+Asm+GPU), 6 plugins, AG-UI themes, floating widgets, and 7-phase native agent loop.

> Fork maintained at [peterlodri-sec/ultrawhale](https://github.com/peterlodri-sec/ultrawhale). Part of the [vaked-base](https://github.com/peterlodri-sec/vaked-base) monorepo.


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
