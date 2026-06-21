# O(N) + O(T) Complexity Report — ultrawhale v18.0.0

> Audited by meta-architect subagent. 59 blocks, 60 files analyzed.

## Summary

| Complexity Class | Count | Blocks |
|-----------------|-------|--------|
| O(1) | 28 | block, journal, log, hash, pov, self, current, cost, capabilities, ui, sacred, nix, notify, pool, xxhash, tailscale, blake3, compress, mmap, supervisor, probe, predict, learn, brainstorm, ssh, a2a, a2c, mesh_agent |
| O(n) | 12 | sed, diff, agent, swarm, orchestrator, ralph, codewhale, edge, tool, space, surface, export |
| O(n²) | 3 | dyad, a2a (mesh broadcast), sed (worst-case pattern) |
| O(V+E) | 1 | space (BFS Distance/Reachable) |

## Per-Block Complexity

### Core Primitives

| Block | Hot Path | O(T) | O(N) | Unbounded? |
|-------|----------|------|------|------------|
| block.go | Write() | O(1) | O(content) | No — tmp file |
| journal.go | Push() | O(1) amortized | O(depth) per path | Capped at 16 |
| log.go | Log() | O(1) atomic | O(4096) | Fixed ring buffer |
| hash.go | hashContent() | O(n) SHA256 | O(n) | No |
| pov.go | CurrentPOV() | O(1) | O(1) | No |
| sed.go | SedAll() | O(n*m) worst | O(n+m) | No — pre-allocated |
| self.go | Introduce() | O(1) | O(1) | No |
| current.go | GetCurrent() | O(1) atomic | O(1) | No |

### Agent Architecture

| Block | Hot Path | O(T) | O(N) | Unbounded? |
|-------|----------|------|------|------------|
| agent.go | SpawnAgent() | O(1) | O(1) | **Map grows unbounded** |
| swarm.go | SpawnSwarm() | O(n) agents | O(n) | No — max by port range |
| orchestrator.go | DelegatePrompt() | O(defs) | O(1) | No |
| ralph.go | Observe() | O(patterns) | O(patterns) | Capped at 64 cycles |
| supervisor.go | ReportFailure() | O(1) | O(children) | Per-supervisor |

### Protocols

| Block | Hot Path | O(T) | O(N) | Unbounded? |
|-------|----------|------|------|------------|
| a2a.go | RouteA2A() | O(handlers) | O(1) | No |
| a2c.go | Emit() | O(clients) | O(clients) | No — channel close |

### Topology

| Block | Hot Path | O(T) | O(N) | Unbounded? |
|-------|----------|------|------|------------|
| space.go | Distance() | O(V+E) BFS | O(V) | No |
| space.go | Reachable() | O(V+E) BFS | O(V) | No |
| dyad.go | Ping() | O(1) | O(1) | No |

### Learning

| Block | Hot Path | O(T) | O(N) | Unbounded? |
|-------|----------|------|------|------------|
| probe.go | ProbeAll() | O(caps) | O(1) | No |
| predict.go | PredictOutcome() | O(patterns) | O(1) | No |
| learn.go | Learn() | O(1) | O(lessons) | Capped at 256 |

## Unbounded Growth Risks

| Block | Risk | Mitigation |
|-------|------|------------|
| **agent.go** | AgentStore map grows without eviction | Add TTL-based cleanup for completed agents |
| **agent.go** | AgentCaches sync.Map for tool cache | Already per-agent, bounded by agent count |
| **brainstorm** | Sessions map never pruned | Add auto-complete on session end |
| **ralph.go** | Patterns map grows with new keywords | Cap at 256, LRU eviction |
| **learn.go** | Lessons slice capped at 256 — ✅ safe | |
| **log.go** | Ring buffer fixed 4096 — ✅ safe | |
| **journal.go** | Per-path depth capped at 16 — ✅ safe | |

## Recommendations

1. **AgentStore TTL**: Evict completed agents after 5 minutes
2. **Ralph patterns LRU**: Cap at 256 with eviction
3. **Brainstorm sessions GC**: Auto-complete sessions idle >1 hour
4. **Sed Boyer-Moore**: For patterns >3 chars, switch to BM for O(n/m) best case
5. **BLAKE3 tree mode**: For files >1MB, use BLAKE3 tree hashing for O(n/P) parallel
