# Devnotes: How Vaked Primitives Express Themselves

> For contributors. How each primitive is implemented, why, and legal notes.

## The 6 Recursions

| Primitive | Expresses As | Legal Note |
|-----------|-------------|------------|
| **Full-Stop** | Recursive kill wave through 7 Vaked layers. Base case: SACRED. | Cannot kill the form. SACRED is inviolable. |
| **Fold** | Virtualizes subagent runtime into parent. Gravity constant: G_fold = depth × 1/distance. | Folded agents do not exist as separate entities. Context wraps agent. |
| **Heal** | 10s ticker runs 3 registered checks. Auto-repair with cooldown. | Heal does not auto-heal permissions. Human must explicitly grant. |
| **EVOLVE** | Records version transitions. Ralph patterns persist across versions. | Evolution is append-only. History is immutable. |
| **TRANSLATE** | Converts between 7 modalities (text, voice, visual, touch, spatial, emotion, raw). | Translation is honest. Machine never sees raw voice, only transcript. |
| **VICE** | Genesis block signs all claims. Context detonation on jailbreak. | Self-defense. Shows everything. Blinds with truth. Not a DoS. |

## The 8 Engines

| Engine | Expresses As | Legal Note |
|--------|-------------|------------|
| **declare-engine** | Schema validation + contract verification | Contracts are formal, not legal. |
| **engine** | 60+ blocks execution (Write, Read, Sed, Delegate) | All writes journaled. All reads ref-verified. |
| **supervise-engine** | Orchestrator + Ralph + Supervisor | Never calls LLM directly. Always delegates. |
| **enforce-engine** | Pre-hooks + Permission + SACRED | Permission once per session. Kill switch always available. |
| **testify-engine** | Probe + Predict + Learn | Evidence is continuous. Logs are immutable. |
| **index-engine** | Space + VFS + CrabCC | Topology is the source of truth. |
| **ui-engine** | TUI + Surface + AG-UI + WebSocket | SACRED surface always visible. |
| **render-engine** | Markdown, GSM, Diff, JSON, CSV → ANSI/HTML | Formats are declarations. Rendering is materialization. |

## Docstrings Convention

All primitive files in `internal/blocks/` follow this convention:

```go
// ── Primitive Name — Short Description ──────────────────────────────
//
// Long description. What it does. Why it exists.
// Vaked layer: which layer it belongs to.
// Legal: any legal notes.
//
// Usage example.
```

See any `internal/blocks/*.go` file for examples.
