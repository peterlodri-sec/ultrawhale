---
license: mit
task_categories:
  - question-answering
  - text-generation
language:
  - en
tags:
  - ultrawhale
  - vaked
  - dogfeed
  - llm
  - fine-tuning
  - research
  - multi-model
  - council-of-llms
  - openrouter
  - deepseek
pretty_name: ultrawhale-dogfood
size_categories:
  - n<1K
---

# 🐕 ultrawhale-dogfood

A continuous dataset of human↔LLM interactions collected via ultrawhale's
**Dog Feed** loop. Designed for LLM fine-tuning, alignment research, and
multi-model comparison.

## 📊 Dataset Summary

| Property | Value |
|----------|-------|
| Samples | 60 (v3-enriched) |
| Fields | 13 |
| Topics | 20 (CS fundamentals + Vaked philosophy) |
| Models | DeepSeek V4 Flash (primary) + Gemma 3 4B (free) |
| Format | JSONL |
| License | MIT + CC-BY-4.0 |

## 📁 Files

| File | Samples | Description |
|------|---------|-------------|
| `dogfeed-v3-enriched.jsonl` | 60 | Full dataset with brain_context + memory_ref |
| `dogfeed-v2-science.jsonl` | 60 | Base dataset without enrichment |
| `dogfeed-v1-initial.jsonl` | 3 | Seed samples |
| `MANIFEST-v3.json` | — | VICE Genesis signed manifest |

## 🔧 Format

```json
{
  "id": "dogfeed-0001",
  "user_message": "what is recursion?",
  "free_response": "Recursion is when a function calls itself...",
  "free_model": "google/gemma-3-4b-it:free",
  "deepseek_response": "Recursion is a technique where...",
  "timestamp": "2026-06-21T12:00:00Z",
  "session_id": "v63.0.0",
  "topic": "recursion",
  "brain_context": {
    "pov": "M1/arm64/go",
    "session": "v63.0.0",
    "capabilities": "FULL"
  },
  "memory_ref": "a1b2c3d4e5f6",
  "pipeline": "dogfeed → brain → memory → dataset → HF"
}
```

## 🧠 Enrichment

v3 adds brain context from ultrawhale's memory system:
- `brain_context` — Point of View, session ID, capabilities
- `memory_ref` — SHA256 content-addressed reference
- `pipeline` — Full E2E trace of data provenance

## 🔒 PII & Trust

- All datasets are **zero-trust PII scrubbed** before publishing
- Every export is signed by the **VICE Genesis block**
- Trust score: **1.0000** · Verified: **true**
- Manifest hash: `0cf2f6944e2f865c`

## 🚀 Usage

```python
from datasets import load_dataset
dataset = load_dataset("PeetPedro/ultrawhale-dogfood", split="train")
for sample in dataset:
    print(sample["user_message"], "→", sample["free_response"][:50])
```

## 🏗️ Pipeline

```
Human types → SACRED surface → Brain short-term (32 turns)
  → Brain long-term → Dog Feed loop (free model)
  → Enrich (brain_context + memory_ref)
  → Export JSONL → vaked.dev → CI → HuggingFace
  → Webhook → Vaked liveness pulse → Telemetry Tree ring
```

## 📜 Citation

```bibtex
@dataset{ultrawhale-dogfood-2026,
  author = {Peter Lodri},
  title = {ultrawhale-dogfood: Human↔LLM Interaction Dataset},
  year = {2026},
  version = {v3-enriched},
  publisher = {HuggingFace},
  url = {https://huggingface.co/datasets/PeetPedro/ultrawhale-dogfood},
  note = {Collected via ultrawhale Dog Feed loop. VICE Genesis signed.}
}
```

## 🔗 Links

- **ultrawhale**: [github.com/peterlodri-sec/ultrawhale](https://github.com/peterlodri-sec/ultrawhale)
- **Live counter**: [vaked.dev/ultrawhale/dogfood](https://vaked.dev/ultrawhale/dogfood)
- **Docs**: [docs/council-of-llms.md](https://github.com/peterlodri-sec/ultrawhale/blob/main/docs/council-of-llms.md)

---

*⚠️ ULTRA-RESEARCH-STATE: This dataset is collected by experimental software.
Proceed with ultra-care. Peace 'n enjoy.*
