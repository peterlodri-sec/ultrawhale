# Prompts for Web LLMs — Generate Synthetic Training Data

> One-shot prompts you can paste into ChatGPT, Claude, Gemini, or any web LLM.
> Each prompt generates structured training data for the ultrawhale-dogfood dataset.
> The output feeds directly into `dogfeed-v4-web-llm.jsonl`.

---

## Prompt 1: Architecture Summary Generator

Paste this into any web LLM:

```
You are a synthetic data generator for the ultrawhale project — a recursive
self-improving coding agent built on the Vaked capability-graph philosophy.

Generate 5 JSONL entries about ultrawhale's architecture. Each entry must be
a valid JSON object on one line with exactly these fields:
  "user_message": a question someone might ask about ultrawhale
  "free_response": a concise answer about the architecture
  "topic": one of: "vaked", "recursion", "blocks", "protocols", "dyad"

Rules:
- Be accurate. The system has 148 blocks, 7 recursions, 14 protocols.
- Use real block names: sacred.go, promise.go, observer.go, space.go
- Reference real commands: /doctor, /promise, /observer, /deep
- The dyad connects M1 (lodris-macbook-pro) and M3 (m3-macbook)

Output ONLY 5 JSONL lines. No explanation. No markdown. Just JSONL.
```

**Expected output:**
```jsonl
{"user_message":"What is the Vaked pipeline?","free_response":"The Vaked pipeline has 7 layers: Declares -> Materializes -> Supervises -> Enforces -> Testifies -> Indexes -> Reveals. It is a strange loop.","topic":"vaked"}
{"user_message":"How many blocks does ultrawhale have?","free_response":"148 content-addressed blocks. Each block has a Status() function and a VakedFit() declaration.","topic":"blocks"}
// ... 3 more
```

---

## Prompt 2: Concept Explainer (Deep Dive)

Paste this into any web LLM:

```
You are generating training data for a self-improving coding agent.

Generate 8 JSONL entries explaining advanced concepts. Each line must be
a valid JSON object with:
  "user_message": "Explain [concept] in simple terms"
  "free_response": a 2-3 sentence explanation
  "topic": the category

Concepts to cover (pick 5-8 from this list):
- SPACE+TIME PROOF
- SEALING (10% reserve)
- The Observer block
- The Honesty Loop
- SHEET (Sacred Hypertext Element Engine)
- STR (string templates)
- OSCE protocol
- FORWARD INJECT vs BACKWARD EXPAND
- The Ralph Loop
- DogFeed pipeline
- Strange loops in software
- Gödel machine implementation

Rules:
- Keep responses under 200 characters each
- Be technically accurate
- Use real file paths where relevant (e.g., internal/blocks/observer.go)
- Output ONLY valid JSONL, one object per line
```

**Expected output:**
```jsonl
{"user_message":"Explain SPACE+TIME PROOF","free_response":"A cryptographic proof of recording. SPACE (machine+arch) + TIME (Lamport+UTC) + WATERMARK. Like a bodycam for code.","topic":"proof"}
{"user_message":"Explain the Observer block","free_response":"The Observer (internal/blocks/observer.go) watches the recursive feeding loop from outside. It is the strange loop made executable.","topic":"recursion"}
```

---

## How to Use

1. Copy either prompt into any web LLM
2. Run it 3-5 times to generate 15-40 JSONL entries
3. Save the output to `~/.ultrawhale/dogfeed/dogfeed-v4-web-llm.jsonl`
4. The next HF auto-push will include it

```bash
# Append generated data
cat ~/Downloads/output.jsonl >> ~/.ultrawhale/dogfeed/dogfeed-v4-web-llm.jsonl
```

Each run generates fresh data. The dataset grows with every prompt.
