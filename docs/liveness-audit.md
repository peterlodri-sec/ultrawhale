# Liveness Audit — v12.4.0

Liveness = "is the system still alive?" — distinct from safety and security.

## Layers Audited (10)

| Layer | Liveness Status | Gaps Found |
|-------|----------------|------------|
| Blocks engine | ✅ Write falls back, journal rollback works, BLAKE3 fallback active | 0 |
| Agent | ✅ Supervisor restarts (3 retries, exponential backoff) | 0 |
| Orchestrator | ✅ Fallback to explore agent if classify fails | 1 (fixed) |
| A2A | ✅ Dead agent check before send | 1 (fixed) |
| A2C | ✅ Channel close on disconnect | 0 |
| Dyad | ✅ Failover at 30s, health endpoint | 1 (fixed) |
| Surface | ✅ /api/v1/health endpoint | 0 |
| AGUI | ✅ Bubble Tea handles panic recovery | 0 |
| Pre-hooks | ✅ 10s timeout per hook | 1 (fixed) |
| Ralph | ✅ Manual /ralph rollback override | 0 |

## Gaps Fixed (v12.4.0)

| Gap | Severity | Fix |
|-----|----------|-----|
| Orchestrator no fallback | High | classifyByCapability falls back to explore agent |
| A2A to dead agent | High | Check agent existence before send, return error |
| Dyad health opaque | Medium | DyadHealth() returns full status map |
| Pre-hook indefinite block | Medium | 10s timeout per Validate() call |

## Remaining (low priority)

| Gap | Why not critical |
|-----|-----------------|
| TUI widget crash | Bubble Tea recovers via panic handler |
| Ring buffer overflow | 4096 slots, oldest evicted silently — acceptable |
| NATS down during dyad | Dyad degrades to "degraded" mode, self continues |
| SSE client heartbeat | Channel close on disconnect is sufficient for now |
