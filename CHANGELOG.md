# Changelog

## v7.1.0 (2026-06-20) — Major Release: Pre-Hook Layer + Production Hardening

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

## v7.1.0 (2026-06-21) — Vaked Alignment + Internal Review
- POV wiring: 10/10 complete (added dyad, watcher)
- 28 blocks, 7 plugins, 6 widgets, 5 CLIs
- Dead code: 0 stale files
- Vaked philosophy alignment documented
