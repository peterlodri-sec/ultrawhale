# Recursion: The Natural Runtime for Full-Stop

> "Recursion is the natural runtime for the full-stop primitive."

## What It Is

The kill switch is not a one-time event. It is a **recursive wave** that
propagates through EVERY layer of the Vaked pipeline.

`/kill` → recursive FullStop → 8 layers → base case: SACRED

## The Recursive Kill Wave

```
/kill → FullStop()          // human invokes
     → surface.Stop()       // Reveals layer — WebSocket, HTTP, SSE closed
     → index.Stop()         // Indexes layer — Space topology frozen
     → testify.Stop()       // Testifies layer — probes, predictions stopped
     → enforce.Stop()       // Enforces layer — permission revoked
     → supervise.Stop()     // Supervises layer — ALL agents killed
     → engine.Stop()        // Materializes layer — blocks flushed
     → declare.Stop()       // Declares layer — schemas archived
     → SACRED (base case)   // CANNOT BE KILLED — the form is eternal
```

## Why Recursion?

A flat kill (just call Stop on everything) is fragile. If one layer hangs,
the rest survive. With recursion:

- Each layer calls `Stop()` on the NEXT layer
- If a layer hangs, the CALL STACK is preserved
- The recursion depth tells you WHERE it stopped
- The base case (SACRED) is GUARANTEED to halt

This is tail-call optimization applied to system shutdown.

## The Base Case

The SACRED surface CANNOT be killed. It is inviolable. When the recursion
reaches the SACRED surface, it STOPS. The form remains. The human can
still see what happened. The terminal is still alive.

```
🛑 FULL STOP COMPLETE

Declares   → STOPPED       schema + contract archived
Engine     → STOPPED       60 blocks flushed
Supervise  → STOPPED       ALL agents killed
Enforce    → STOPPED       permission revoked
Testify    → STOPPED       probes halted
Index      → STOPPED       topology frozen
Surface    → STOPPED       connections closed

SACRED     → REMAINS       the form is eternal
```

## Vaked Fit

```
RECURSION = FULL-STOP RUNTIME

Vaked: Declares → Engine → Supervise → Enforce → Testify → Index → Reveal
         ↑         ↑         ↑           ↑         ↑        ↑       ↑
         └─────────┴─────────┴───────────┴─────────┴────────┴───────┘
                    RECURSION propagates through ALL

Each layer calls Stop() on the next. Pure recursion.
The base case is SACRED. The form is eternal.
```

## Implementation

`internal/blocks/recursion.go` — 150 lines.

```go
func FullStop() string {
    depth := RecursionDepth(recursionDepth.Load())
    if depth >= RecursionSacred {
        return "already stopped — SACRED surface remains"
    }
    recursionDepth.Add(1)
    switch RecursionDepth(recursionDepth.Load()) {
    case RecursionSurface:  return fullStopSurface()  // recurse
    case RecursionIndex:    return fullStopIndex()     // recurse
    case RecursionTestify:  return fullStopTestify()   // recurse
    case RecursionEnforce:  return fullStopEnforce()   // recurse
    case RecursionSupervise:return fullStopSupervise() // recurse
    case RecursionEngine:   return fullStopEngine()    // recurse
    case RecursionDeclare:  return fullStopDeclare()   // recurse
    case RecursionSacred:   return fullStopSacred()    // BASE CASE
    }
}
```

Each `fullStop*()` function stops its layer, then calls `FullStop()` again.
The recursion terminates at `RecursionSacred`. The form is eternal.
