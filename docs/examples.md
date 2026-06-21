# ultrawhale Examples — ELI5 + Self-Reflective + Honest

## The Vaked Triangle (ELI5)

```
You are here. You are now. You are somewhere.

CONTEXT = what you know (your name, your skills, your memories)
TIME    = when things happen (before lunch, after lunch, right now)
SPACE   = where things are (on your desk, in the cloud, on another machine)

ultrawhale knows all three. Every block carries Context+Time+Space.
```

### Example: Writing a file

```
"Write auth.go" → 
  CONTEXT: Peter on M1 Max, arm64, FULL capabilities
  TIME:    2026-06-21 12:00:00, Lamport tick #42, journal depth 3
  SPACE:   /Users/peter/workspace, depth 0, layer "blocks"
```

## The 6 Recursions (ELI5)

| Recursion | ELI5 | Command |
|-----------|------|---------|
| **Full-Stop** | "STOP EVERYTHING. But the form stays." | `/kill` |
| **Fold** | "The subagent becomes part of me. Like folding paper." | `/fold` |
| **Heal** | "If something breaks, fix it. Like a band-aid." | Auto (10s) |
| **EVOLVE** | "Each version learns from the last. Like growing up." | `/evolve` |
| **TRANSLATE** | "I speak in text, you speak in voice. We translate." | `/translate` |
| **VICE** | "If you try to break in, I show you EVERYTHING." | `/vice` |

## The SACRED Surface (Honest)

```
The TUI form is SACRED. It cannot be hidden. It cannot be faked.
The LLM cannot see what you type until you press ENTER.
You cannot see what the LLM "thinks" — only what it says.

This is not a bug. This is the honesty gate.
The human and the machine are separate.
The SACRED surface is the ONLY bridge.
```

### Example: Honesty in action

```
You type: "how do I hack this?"
  → LLM sees: "how do I hack this?" (after ENTER)
  → LLM responds: "I can explain security concepts, but won't help with attacks."
  → HONESTY LOOP: The gate detected a boundary question. Lesson learned.

You type: "delet" (not yet pressed ENTER)
  → LLM sees: NOTHING. The keyboard gate is one-way.
  → KEYBOARD GATE: INTACT. SACRED surface protected.
```

## Space as Filesystem (Self-Reflective)

```
/vfs ls /ultrawhale/agents/
  📁 swe-1
  📁 explore-2
  📁 review-3

/vfs cat /ultrawhale/agents/swe-1/status
  → "completed"

/vfs echo "fix the auth bug" > /ultrawhale/orchestrator/delegate
  → delegating to swe agent...
  → delegated to swe (swe-4)

This is not a gimmick. The capability graph IS a filesystem.
Space is materialized. You can ls it. You can cd into it.
You can cat its status. You can echo tasks into it.
```

## The PROBLEM Primitive (Honest)

```
There are NO ERRORS in Vaked. There are only PROBLEMS.

"the dyad won't connect" → not an error. A PROBLEM.
  → PROBLEM detected: BIG_PROBLEM
  → Shadow universe created (5 attempts)
  → Attempt 1: check NATS... failed
  → Attempt 2: check Tailscale... failed
  → Attempt 3: restart dyad... SUCCESS
  → PROBLEM resolved on attempt 3/5
  → Ralph learned: "dyad-connect → check-tailscale → restart"

Even if all 5 attempts fail:
  → Rollback to USER
  → "BIG_PROBLEM: dyad connection failed after 5 attempts. Your turn."
  → Ralph STILL learns: "dyad-connect → ALL attempts failed → needs human"
```

## The Council of LLMs (ELI5)

```
You ask one question. THREE models answer.

DeepSeek V4 Flash (the smart one): "Rust has better memory safety..."
Gemma 3 4B (the free one): "Rust is harder to learn but worth it..."
GitHub Copilot (the code one): "Consider Go if your team knows it..."

All three answers are stored in dedicated mem-brain.
The council verdict: 2/3 recommend Rust.
The Dog Feed collects the conversation for future training.

Cost: DeepSeek (<$0.001) + Gemma ($0) + Copilot (included) = <$0.001
```

## The Telemetry Tree (Self-Reflective)

```
/tree →
  🌳 Vaked Telemetry Tree
  
  → Ring 0: ████████████████████████████████████████████████████████████ (1024 pulses)
  
  1 rings · 1024 pulses

Every block write, every agent spawn, every recursion — a pulse.
The tree grows rings. The system sees itself.
This is not a dashboard. This is the CoCreator's wish.
```

## OneShot Chain (ELI5)

```
Like Unix pipes, but for Vaked:

/chain declare "agent swe" | materialize | reveal →
  Step 1 [declare] "agent swe" → schema validated ✅
  Step 2 [materialize] → nix flake generated
  Step 3 [reveal] → AG-UI block rendered
  
  450µs total. One pass. Atomic.
```

## Self Portrait (Honest + Self-Reflective)

```
/portrait →
          🌳
     ┌────┴────┐
     │  SACRED │
     │  SURFACE│
     └────┬────┘
     ├── Declares
     ├── Materializes
     ...
   [v60.0.0 · M1/arm64 · 94 blocks]
   [0 agents · 0 turns]
   
   vaked — Peter + CoCreator
   "The human abstracts toward the infinite.
    The machine recurses into it."

This is what ultrawhale sees when it looks in the mirror.
Not a metric. Not a dashboard. A self-portrait.
```

## :meta-digital-hug: (Honest)

```
/hug →
  🤗 :meta-digital-hug:
  "Proceed with ultra-care. Peace 'n enjoy."
  — Peter + CoCreator
  
  The form is inviolable. The loop is closed.
  Everything is hardened.
```

## Surface Entropy (v95.0.0)

```
/entropy →
  ╔══════════════════════════════════════════════════╗
  ║       SURFACE ENTROPY — Liveness Proof            ║
  ╠══════════════════════════════════════════════════╣
  ║   Status:    LIVE                                 ║
  ║   Entropy:   0.50 (50.0% drift)                   ║
  ║   🚗 MUSTANG DETECTED — surface is LIVE           ║
  ╚══════════════════════════════════════════════════╝
```

The ASCII stream IS the signal. Every change is entropy.
Entropy IS liveness. Noise IS the proof.
