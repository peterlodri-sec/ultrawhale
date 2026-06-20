# Changelog

## v5.0.0 (2026-06-20) — First Complete Architecture Release
- P0: Workflow ↔ Orchestrator complete wire. classifyPrompt matches 'workflow' globally.
- P1: Ralph patterns persist to brain long-term memory (survive reboots).
- P3 foundation: Auto-rollback on 3 consecutive failures.
- 21 blocks primitives, 6 plugins, 5 widgets, 5 CLIs.
- Tailscale first-class primitive. SSH bulletproofing (6 fixes).
- BLAKE3 + xxHash + diff + pool primitives.
- Kernel optimizations: lock-free log, sharded journal, tmp cleanup.
- ultrawhale-shell remote daemon (5MB static, macOS+Linux).
- Documentation site at vaked.dev/ultrawhale.

## v4.x (2026-06-20)
- SSH primitive (Tool-level, bao+local key mgmt, PID lifecycle)
- Kernel optimizations (lock-free log, sharded journal)
- Ralph Loop (self-improving agent cycle, versioned, rollback)
- Tool primitive v2 (typed, scoped, asm-accelerated, 14 tools)
- Floating ControlPanel widget, InfraBar, Sidepanel
- AG-UI ChatBlock + Shader hooks, render.go refactor
- bench-tui v2 (load simulation + screenshot + JSON/MD reports)

## v3.x (2026-06-20)
- Swarm mode (persistent workers + nested AgentField)
- Edge agent primitive (CF Workers + fiber journal)
- Orchestrator agent loop wire
- Tool cache (KV, 5-min TTL, per-agent)
- Plugin refactor (zero boilerplate, skillsimprover deleted)
- POV 10/10 wiring complete

## v2.x (2026-06-20)
- Blocks engine (14 primitives, 3-tier hash)
- Floating widgets (5 total)
- Ultracode 7-phase loop
- macOS Apple Silicon support (28MB Mach-O arm64)
- YOLO mode defaults, subagent delegation

## v1.x (2026-06-20)
- Semver versioning, HUD statusline
- /reload command, deep hooks (10 events, 9 async)
- Superpowers plugin (bao, langfuse, NATS)
- Repomap SIMD (2,361 MB/s)
- AG-UI themes (dense, cyberpunk, graveyard)
- Native Go tools (gh, grep, git)
- macOS cross-compile, native M1 binary
