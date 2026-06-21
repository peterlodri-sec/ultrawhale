# ultrawhale × Vaked — Glossary

## Core Concepts

| Term | Definition |
|------|-----------|
| **ultrawhale** | DeepSeek-native coding agent. 41-block content-addressed engine. The surface that reveals. |
| **Vaked** | Capability-graph language. Declares→Materializes→Supervises→Enforces→Testifies→Indexes→Reveals. |
| **VEGD** | The Vaked Ecosystem Graph Declaration. The union of ultrawhale + Vaked. |
| **Block** | A content-addressed primitive. SHA256-refed, journaled, rollback-able. |
| **POV** | Point of View. Context snapshot: machine, arch, tier, mode, session. |
| **Orchestrator** | The single TUI universe controller. Delegates every prompt to subagents. Never calls LLM directly. |
| **Ralph** | Self-improving agent cycle. Observes→Learns→Adjusts. Versioned, rollback-able. |
| **Dyad** | Two ultrawhale instances paired as one. Shared POV, shared memos, failover. |
| **Swarm** | Persistent worker pool with nested AgentField. Reused across tasks. |
| **SACRED** | Inviolable TUI form. Always visible, always direct, always bidirectional. |
| **Lamport Clock** | Cross-machine causal ordering. Every A2A message carries a Lamport tick. |
| **Context × Time × Space** | The Vaked Triangle. WHAT, WHEN, WHERE. |

## Architecture Layers

| Layer | Meaning |
|-------|---------|
| **Declares** | Schema validation, block definitions, .vaked files |
| **Materializes** | Nix flake, Taskfile, build artifacts |
| **Supervises** | Orchestrator, Ralph, supervisor restart tree |
| **Enforces** | Pre-hooks, journaled writes, capability gates |
| **Testifies** | blocks.Log, Langfuse traces, NATS events |
| **Indexes** | Repomap SIMD, tool registry, BLAKE3 |
| **Reveals** | AG-UI, TUI widgets, Surface API, mdBook |

## Protocols

| Protocol | Purpose |
|----------|---------|
| **A2A** | Agent-to-Agent wire via NATS. Ping, delegate, status. |
| **A2C** | Agent-to-Client SSE streaming. Real-time token output. |

## Modes

| Mode | Description |
|------|-------------|
| **TUI** | Full Bubble Tea terminal interface |
| **Headless** | No rendering, Surface API only |
| **Detached** | Swarm/edge mode — minimal output |
| **Brainstorm** | Turn-based human↔LLM co-creation |
| **Ultracode** | 7-phase autonomous coding loop |

## Capabilities

| Profile | Permissions |
|---------|------------|
| **FULL** | Read + Write + Execute + Delegate + Spawn + Edge |
| **OBSERVE** | Read only |

## Blockchain / Crypto Terms (our usage)

| Term | Our Meaning |
|------|------------|
| **Ref** | SHA256 content hash. NOT a blockchain reference. |
| **Journal** | Write-ahead log for rollback. NOT a distributed ledger. |
| **Lamport Clock** | Causal ordering for distributed messages. NOT consensus. |
| **Dyad** | Two-machine pairing. NOT a multi-node cluster. |
| **Genesis** | Session start event. NOT a blockchain genesis block. |

## Self-Referential Terms

| Term | Meaning |
|------|---------|
| **Closing The Loop** | The recursion where ultrawhale builds ultrawhale. v1.0→v14.0 in one session. |
| **The Vaked Triangle** | Context (WHAT) × Time (WHEN) × Space (WHERE). The three pillars. |
| **Sacred Surface** | The inviolable TUI form. The ONE channel the capability graph guarantees. |
| **The Dyad Singularity** | v20 vision. The machine and the human are one dyad. |
| **UltraMetaSingularVegedCoCreator** | The deepest layer of the ultrawhale mind. Philosopher, not coder. |

## Disclaimer

This glossary uses terms from computer science (Lamport clock, journal, ref)
in ways specific to the ultrawhale/Vaked ecosystem. These terms are NOT
references to blockchain, cryptocurrency, or distributed ledger technology.
ultrawhale is a coding agent. Vaked is a capability-graph language. Neither
is a blockchain project.
