# Primitive Mapping — ultrawhale v18.0.0

## Vaked Layer Coverage

| Layer | Blocks | Count |
|-------|--------|-------|
| **Declares** | schema, capabilities, tool, prehook | 4 |
| **Materializes** | nix, dotwhale, io, export | 4 |
| **Supervises** | pov, self, current, cost, codewhale, agent, swarm, orchestrator, ralph, brainstorm, supervisor | 11 |
| **Enforces** | block, journal, log, hash, sed, diff, compress, mmap, pool, glob, blake3, xxhash | 12 |
| **Testifies** | probe, predict, learn, edge, ssh, tailscale, watch | 7 |
| **Indexes** | repomap, space | 2 |
| **Reveals** | sacred, surface, ui, a2a, a2c, mesh_agent, dyad, notify | 8 |

**Total: 53 blocks mapped to 7 Vaked layers.**

## Dimension Coverage

| Dimension | Blocks | Count |
|-----------|--------|-------|
| **Context (WHAT)** | pov, self, current, cost, codewhale, capabilities, schema | 7 |
| **Time (WHEN)** | journal, log, prehook, ralph, brainstorm, probe, predict, learn | 8 |
| **Space (WHERE)** | space, dyad, ssh, tailscale, edge, mesh_agent, glob | 7 |

**26 blocks carry a specific dimension. 22 are infrastructure (no single dimension).**

## Gap Analysis

| Gap | Detail |
|-----|--------|
| Declares layer | 4 blocks — could add `contract` block for formal verification |
| Materializes layer | 4 blocks — could add `package` block for brew/docker/npm |
| Indexes layer | 2 blocks — could add `crabcc` block for symbol indexing |
| Context dimension | 7 blocks — could add `state` block for state machine |
| Time dimension | 8 blocks — Lamport clock already covers cross-machine ordering |
| Space dimension | 7 blocks — topology already covers graph structure |
