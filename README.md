# ultrawhale

**vaked-base fork of DeepSeek Code Whale.** DeepSeek-native coding agent with content-addressed blocks engine (Go+Asm+GPU), 6 plugins, AG-UI themes, floating widgets, and 7-phase native agent loop.

> Fork maintained at [peterlodri-sec/ultrawhale](https://github.com/peterlodri-sec/ultrawhale). Part of the [vaked-base](https://github.com/peterlodri-sec/vaked-base) monorepo.

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
go build -ldflags="-X github.com/peterlodri-sec/ultrawhale/internal/build.Version=v2.0.0" ./cmd/whale
```


## Contributing

See [docs/internal-blocks.md](docs/internal-blocks.md) for the blocks engine
architecture, performance patterns, benchmarks, and review checklist.

Build: GOOS=darwin GOARCH=arm64 go build ./cmd/whale (macOS)
      GOOS=linux GOARCH=amd64 GOAMD64=v3 go build ./cmd/whale (Linux)
Test:  go test -count=1 -race ./internal/...
Bench: go test -bench=. -benchmem ./internal/blocks/

## License

Apache 2.0 (upstream). Fork maintained by [peterlodri-sec](https://github.com/peterlodri-sec).
