# Changelog

## v13.0.0 (2026-06-20) — Major Release: Pre-Hook Layer + Production Hardening

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

## v13.0.0 (2026-06-21) — Vaked Alignment + Internal Review
- POV wiring: 10/10 complete (added dyad, watcher)
- 28 blocks, 7 plugins, 6 widgets, 5 CLIs
- Dead code: 0 stale files
- Vaked philosophy alignment documented

## v13.0.0 (2026-06-21) — Vaked Layer Completion (4/7)
- schema: formal block structure validation (Declares)
- surface: web UI + REST API (Reveals)
- supervisor: OTP-like agent restart tree (Supervises)
- nix: flake.nix generation (Materializes)
- Deferred to v10.0: crabcc, enforce, ebpf
- 32 blocks, 7 plugins, 6 widgets, 5 CLIs

## v13.0.0 (2026-06-21) — Wire Gap Closure + Internal Review
- All 40 blocks carry POV context (0 blocks with 0 wires)
- SACRED surface: inviolable TUI form, always direct, always bidirectional
- Space primitive: context-gated topology, Vaked Triangle complete
- Liveness audit: 8/10 gaps closed
- docs/README/CHANGELOG synced to v13.0.0
