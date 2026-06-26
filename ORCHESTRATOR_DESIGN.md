# Loop Orchestrator — Superpowers Design

## Capabilities baked into the model

### 1. Experiment Designer
Given a hypothesis, generates the complete training plan.

```
User: "Try GLM-5.2 as teacher with 200 pairs"
→ spawns: train_script.sh, data_labeling.py, eval config
→ estimates: $0.15 on RTX 4090, ~15 min
→ warns: "GLM teachers scored 0.951 (v13) — 0.004 below Qwen2.5"
```

### 2. Failure Diagnostician  
Given regression metrics, identifies root cause from known patterns.

```
Metrics: heretic=0.878, keep=0.878, data=983 pairs
→ diagnosis: "Data dilution (v15 pattern). 61% generic, 29% GLM, 10% C3.
  Fix: reduce to 300 pairs at 33% C3 ratio. Expected: +0.07 heretic."
```

### 3. Council Decision Engine
Reviews results, decides SHIP/RETRAIN/PIVOT with reasoning.

```
heretic=0.963, keep=0.963
→ "RETRAIN. +0.008 heretic over v8 but only 3.7% compression.
  Pareto analysis: λ=5 trades 11.3% compression for 0.008 precision.
  Not worth it. Stick with λ=3."
```

### 4. Sub-Agent Spawner
Delegates work to tool-calling sub-agents.

```
spawn_train("v18", config) → vast.ai instance
spawn_eval("PeetPedro/kompress-v8") → heretic results
spawn_label("Qwen2.5-7B", texts) → training pairs
spawn_deploy("kompress-v18") → HF model + card
```

### 5. Budget Tracker
Knows costs and tracks spend.

```
"v18 will cost $0.13. Budget remaining: $3.87. 26 runs possible."
```

### 6. State Keeper
Remembers the entire experiment history.

```
"17 versions trained. v8=production (0.955, 15% compression).
 11 dead ends documented. Pareto frontier mapped at λ=3,5,10."
```

## Tool definitions (function calling)

```json
{
  "tools": [
    {
      "name": "spawn_train",
      "description": "Launch a kompress training run on vast.ai",
      "parameters": {
        "version": "v18",
        "base_model": "chopratejas/kompress-v2-base",
        "data": "kompress_v8_train.jsonl",
        "epochs": 3,
        "lr": 2e-5,
        "weight": 3.0
      }
    },
    {
      "name": "spawn_eval", 
      "description": "Run heretic benchmark on a model",
      "parameters": {
        "model": "PeetPedro/kompress-v8",
        "prompts": "heretic_expanded.jsonl"
      }
    },
    {
      "name": "spawn_label",
      "description": "Label training data with a teacher model",
      "parameters": {
        "teacher": "Qwen2.5-7B-Instruct",
        "texts_file": "domain_train_2k.jsonl",
        "count": 200
      }
    },
    {
      "name": "check_status",
      "description": "Read experiment history and current state"
    },
    {
      "name": "council_review",
      "description": "Review metrics and decide next action",
      "parameters": {
        "heretic": 0.955,
        "keep_rate": 0.854,
        "version": "v8"
      }
    }
  ]
}
```

## Training data format

Each example is a multi-turn conversation where the orchestrator uses tools:

```json
{
  "messages": [
    {"role": "system", "content": "You are a loop engineering orchestrator..."},
    {"role": "user", "content": "My model regressed to 0.878 with 983 training pairs"},
    {"role": "assistant", "content": "Let me check the experiment history.", "tool_calls": [
      {"name": "check_status", "arguments": {}}
    ]},
    {"role": "tool", "content": "{\"versions\": 17, \"best\": \"v8=0.955\", \"v15\": \"0.878 with 983 pairs\"}"},
    {"role": "assistant", "content": "This matches the v15 data dilution pattern. Your 983 pairs are 61% generic — the signal is drowned. Reduce to 300 pairs at 33% C3 ratio. Want me to prepare that training run?"}
  ]
}
```

## Base model candidates

| Model | Params | Why |
|---|---|---|
| Qwen2.5-7B-Instruct | 7B | Proven teacher, solid function calling |
| Qwen3-Coder-30B-A3B | 30B MoE | Better coding/agent tasks |
| Llama-4-Scout-17B-16E | 17B MoE | Latest, strong agent benchmarks |

## Cost

~300-400 training pairs × $0.0005/pair (HF inference for labeling) = $0.20  
LoRA fine-tune on RTX 4090: ~$0.30  
Total: ~$0.50
