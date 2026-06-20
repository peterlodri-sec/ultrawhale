# Native Agent Loop — ultracode

The 7-phase autonomous coding loop: plan → implement → test → review → fix → verify → commit. All writes journaled via blocks, all traces to Langfuse, all events to NATS.

## Phases

| # | Phase | Failure behavior |
|---|-------|-----------------|
| 1 | plan | Retry with different approach |
| 2 | implement | blocks.Rollback restores pre-write state |
| 3 | test | Failure → rollback to pre-implement |
| 4 | review | Issues → phase 5 (fix) |
| 5 | fix | Apply review fixes |
| 6 | verify | Re-test after fixes |
| 7 | commit | Sign + commit |

## TUI commands

```
/ultracode start   → begin loop
/ultracode status  → ●→✓→✓→·→·→·→·
/ultracode next    → advance phase
/ultracode fail    → mark current failed
```

## Phase summary icons

```
● running  ✓ passed  ✗ failed  ◌ skipped  · pending
```

## Plugin reuse

All 6 plugins participate:
- **repomap**: context injection per phase
- **blocks**: journaled writes, rollback on failure
- **langfuse**: phase traces with POV metadata
- **nats**: phase lifecycle events
- **superpowers**: bao secrets for commit signing
- **agentfield**: DID identity, workflow persistence

## POV — Point of View

Every phase carries a POV context:

```go
type POV struct {
    Agent    string // "ultrawhale"
    Version  string // "v8.1.0"
    Machine  string // "M1" | "dev-cx53"
    Arch     string // "arm64" | "amd64"
    Tier     string // "go" | "asm" | "gpu"
    Session  string
    Mode     string // "agent" | "ultracode"
}
```

POV is injected into: LogSink toasts, Langfuse trace metadata, AgentField API responses, HUD right section.

## HUD statusline

```
[deepseek-v4-flash]  ⎇ main     ● 2:35 · 4821t · 85/s · ⎆98%   dev-cx53·amd64  342MB · ⚙6
 ─── left ───────────     ────── center ────────────────────   ─── right ───────
```

Right section shows POV: `M1·arm64` or `dev-cx53·amd64`.


## Self vs Current

/self: who am I (ultrawhale v8.1.0, 6 plugins, DID).
/current: what is happening now (idle/busy, tokens, cache, memory, cost).

## Full TUI Commands

/reload all|status|hooks|theme|doctor
/ultracode start|status|next|fail
/self /current
Ctrl+Shift+T: cycle themes
Ctrl+Shift+Z: zen mode
Ctrl+Shift+B: shader toggle
Ctrl+Shift+S: sidebar toggle
