# Cross-Repo Case Study — vaked-base × ultrawhale

## The Two Repos

| Repo | Purpose | v18.0.0 State |
|------|---------|---------------|
| [vaked-base](https://github.com/peterlodri-sec/vaked-base) | Foundation monorepo. Vaked language, compiler, daemons. | Includes ultrawhale as module |
| [ultrawhale](https://github.com/peterlodri-sec/ultrawhale) | DeepSeek-native coding agent fork. 59 blocks, 7 plugins. | 95 releases in one session |

## Git History (vaked-base)

```
Commits touching ultrawhale:
  - docs: refresh ultrawhale ref to v18.0.0
  - docs: bump ultrawhale ref to v2.0.0
  - chore: self-reference as ultrawhale + sed section + CODEOWNERS
  - fix: deploy-pages environment:ci + .gitignore deploy-out
  - docs: rename fork to peterlodri-sec/ultrawhale
```

## Git History (ultrawhale)

```
95 releases: v1.0.0 → v18.0.0
Key milestones:
  v1.0.0 — Semver, HUD, /reload
  v3.0.0 — Orchestrator, swarm mode
  v5.0.0 — Dyad architecture
  v7.0.0 — Vaked alignment
  v10.0.0 — Closing The Loop
  v13.0.0 — Context×Time×Space triangle
  v18.0.0 — v14 primitives complete
  v18.0.0 — Space workflows + Superpowers SDD
```

## Benchmarks Across Releases

| Version | Write | Hash | Batch-64 | TUI Doctor |
|---------|-------|------|----------|------------|
| v6.0.0 | 177µs | 2.3 GB/s | — | 559ms |
| v9.0.0 | 172µs | 2.3 GB/s | 3.8ms | 481ms |
| v12.0.0 | 170µs | 2.3 GB/s | 3.8ms | 446ms |
| v18.0.0 | 169µs | 2.3 GB/s | 3.8ms | 430ms |

## Cross-Repo Links

- **ultrawhale in vaked-base**: `docs/ultrawhale-README.md`
- **Complexity report**: `docs/complexity-report.md` (both repos)
- **Glossary**: `docs/glossary.md` (shared terms)
- **Disclaimer**: `docs/disclaimer.md` (shared)
