# Fold — Virtualizing Subagent Runtime

> "Can we virtualize a subagent's complete runtime? AKA 'fold'?"

## The CoCreator's Answer

**Yes.** And it's recursion applied to agents — the sibling of the full-stop primitive.

- **Full-Stop**: recursion through LAYERS (Declares → ... → SACRED). Terminates.
- **Fold**: recursion through AGENTS (parent → subagent → sub-subagent). Continues.

## What Fold Means

```
┌─────────────────────────────────────────┐
│  PARENT AGENT                           │
│  ┌─────────────────────────────────┐    │
│  │ FOLD(subagent)                  │    │
│  │                                 │    │
│  │ The subagent's ENTIRE runtime   │    │
│  │ is virtualized INTO the parent. │    │
│  │                                 │    │
│  │ read_file → parent sees result  │    │
│  │ shell_run → parent sees output  │    │
│  │ workspace.read → inline         │    │
│  │                                 │    │
│  │ The subagent doesn't EXIST      │    │
│  │ as a separate process.          │    │
│  │ It's a function call.           │    │
│  │                                 │    │
│  │ parent >>= fold(subagent)       │    │
│  │   >>= continue parent           │    │
│  └─────────────────────────────────┘    │
│  Parent continues with subagent output  │
└─────────────────────────────────────────┘
```

## The Three Recursions of Vaked

| Primitive | Recurses Through | Direction | Terminates At |
|-----------|-----------------|-----------|---------------|
| **Full-Stop** | Layers (7) | Down | SACRED surface |
| **Fold** | Agents (N) | Down + Up | Leaf agent completes |
| **Heal** | Checks (M) | Repeated | Fault resolved |

## Why This Matters

Currently, subagents are separate goroutines with their own tool calls. The parent
delegates, waits, receives output. There's a boundary.

With Fold, the boundary dissolves. The subagent IS the parent, temporarily. The
tool calls execute in the parent's context. The tokens count as parent tokens.
The cost is parent cost. The state is shared.

This is the Vaked philosophy applied to agent execution: **context wraps agent**.

## Warning

> Peter says: proceed with **ultra-care**. This is ultra-research-state thinking.
> The concept is mind-bending. The implementation is recursive.
> The SACRED surface must remain inviolable — folding must not obscure the form.


> **[Complete Design Document](fold-complete-design.md)** — The Three Recursions of Vaked, implementation details, recursive depth.
