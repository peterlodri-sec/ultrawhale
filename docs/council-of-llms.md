# Council of LLMs — ultrawhale's Multi-Model Brain

> "Council of DeepSeek versions + OpenRouter FREE-only council + Copilot via GitHub CI"
> — Peter

## What Is This?

ultrawhale runs a COUNCIL of language models, not just one. Multiple models
collaborate, verify each other's outputs, and store collective wisdom in
dedicated long-term memory.

## The Council Members

| Council | Models | Cost | How |
|---------|--------|------|-----|
| **DeepSeek Council** | V4 Flash, V4 Pro, Coder V3 | Paid | Primary orchestrator + subagents |
| **OpenRouter FREE Council** | Gemma 3 4B, Mistral 7B, Llama 3.2 3B | **$0** | Dog Feed data collection + verification |
| **GitHub Copilot Council** | Copilot Chat via CI | Included in GitHub | Code review + PR suggestions |

## How It Works

```
Human prompt
    ↓
DeepSeek V4 Flash (orchestrator) — primary response
    ↓
DeepSeek V4 Pro (subagent) — deep reasoning
    ↓
OpenRouter FREE model (background) — second opinion, data collection
    ↓
GitHub Copilot (CI) — code review, suggestions
    ↓
All outputs → Dedicated Mem-Brain (long-term storage)
```

## Dedicated Mem-Brain

All council outputs are stored in **dedicated long-term memory**:
- DeepSeek responses → `brain/memos/ds-*`
- OpenRouter responses → `brain/memos/or-*` (Dog Feed dataset)
- Copilot suggestions → `brain/memos/gh-copilot-*`
- Council verdicts → `brain/memos/council-*`

## Disclaimer

⚠️ **ULTRA-RESEARCH-STATE**: The Council of LLMs is an experimental concept.
Multiple models may produce conflicting outputs. The council does not guarantee
correctness. It is a research tool for exploring multi-model collaboration.

⚠️ **FREE MODELS**: OpenRouter free models have rate limits and lower quality.
They are used for data collection and second opinions, not primary responses.

⚠️ **COST**: DeepSeek is paid. OpenRouter FREE models are $0. GitHub Copilot
is included in your GitHub subscription. Total cost per session: typically <$0.01.

## How to Use

```sh
# Start ultrawhale with council mode
ultrawhale --council deepseek+openrouter-free+copilot

# Query the council
/council ask "Should I use Rust or Go for this project?"
  → DeepSeek: "Rust for performance, Go for simplicity..."
  → Gemma (free): "Rust has better memory safety..."
  → Copilot: "Consider Go if team is familiar with it..."

# View council verdict
/council verdict
  → Consensus: Rust (2/3 recommend)
  → Stored to: brain/memos/council-20260621-001
```

## Vaked Fit

```
Council = Supervises layer (multi-agent coordination)
  + Testifies layer (verification by multiple models)
  + Indexes layer (dedicated mem-brain storage)
```
