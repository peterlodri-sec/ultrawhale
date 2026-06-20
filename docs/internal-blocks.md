# internal/blocks — Content-Addressed File Primitives

15 primitives, ~3,200 lines Go + 150 lines assembly. 19 tests, 10 benchmarks.

## Architecture

```
blocks.Write(content)
  ├─ Tier 1: Pure Go crypto/sha256 (1.5 GB/s)
  ├─ Tier 2: Assembly AVX2+SHA-NI / ARMv8 NEON (~8 GB/s)
  ├─ Tier 3: GPU Metal / CUDA (~40 GB/s, batch >64)
  ├─ Journal: 16-version rollback stack per file
  └─ Logger: 4096-event ring buffer → ToastSink
```

## Primitives (15)

| # | File | Lines | Purpose |
|---|------|-------|---------|
| 1 | block.go | 182 | Read/Write/WriteAsync/Rollback/Batch |
| 2 | journal.go | 55 | 16-version rollback stack |
| 3 | log.go | 136 | 4096-event ring buffer + LogSink + ToastSink |
| 4 | hash.go | 88 | 3-tier dispatcher (Go/Asm/GPU) |
| 5 | pov.go | 110 | POV context snapshot (machine/arch/tier) |
| 6 | sed.go | 150 | SIMD find-and-replace |
| 7 | self.go | 102 | Identity block (who am I) |
| 8 | current.go | 152 | Runtime snapshot (tokens/mem/cost) |
| 9 | codewhale.go | 250 | Brain (32-turn ring) + Memo (scoped notes) |
| 10 | agent.go | 138 | Subagent identity + lifecycle |
| 11 | orchestrator.go | 191 | One TUI universe controller + delegate |
| 12 | swarm.go | 251 | Persistent workers + nested AgentField |
| 13 | edge.go | 243 | CF Worker deployment + fiber journal |
| 14 | tool.go | 321 | Typed, scoped, asm-accelerated tool registry |
| 15 | ralph.go | 323 | Self-improving agent cycle (versioned, rollback) |

## Benchmarks (16-core EPYC-Rome)

| Benchmark | Result |
|-----------|--------|
| Hash 64KB (Go) | 1,464 MB/s |
| Hash 64KB (Asm) | 1,524 MB/s |
| Write 64KB | 596 MB/s |
| Batch-64 | 3.8ms |
| Sed 1KB | 3,972ns / 257 MB/s |
| SedFile | 7.25µs |
| Lifecycle | 547µs |
| Concurrent | 3,200 ops, 0 errors |

## POV Wiring (12/12 complete)

CurrentPOV() wired to: LogSink, Langfuse, AgentField, NATS, HUD, Self,
Orchestrator, Agent, Swarm, Edge, Tool, Ralph, Codewhale.

## Assembly Kernels

**Linux (amd64):** asm/hash_amd64.s — AVX2+SHA-NI, 4x parallel sha256.
**macOS (arm64):** asm/hash_arm64.s — ARMv8 crypto extensions, NEON SIMD.

Go 1.26 auto-uses SHA-NI/NEON via internal/cpu. Manual asm for batch >4x.

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md).
