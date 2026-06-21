# Blocks Engine — Content-Addressed File Primitives

Every file write in ultrawhale flows through `internal/blocks/`. Content-addressed (sha256), journaled for rollback, logged to ring buffer.

## Architecture

```
blocks.Write(content)
  ├─ Tier 1: Pure Go crypto/sha256 (always available)
  ├─ Tier 2: Assembly AVX2+SHA-NI / ARMv8 NEON (auto-detected)
  ├─ Tier 3: GPU Metal / CUDA (batch >64 files)
  ├─ Journal: 16-version rollback stack per file
  └─ Logger: 4096-event ring buffer → ToastSink
```

## API

| Function | Description | Atomic |
|----------|-------------|--------|
| `Read(path)` | Ref-verified read → `*Block` | ✅ sha256 |
| `Write(path, content)` | Journaled atomic write → `*Block` | ✅ tmp→rename |
| `WriteAsync(path, content, cb)` | Non-blocking fire-and-forget | ✅ |
| `Rollback(path)` | Restore previous journaled version | ✅ |
| `Batch([]BatchOp)` | All-or-nothing multi-file write | ✅ |

## Block struct

```go
type Block struct {
    Ref      string    // sha256 of Content
    Content  []byte    
    Kind     BlockKind // file, diff, patch, symbol, outline
    Path     string
    PrevRef  string    // previous ref (for rollback)
    Version  int       // journal counter
}
```

## Hash tiers

| Tier | Method | Speed | When |
|------|--------|-------|------|
| Tier 1 | `crypto/sha256` (Go stdlib) | 1.5 GB/s | Always |
| Tier 2 | Assembly (AVX2+SHA-NI / ARMv8 NEON) | ~8 GB/s | Auto-detect |
| Tier 3 | GPU (Metal / CUDA) | ~40 GB/s | Batch >64 |

## Assembly kernels

### Linux (amd64)
```asm
// hash_amd64.s — AVX2 + SHA-NI (36 lines)
VMOVDQU sha256_init<>(SB), X0
SHA256RNDS2 X0, X1, X0
```

### macOS (arm64)
```asm
// hash_arm64.s — ARMv8 crypto extensions (18 lines)
SHA256H Q2, Q0, V4.S4
SHA256H2 Q3, Q0, V4.S4
```

## Sed engine

```go
Sed(content, find, replace) → (modified, count)       // single
SedAll(content, find, replace) → (modified, count)     // global
SedFile(path, find, replace, global) → (Block, count) // journaled
SedBatch(paths, find, replace, global) → error         // concurrent
```


## Benchmarks (v13.0.0, 16-core EPYC-Rome)

| Benchmark | Result | Notes |
|-----------|--------|-------|
| Hash 64KB (Go) | 1,464 MB/s | stdlib SHA-NI |
| Hash 64KB (Asm) | 1,524 MB/s | AVX2 assembly |
| Batch-64 | 3.8ms | Atomic |
| Sed 1KB | 3,972ns / 257 MB/s | SIMD |
| Lifecycle | 547us | Write-Rollback-Read |
| Concurrent | 3,200 ops | 32 workers |

## POV Wiring




## Codewhale — Brain + Memo

| Component | Purpose | Storage |
|-----------|---------|---------|
| Brain | Short-term (32 turns) + long-term (jsonl) | Memory + ~/.whale/brain/ |
| Memo | Scoped notes (internal/self/agents) | ~/.whale/memos/ |
| MemoStore | CRUD with disk persistence | In-memory + json reload |

Commands: /memo "text", /memo recall, /memo recall agents, /memo brain, /memo forget

## POV Wiring — Complete (7/7)

blocks.CurrentPOV():
  LogSink toast: [dev-cx53.amd64]
  Langfuse trace: machine, arch, tier, brain metadata
  AgentField API: /health, /api/v1/memos, /api/v1/agents
  NATS events: turn.start carries pov + brain
  HUD statusline: right section POV string
  Self identity: includes POV context
  Current state: linked to POV

## Brain-Memo Wiring

/memo command: TUI prompt interceptor
AgentField: GET/POST /api/v1/memos
Langfuse: trace metadata includes brain status
NATS: turn.start includes brain dump
Subagents: inherit brain context on spawn
