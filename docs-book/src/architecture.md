# Architecture

28 blocks, 7 plugins, 6 widgets, 5 CLIs.

```
User → TUI → Orchestrator → Subagent/Swarm
         ↓
  Blocks Engine (28 primitives)
  ├── Content-addressed, journaled, 3-tier hash
  ├── BLAKE3 (v6.4), SIMD sed, mmap, compress
  └── POV wired across 12 systems

  Plugins (7): memory, repomap, NATS, Langfuse, superpowers, agentfield, vaked
  Widgets (6): InfraBar, Sidepanel, ControlPanel, Toast, WidgetBase, CardChoice
  CLIs (5): ultrawhale, setup, bench-tui, shell, shell-linux
```

## Layers

| Layer | Count | Purpose |
|-------|-------|---------|
| Blocks | 28 | Content-addressed primitives |
| Pre-hooks | 7 | Validation before operations |
| Plugins | 7 | Capability providers |
| Widgets | 6 | TUI presentation |
| CLIs | 5 | Entry points |
| Tools | 20 | Orchestrator tools |
