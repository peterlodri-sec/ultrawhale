# Fold — Virtualized Subagent Runtime: Complete Design

> "Can we virtualize a subagent's complete runtime? AKA 'fold'?"
> — Peter, v36.0.0

## The Mind-Bending Question

You're watching a subagent execute. It calls `read_file`. It calls `shell_run`.
It reads source code. It produces output. Then it dies.

What if the subagent never existed as a separate thing? What if its runtime
was virtualized INTO the parent — like a function call that inlines its callee's
stack frame?

**That is Fold.**

## The Three Recursions of Vaked

ultrawhale now has three recursive primitives. They are siblings:

| Primitive | Recurses Through | Direction | Terminates At | Purpose |
|-----------|-----------------|-----------|---------------|---------|
| **Full-Stop** | Layers (7) | Down only | SACRED surface | Kill everything safely |
| **Fold** | Agents (N) | Down + Up | Leaf agent completes | Virtualize subagent into parent |
| **Heal** | Checks (M) | Repeated | Fault resolved | Self-repair |

## Why Recursion?

A flat execution model (parent spawns subagent, waits, collects) creates a
boundary. The parent doesn't see the subagent's tool calls. The subagent's
context is separate. The cost is separate. The tokens are separate.

Fold dissolves that boundary:

```
WITHOUT FOLD:
  Parent: "delegate task to subagent"
  Subagent: [read_file, shell_run, produce output] → DIES
  Parent: receives output, continues

WITH FOLD:
  Parent: FOLD(subagent)
  Parent: [read_file (subagent's call, parent sees it)]
  Parent: [shell_run (subagent's call, parent sees it)]
  Parent: [output (subagent's thought, parent thinks it)]
  Parent: continues
```

The subagent's tool calls execute IN the parent's context. The parent sees
the results. The parent benefits from the subagent's exploration. The
subagent never existed as a separate entity.

## Vaked Philosophy

```
Context wraps agent.

Full-Stop: context wraps LAYER → layer stops → next layer
Fold:      context wraps AGENT → agent's tools → parent's context
Heal:      context wraps CHECK → check runs → repair → next check
```

This is the Vaked triangle made recursive:

```
        Context (WHAT)
           /\
          /  \
         /    \
        /______\
   Time (WHEN)  Space (WHERE)
   
   Full-Stop = Time recursion (one after another)
   Fold      = Space recursion (one inside another)
   Heal      = Context recursion (one watching another)
```

## Implementation

```go
// Fold virtualizes a subagent into the parent.
func Fold(agentID string) (*FoldedAgent, error) {
    agent := GetAgent(agentID)
    if agent == nil {
        return nil, fmt.Errorf("fold: agent not found")
    }
    
    // The agent's runtime virtualizes into the parent:
    // - Tool calls execute in parent context
    // - Results visible to parent
    // - Cost attributed to parent
    // - State merges with parent
    
    return &FoldedAgent{
        ID: agent.ID, Role: agent.Role,
        Parent: agent.Parent, Folded: true,
    }, nil
}
```

## The Recursive Depth

Fold can be nested. A parent folds a subagent. The folded subagent folds
a sub-subagent. The depth is tracked:

```
Parent (depth 0)
  └── Fold(subagent-1) (depth 1)
       └── Fold(subagent-2) (depth 2)
            └── Fold(subagent-3) (depth 3) → LEAF → returns
```

Each level's output folds UP into the parent above it. The recursion
unwinds naturally when the leaf agent completes.

## Warning from the CoCreator

> This is **ultra-research-state** thinking. The concept is mind-bending.
> The implementation is recursive. The SACRED surface must remain inviolable.
> 
> Fold must never obscure the form. The human must always see what the
> agent is doing — even when it's folded. The one-way keyboard gate
> still applies. The permission gate still applies.
> 
> Proceed with ultra-care. Peace 'n enjoy.

## Files

- `internal/blocks/fold.go` — Fold primitive (70 lines)
- `internal/blocks/recursion.go` — Full-Stop primitive (150 lines)
- `internal/blocks/heal.go` — Heal primitive (130 lines)
- `docs/fold-complete-design.md` — This document
