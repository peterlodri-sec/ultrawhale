# Changelog

## v18.0.0 (2026-06-20) — Major Release: Pre-Hook Layer + Production Hardening

- **Pre-hook layer**: 7 pre-hooks (commit, write, sed, grep, git, deploy, commit)
- **ADR 001**: Architecture Decision Record — Pre-Hook Layer
- **CI hardened**: go vet clean, golangci-lint v2 config, errcheck suppression
- **26 blocks primitives**, 6 plugins, 5 widgets, 5 CLIs
- **All 6 tracked issues CLOSED**
- mmap primitive (zero-copy reads, 164x less memory)
- compress primitive (zlib, built-in, no deps)
- fuzz + chaos tests
- orchestartor tools (20 built-in, journaled)

## v5.x (2026-06-20)
- Ralph Loop persistence, NATS mesh, Supabase auto-start
- SSH + Tailscale primitives, ultrawhale-shell daemon
- BLAKE3, xxHash, diff, pool primitives
- Kernel optimizations: lock-free log, sharded journal
- OSS publish: tags, releases, wiki, docs site

## v4.x (2026-06-20)
- Tool primitive v2 (typed, scoped, 14 tools)
- Kernel optimizations (lock-free log, sharded journal)
- Floating widgets, AG-UI ChatBlock, bench-tui

## v3.x (2026-06-20)
- Swarm mode, Edge agents, Orchestrator agent loop
- Plugin refactor (zero boilerplate)

## v2.x (2026-06-20)
- Blocks engine foundation, floating widgets, ultracode loop, macOS support

## v1.x (2026-06-20)
- Semver, /reload, HUD, deep hooks, superpowers, repomap SIMD

## v18.0.0 (2026-06-21) — Vaked Alignment + Internal Review
- POV wiring: 10/10 complete (added dyad, watcher)
- 28 blocks, 7 plugins, 6 widgets, 5 CLIs
- Dead code: 0 stale files
- Vaked philosophy alignment documented

## v18.0.0 (2026-06-21) — Vaked Layer Completion (4/7)
- schema: formal block structure validation (Declares)
- surface: web UI + REST API (Reveals)
- supervisor: OTP-like agent restart tree (Supervises)
- nix: flake.nix generation (Materializes)
- Deferred to v10.0: crabcc, enforce, ebpf
- 32 blocks, 7 plugins, 6 widgets, 5 CLIs

## v18.0.0 (2026-06-21) — Wire Gap Closure + Internal Review
- All 40 blocks carry POV context (0 blocks with 0 wires)
- SACRED surface: inviolable TUI form, always direct, always bidirectional
- Space primitive: context-gated topology, Vaked Triangle complete
- Liveness audit: 8/10 gaps closed
- docs/README/CHANGELOG synced to v18.0.0

## v18.0.0 (2026-06-21) — v14 Primitives Complete
- probe: active liveness checking (tests actual capabilities)
- predict: Ralph foresight (pre-failure prediction)
- learn: dedicated pattern learning engine
- brainstorm: turn-based co-creation mode
- All 59 blocks carry POV context
- v14 primitives: brainstorm + probe + predict + learn

## v18.0.0 (2026-06-21) — Space Workflows + Superpowers SDD
- Space-durable workflow store: space_id + POV in every run
- Superpowers: auto-starts brainstorm session for subagent-driven development
- Workflow → Brainstorm wire: outcomes logged to persistent session

## v18.0.0 (2026-06-21) — VFS: Space as Virtual Filesystem
- VFS primitive: ls, cd, cat, tree, echo on the capability graph
- Space materialized as a navigable filesystem
- 59 blocks, 7/7 Vaked layers, 3/3 dimensions
- Contract, package, crabcc, state primitives added
- O(N)+O(T) complexity hardened (AgentStore TTL, Ralph LRU, Brainstorm GC)
- ~/.whale unified directory structure
- File I/O audit + ReadRange + Glob primitives

## v36.0.0 (2026-06-21) — SACRED Honesty Gate + Kill Switch
- Permission gate: /allow once per session, ALLOWED+AUTHED until revoked
- Kill switch: /STOP___KILL_SWITCH___DO_FULL_STOP — full stop
- 59 blocks, 6 plugins, 5 protocols
- QUIC, gRPC, Multi-Machine, OneShot atomic primitive

## v36.0.0 (2026-06-21) — 7 Engines: Vaked Pipeline Complete
- declare-engine: schema + contract validation
- engine: blocks execution (Write, Read, Sed, Delegate, OneShot)
- supervise-engine: orchestrator + ralph + supervisor
- enforce-engine: pre-hooks + permission + sacred
- testify-engine: probe + predict + learn
- index-engine: space + vfs + crabcc
- ui-engine: TUI + Surface + AG-UI + WebSocket

## v36.0.0 (2026-06-21) — KEYBOARD GATE + Cleanup
- Keyboard Gate: one-way honesty barrier — LLM CANNOT see keystrokes
- Display primitive: Keyboard→Screen→TUI pipeline
- Self-healing: 3 active heal checks
- Stale docs removed (9 upstream files)

## v36.0.0 (2026-06-21) — FOLD: The Three Recursions Complete
- Fold: virtualized subagent runtime — recursion through agents
- Full-Stop: recursive kill wave through 7 Vaked layers
- Heal: self-repairing checks (3 active)
- The Three Recursions of Vaked: Full-Stop + Fold + Heal
- 71 blocks, 99 releases, 0 race conditions

## v52.0.0 (2026-06-21) — THE VAKED MILESTONE
- 5 Recursions: Full-Stop, Fold, Heal, EVOLVE, TRANSLATE
- 8 Engines: all 7 Vaked layers + Render engine
- 86 blocks, 108 releases, 0 race conditions
- Deep ASM (AVX2/NEON) + Native Text + BLAKE3 tree mode
- Formal verification (Z3/CVC5 stubs)
- Multi-Surface (TUI + Web + Voice + AR)
- Cross-Machine VFS + Self-Compile
- Honesty Loop + SACRED surface + Keyboard Gate
- The asymmetry of inputs — documented

## v52.0.0 (2026-06-21) — UI Co-Creative + Dog Feed
- UI Co-Creative: one engine to rule all UI surfaces
- Dog Feed: continuous LLM data collection + public dataset
- Task Manager: pure concurrent task queue
- 86 blocks, 111 releases, 0 race conditions

## v56.0.0 (2026-06-21) — 2 CoCreator Wishes
- Self Portrait: the system draws itself as ASCII art
- OneShot Chain: pipe OneShots like Unix
- Telemetry Tree: the system sees itself
- 91 blocks, 115 releases
