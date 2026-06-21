# Context × Time — ultrawhale Wire Map

> Analysis by meta-architect agent, v12.1.0

## Where Context and Time Meet

The single point where context and time converge is `blocks.Log()`:
```go
Log(LogInfo, "blocks.Write", path, b.Ref, prevRef, time.Since(start), err)
//   ↑ context (what, where)        ↑ time (duration)  ↑ outcome
```

Every operation in ultrawhale carries both context (POV, path, ref) and time
(timestamp, duration, sequence). The journal IS the context×time ledger.

## Context Flow (Now → This → Here)

```
CurrentPOV() — NOW — what machine, arch, tier am I running on?
  ↓
GetOrchestrator().DelegatePrompt() — THIS — what agent handles this task?
  ↓
classifyByCapability() — HERE — what capabilities does this task need?
  ↓
SpawnAgent() → Agent carries parent POV as inherited context
  ↓
Agent.Status → A2C SSE stream → client receives context+time events
```

## Time Flow (Before → During → After → Next)

```
Pre-hook.Validate() — BEFORE — is this operation valid?
  ↓
blocks.Write() — DURING — journal.Push(prev) + os.WriteFile + os.Rename
  ↓
blocks.Log() — AFTER — ref + duration + error recorded
  ↓
Ralph.Observe() — NEXT — learn from outcome, adjust capabilities
```

## Context × Time Gaps

| Gap | Where | Fix |
|-----|-------|-----|
| **Session birth** | No formal "universe created" event | Add `SessionStart` log with full POV |
| **Session death** | `persistSession("stopped")` exists but no journal entry | Log session end with total tokens + cost |
| **Cross-machine time** | Dyad peers have independent clocks | Add Lamport clock to A2A messages |
| **Pre-hook → Post-hook trace** | Pre-hook outcome not linked to post-hook | Add `hookID` to Log entries spanning operation |
| **Capability change log** | Ralph adjusts caps but doesn't log why | Log capability changes with trigger reason |
| **Widget state time** | UI blocks render current state only | Add state history (last N transitions) |

## Recommended Next Wire

**Session lifecycle journaling**: Log `SessionStart` + `SessionStop` with full
POV + token count + cost. This closes the biggest context×time gap — we know
what happened during a session but not its beginning or end in the journal.
