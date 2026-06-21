# Brain → Memory → Files → Dataset → HuggingFace — E2E Pipeline

## The Loop

```
Human types → SACRED surface
    ↓
Brain short-term memory (32 turns)
    ↓
Brain long-term memory (JSONL append)
    ↓
Dog Feed loop (background, free model)
    ↓
DogFeed samples (user, free_response, deepseek_response)
    ↓
Enrich with brain_context + memory_ref
    ↓
Export JSONL → ~/.ultrawhale/dogfeed/
    ↓
task dev-deploy → vaked.dev/ultrawhale/dogfood/
    ↓
CI auto-publish → HuggingFace datasets
    ↓
HF webhook → Vaked liveness pulse
    ↓
Telemetry Tree ring grows
```

## Dataset Versions

| Version | Samples | Fields | Enriched? |
|---------|---------|--------|-----------|
| v1-initial | 3 | 6 | ❌ |
| v2-science | 60 | 9 | ❌ |
| v3-enriched | 60 | 13 | ✅ brain_context, memory_ref, pipeline |

## Enrichment Fields

| Field | Purpose |
|-------|---------|
| `brain_context` | POV, session, capabilities, space node |
| `memory_ref` | SHA256 hash of user message (content-addressed) |
| `enriched_at` | ISO 8601 timestamp of enrichment |
| `pipeline` | Full trace: dogfeed → brain → memory → dataset → HF |

## Hardening

- All datasets PII-scrubbed before enrichment
- Memory refs are content-addressed (SHA256)
- Brain context is immutable per session
- Pipeline trace is append-only
- VICE Genesis block signs every export manifest

## Usage

```python
from datasets import load_dataset
dataset = load_dataset("PeetPedro/ultrawhale-dogfood")
# Enriched samples have brain_context and memory_ref fields
```
