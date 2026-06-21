# ultrawhale DogFood Dataset

Public dataset of human↔LLM interactions collected via the Dog Feed loop.

## What it is

A continuous dataset of (user_message, free_model_response, deepseek_response) triples.
Collected in the background by ultrawhale's Dog Feed primitive.
Designed for LLM fine-tuning and alignment research.

## Format

JSONL — one JSON object per line:

```json
{
  "user_message": "explain async rust",
  "free_response": "[dogfeed:gemma] async rust is...",
  "free_model": "google/gemma-3-4b-it:free",
  "deepseek_response": "Async Rust uses...",
  "timestamp": "2026-06-21T12:00:00Z",
  "session_id": "v52.0.0"
}
```

## Usage

```python
import json

# Load dataset
samples = []
with open("dogfeed-20260621.jsonl") as f:
    for line in f:
        samples.append(json.loads(line))

# Fine-tune on user→free_response pairs
train_data = [{"messages": [
    {"role": "user", "content": s["user_message"]},
    {"role": "assistant", "content": s["free_response"]}
]} for s in samples]
```

## License

MIT + CC-BY-4.0 dual license. Free for research, commercial use, fine-tuning.
Attribution: "ultrawhale DogFood Dataset — github.com/peterlodri-sec/ultrawhale"

## Live Counter

See [vaked.dev/ultrawhale/dogfood](https://vaked.dev/ultrawhale/dogfood) for live sample count.
