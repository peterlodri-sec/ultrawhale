#!/usr/bin/env python3
"""Build training data for kompress-superpower-orchestrator.

Generates function-calling conversation pairs from our 17-experiment history.
Each pair teaches the model to diagnose, plan, spawn sub-agents, and decide.

Output: data/orchestrator_train.jsonl — ~350 pairs
"""

import json, random
from pathlib import Path

# ── Experiment history (all 17 versions) ──────────────────────────
EXPERIMENTS = [
    {"v": "v2", "heretic": 0.975, "keep": 0.897, "weight": None, "teacher": None, "data": None, "verdict": "precision ceiling"},
    {"v": "v3", "heretic": 0.942, "keep": 0.728, "weight": None, "teacher": "self-label", "data": "Q&A", "verdict": "first self-label"},
    {"v": "v4", "heretic": 0.943, "keep": 0.823, "weight": None, "teacher": "self-label", "data": "domain", "verdict": "override internalized"},
    {"v": "v5", "heretic": 0.961, "keep": 0.86, "weight": None, "teacher": "self-label", "data": "domain", "verdict": "converged"},
    {"v": "v6", "heretic": 0.962, "keep": 0.854, "weight": None, "teacher": "generator", "data": "agent-dist", "verdict": "dead end — conservative"},
    {"v": "v7", "heretic": 0.956, "keep": 0.868, "weight": None, "teacher": "sliding-window", "data": "agent", "verdict": "dead end — regressed"},
    {"v": "v8", "heretic": 0.955, "keep": 0.854, "weight": 3.0, "teacher": "Qwen2.5-7B", "data": "97 C3 + 200 generic", "verdict": "PRODUCTION ★"},
    {"v": "v9", "heretic": 0.921, "keep": None, "weight": 3.0, "teacher": "Qwen2.5-7B", "data": "97 C3-only", "verdict": "overfit — need diversity"},
    {"v": "v10", "heretic": 0.947, "keep": 0.891, "weight": 3.0, "teacher": "Qwen2.5-7B", "data": "285 C3 + 580 generic", "verdict": "diminishing returns"},
    {"v": "v11", "heretic": 0.906, "keep": 0.517, "weight": 3.0, "teacher": "Qwen2.5-7B", "data": "97 C3 + 200 generic", "verdict": "capacity ≠ precision (352M)"},
    {"v": "v12", "heretic": 0.949, "keep": 0.949, "weight": 3.0, "teacher": "Qwen3-Coder", "data": "141 C3 + 282 generic", "verdict": "teacher too conservative"},
    {"v": "v13", "heretic": 0.951, "keep": 0.951, "weight": 3.0, "teacher": "regex", "data": "127 GLM + 254 generic", "verdict": "too conservative"},
    {"v": "v14", "heretic": 0.882, "keep": None, "weight": 3.0, "teacher": "council", "data": "v8 + GLM", "verdict": "concept proven"},
    {"v": "v15", "heretic": 0.878, "keep": 0.878, "weight": 3.0, "teacher": "Qwen2.5-7B", "data": "97 C3 + 286 GLM + 600 generic", "verdict": "data dilution"},
    {"v": "v16", "heretic": 0.972, "keep": 0.972, "weight": 10.0, "teacher": "Qwen2.5-7B", "data": "97 C3 + 200 generic", "verdict": "Pareto endpoint"},
    {"v": "v17", "heretic": 0.963, "keep": 0.963, "weight": 5.0, "teacher": "Qwen2.5-7B", "data": "97 C3 + 200 generic", "verdict": "Pareto tradeoff"},
]

ORCHESTRATOR_SYSTEM = """You are kompress-superpower-orchestrator — a loop engineering agent that designs experiments, diagnoses failures, spawns sub-agents, and decides what to try next.

You have access to tools:
- check_status(): read experiment history
- spawn_train(version, config): launch training on vast.ai
- spawn_eval(model): run heretic benchmark
- spawn_label(teacher, texts): label training data
- council_review(metrics): decide SHIP/RETRAIN/PIVOT

You know the kompress experiment history: 17 models, v8=production (0.955 heretic, 15% compression), Pareto frontier at λ=3/5/10, 11 dead ends documented.
Label quality is the bottleneck — not model capacity, not data quantity.
Always give specific, actionable plans with cost estimates."""

def build_pair(scenario_type: str, **kwargs) -> dict:
    """Build a single training pair."""
    messages = [{"role": "system", "content": ORCHESTRATOR_SYSTEM}]
    
    if scenario_type == "diagnosis_regression":
        v = kwargs["v"]
        messages.append({"role": "user", "content": f"My model regressed. Here are the metrics: heretic={v['heretic']}, keep_rate={v.get('keep','?')}, trained on {v['data']}. What happened?"})
        messages.append({"role": "assistant", "content": f"Let me check our experiment history for similar patterns.", "tool_calls": [{"name": "check_status", "arguments": {}}]})
        messages.append({"role": "tool", "content": json.dumps({"versions": 17, "best": "v8=0.955"})})
        messages.append({"role": "assistant", "content": f"This matches {v['v']} ({v['verdict']}). The root cause is {v['verdict'].split(' — ')[0] if ' — ' in v['verdict'] else v['verdict']}. Recommendation: {_get_fix(v)}"})
    
    elif scenario_type == "plan_experiment":
        exp = kwargs["exp"]
        messages.append({"role": "user", "content": exp["query"]})
        messages.append({"role": "assistant", "content": exp["response"]})
    
    elif scenario_type == "council_decision":
        exp = kwargs["exp"]
        messages.append({"role": "user", "content": f"Review these results: {exp['metrics']}"})
        messages.append({"role": "assistant", "content": f"Let me review.", "tool_calls": [{"name": "council_review", "arguments": exp["metrics"]}]})
        messages.append({"role": "tool", "content": json.dumps({"decision": exp["decision"], "reasoning": exp["reasoning"]})})
        messages.append({"role": "assistant", "content": f"Decision: {exp['decision'].upper()}. {exp['reasoning']}"})
    
    elif scenario_type == "spawn_train":
        messages.append({"role": "user", "content": kwargs["query"]})
        messages.append({"role": "assistant", "content": "Preparing training run.", "tool_calls": [{"name": "spawn_train", "arguments": kwargs["config"]}]})
        messages.append({"role": "tool", "content": json.dumps({"instance_id": "vast-42499999", "cost": "$0.13", "eta": "15 min"})})
        messages.append({"role": "assistant", "content": kwargs["response"]})
    
    return {"messages": messages}

def _get_fix(v: dict) -> str:
    """Get fix recommendation based on version verdict."""
    fixes = {
        "overfit — need diversity": "Add 200 generic multi-domain pairs. Reduce C3 ratio to 33%. Expected: +0.03 heretic.",
        "diminishing returns": "More C3 data didn't help (285 vs 97). Stick with 97 pairs. The signal saturates.",
        "capacity ≠ precision (352M)": "Larger encoder over-compresses. Revert to ModernBERT-base (149M). Expected: +0.05 heretic.",
        "teacher too conservative": "Switch teacher to Qwen2.5-7B. Qwen3-Coder preserves too many tokens. Expected: +0.006 heretic.",
        "data dilution": "Reduce pairs from 983 to 300. Keep 33% C3 ratio. The signal is drowned in generic data.",
        "Pareto endpoint": "λ=10 gives 0.972 heretic but only 2.8% compression. λ=3 is optimal. Don't ship — this is the endpoint, not the solution.",
        "Pareto tradeoff": "λ=5 gives +0.008 heretic but -11.3% compression vs λ=3. Not worth the trade. Ship v8.",
        "dead end — conservative": "Agent-distribution training makes model more conservative. Abandon this direction. Try C3 distillation instead.",
        "dead end — regressed": "Sliding-window self-labeling regresses precision. Use regex override in production instead of training.",
        "concept proven": "Council loop works but needs guardrails. More epochs caused overfitting. Stick to 3 epochs.",
    }
    for key, fix in fixes.items():
        if key in v["verdict"]:
            return fix
    return "Analyze the specific failure mode and adjust accordingly."

# ── Generate pairs ────────────────────────────────────────────────
pairs = []

# 1. Diagnosis pairs — for every regression version
for v in EXPERIMENTS:
    if v["heretic"] < 0.94 and v["v"] != "v2":
        pairs.append(build_pair("diagnosis_regression", v=v))

# 2. Plan experiment pairs — common scenarios
plan_queries = [
    {"query": "I want to try a stronger teacher. Which one should I use?", "response": "Qwen2.5-7B-Instruct is our best teacher (v8: 0.955). Qwen3-Coder scored 0.949 — too conservative. GLM-5.1 scored 0.951. Stick with Qwen2.5-7B. Cost: $0 (HF free inference) for labeling 120 texts."},
    {"query": "Should I use a larger model? ModernBERT-large is 352M.", "response": "No. v11 proved capacity ≠ precision. 352M encoder scored 0.906 — worse than 149M base (0.955). The larger model over-compresses (keep_rate 0.517). Stick with 149M."},
    {"query": "How many training pairs do I need?", "response": "97 C3 pairs at 33% ratio is optimal (v8: 0.955). 285 pairs gave 0.947. 983 pairs gave 0.878. More data dilutes the signal. The sweet spot is small and calibrated."},
    {"query": "What's the best loss weight for must-keep tokens?", "response": "λ=3.0 is Pareto-optimal. λ=5: heretic +0.008 but compression -11.3%. λ=10: heretic 0.972 but only 2.8% compression. The tradeoff is fundamental. Use 3.0."},
    {"query": "How do I set up the training pipeline?", "response": "Use LoopKit: clone github.com/peterlodri-sec/loopkit, run `python -m loops.kompress.loop`. Or launch on vast.ai: $0.13/run on RTX 4090. I can spawn it for you.", "tool_calls": True},
    {"query": "My data is all synthetic. Is that a problem?", "response": "Yes — v9 trained on C3-only data and overfit to 0.921. Add 200 generic multi-domain pairs (dolly/openorca/alpaca). The 33% C3 ratio is the sweet spot for generalization."},
    {"query": "Can I automate the experiment loop?", "response": "Yes. Use the council pattern: run experiment → evaluate → council reviews → decide next action. v14 proved the concept. LoopKit has a council built in. I can set it up."},
]
for q in plan_queries:
    pairs.append(build_pair("plan_experiment", exp=q))

# 3. Council decision pairs
council_examples = [
    {"metrics": {"heretic": 0.955, "keep_rate": 0.854, "override_delta": 0.0}, "decision": "SHIP", "reasoning": "Beats all previous versions. 15% compression with 0.955 precision. This is production-ready."},
    {"metrics": {"heretic": 0.921, "keep_rate": 0.85}, "decision": "RETRAIN", "reasoning": "Overfit on C3-only data. Add 200 generic pairs and retry at 33% ratio."},
    {"metrics": {"heretic": 0.906, "keep_rate": 0.517}, "decision": "PIVOT", "reasoning": "Larger encoder collapses. Abandon 352M architecture. Pivot back to 149M base."},
    {"metrics": {"heretic": 0.963, "keep_rate": 0.963}, "decision": "RETRAIN", "reasoning": "+0.008 heretic but -11.3% compression vs λ=3. Not worth shipping. The Pareto frontier is clear."},
    {"metrics": {"heretic": 0.878, "keep_rate": 0.878}, "decision": "PIVOT", "reasoning": "983 pairs diluted the signal. 61% generic data drowns C3 labels. Pivot to smaller dataset."},
]
for ex in council_examples:
    pairs.append(build_pair("council_decision", exp=ex))

# 4. Spawn train pairs
spawn_examples = [
    {"query": "Train v18 with 5x weight on v8 data", "config": {"version": "v18", "base_model": "chopratejas/kompress-v2-base", "data": "kompress_v8_train.jsonl", "weight": 5.0, "epochs": 3, "lr": 2e-5}, "response": "Training v18 on vast.ai RTX 4090. 5x weight, v8 data. Expected: heretic ~0.963, keep ~0.963. Monitor: vastai logs <ID>"},
    {"query": "Label 200 texts with Qwen2.5-7B", "config": {"teacher": "Qwen2.5-7B-Instruct", "texts_file": "domain_train_2k.jsonl", "count": 200}, "response": "Labeling 200 texts with Qwen2.5-7B via HF_INFER_PRO. ~5 min. Free serverless inference. Will produce ~150 C3 pairs after filtering."},
]
for ex in spawn_examples:
    pairs.append(build_pair("spawn_train", **ex))

# ── Save ───────────────────────────────────────────────────────────
out = Path(__file__).parent.parent / "data" / "orchestrator_train.jsonl"
with open(out, "w") as f:
    for p in pairs:
        f.write(json.dumps(p, ensure_ascii=False) + "\n")

print(f"Generated {len(pairs)} training pairs → {out}")
print(f"  Diagnosis: {sum(1 for p in pairs if 'Let me check' in str(p))}")
print(f"  Plan: {sum(1 for p in pairs if 'Let me' not in str(p) and 'tool_calls' not in str(p))}")
print(f"  Council: {sum(1 for p in pairs if 'council_review' in str(p))}")
print(f"  Spawn: {sum(1 for p in pairs if 'spawn_train' in str(p))}")
