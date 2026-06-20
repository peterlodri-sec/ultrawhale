# ultrawhale × Vaked — Philosophy Alignment

## Vaked Mantra

> Vaked declares. Nix materializes. OTP supervises. Zig enforces.
> eBPF testifies. CrabCC indexes. Surfaces reveal.

## ultrawhale Mapping

| Vaked Layer | ultrawhale Equivalent | Status |
|-------------|----------------------|--------|
| **Declares** | 29 content-addressed blocks | ✅ Complete |
| **Materializes** | build:* orchestrator tools, Taskfile, Nix shell | ✅ Complete |
| **Supervises** | Orchestrator + Ralph Loop | ✅ Complete |
| **Enforces** | 7 pre-hooks + blocks.Write (journaled) | ✅ Complete |
| **Testifies** | blocks.Log (journaled events) → Langfuse traces | ✅ Complete |
| **Indexes** | Repomap SIMD + tool registry + BLAKE3 | ✅ Complete |
| **Reveals** | AG-UI + 6 widgets + TUI + InfraBar | ✅ Complete |

## Gap Analysis

| Gap | Vaked Layer | Suggested Primitive |
|-----|-------------|-------------------|
| No formal schema for blocks | Declares | `schema` block — validates block structure against schema |
| No Nix materialization from blocks | Materialize | `nix` block — generate flake.nix from capability graph |
| No OTP-like supervision tree | Supervise | `supervisor` block — restart failed agents |
| No Zig-level enforcement | Enforce | `enforce` block — compile-time policy checks |
| No eBPF evidence collection | Testify | `ebpf` block — kernel-level event collection |
| No CrabCC index integration | Index | `crabcc` block — symbol-level code indexing |
| No multi-surface reveal | Reveal | `surface` block — TUI + web + API multi-render |

## Architecture Layers

```
Layer 0: ADR (decisions)
Layer 1: Blocks (29 primitives) + Pre-hooks (7)
Layer 2: Agent Hooks (10 events, 9 async)
Layer 3: Plugins (7 capability providers)
Layer 4: Orchestrator + Ralph Loop
Layer 5: TUI + Widgets (6) + CLI (5)
```

## Dyad — Two Machines, One Context

```
M1 (arm64) ←── NATS dyad channel ──→ dev-cx53 (amd64)
     │                                      │
     ├── Own TUI                            ├── Own TUI
     ├── Own orchestrator                   ├── Own orchestrator
     ├── Own brain                          ├── Own brain
     │                                      │
     └── Shared POV ────────────────────────┘
         Shared memos (dyad scope)
         Dual verification
         Failover (30s)
```
