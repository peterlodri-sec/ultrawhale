# Kompress v3 Fine-Tune Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fine-tune `chopratejas/kompress-v2-base` on ultrawhale Q&A pairs using LoRA r=16, produce a kompress-v3 ONNX artifact that achieves keep_rate < 0.75 while holding exact_keep_pct (must-keep tokens) > 0.95.

**Architecture:** LoRA r=16 on last 4 ModernBERT attention layers (q/k/v). Encoder frozen. Token head and span CNN re-trained. Silver labels from ultrawhale `deepseek_response` / `free_response` pairs — tokens in compressed version get label=1, must-keep patterns (numbers, identifiers, paths, flags) always label=1 with weight 3.0.

**Tech Stack:** PyTorch 2.3, transformers, peft (LoRA), vast.ai RTX 4090, HuggingFace Hub, ONNX opset 17.

## Global Constraints

- Base model: `chopratejas/kompress-v2-base` (merged.pt format, not safetensors)
- Base encoder: `answerdotai/ModernBERT-base` (22 layers, 768-dim hidden)
- LoRA target: layers 18–21 (0-indexed), modules: `query`, `key`, `value`
- ONNX output names: `input_ids`, `attention_mask` → `final_scores` (float32, [0,1])
- HF repo: `peterlodri-sec/kompress-v3`
- Budget: $6–7 on vast.ai (RTX 4090 ~$0.356/hr, full run ~$0.18)
- No Co-Authored-By trailers in commits (project convention)

## File Map (all already implemented)

```
ultrawhale/
  scripts/
    export_for_kompress.py   # pull ultrawhale JSONL from HF, produce training pairs
    train_kompress.py        # LoRA fine-tune with weighted BCE
    eval_kompress.py         # keep_rate / sem_sim / exact_pct metrics
    run_training.sh          # orchestrates 1→2→3→4 on vast.ai instance
    vast_setup.sh            # pip install + clone + pre-cache models
  notebooks/
    kompress_finetune.ipynb  # Colab quick-start + vast.ai production path
  Taskfile.yml               # kompress-data / kompress-vast-launch / kompress-eval targets

pocoo.vaked.dev/
  posts/
    2026-06-25-fine-tuning-kompress-sapir-whorf.md   # published, results TBD
```

---

### Task 1: Smoke-test data export locally

**Files:**
- Run: `scripts/export_for_kompress.py`
- Output: `data/kompress_train.jsonl`

- [ ] **Step 1: Export data**

```bash
cd /Users/lodripeter/workspace/peterlodri-sec/ultrawhale
task kompress-data
```

Expected: `data/kompress_train.jsonl` written, terminal shows record count.

- [ ] **Step 2: Verify output format**

```bash
head -3 data/kompress_train.jsonl | python3 -c "
import json, sys
for line in sys.stdin:
    r = json.loads(line)
    print('keys:', list(r.keys()))
    print('text len:', len(r['text']))
    print('ref len:', len(r['reference']))
    break
"
```

Expected output:
```
keys: ['text', 'reference', 'role', 'source']
text len: <200-600>
ref len: <50-200>
```

- [ ] **Step 3: Count records**

```bash
wc -l data/kompress_train.jsonl
```

Expected: 1800–2500 lines (ultrawhale v1+v2+v3).

- [ ] **Step 4: Commit data manifest**

```bash
echo "data/kompress_train*.jsonl" >> .gitignore
git add .gitignore
git commit -m "chore: ignore generated training data files"
```

---

### Task 2: Smoke-test training locally (1 epoch, CPU)

**Files:**
- Run: `scripts/train_kompress.py`
- Output: `kompress-v3-local-test/`

- [ ] **Step 1: Run 1-epoch smoke test**

```bash
task kompress-train-local
```

This runs:
```bash
uv run python scripts/train_kompress.py \
    --data data/kompress_train.jsonl \
    --epochs 1 \
    --batch-size 8 \
    --output kompress-v3-local-test
```

Expected: completes without error, `kompress-v3-local-test/` directory created with `adapter_config.json`, `adapter_model.safetensors`.

- [ ] **Step 2: Verify checkpoint structure**

```bash
ls kompress-v3-local-test/
```

Expected files:
```
adapter_config.json
adapter_model.safetensors
tokenizer_config.json
tokenizer.json
special_tokens_map.json
```

- [ ] **Step 3: Run quick eval on local checkpoint**

```bash
# Split data first if not already done
python3 - << 'PY'
import json, random, pathlib
records = [json.loads(l) for l in open("data/kompress_train.jsonl")]
random.seed(42)
random.shuffle(records)
split = int(len(records) * 0.9)
pathlib.Path("data").mkdir(exist_ok=True)
with open("data/kompress_test.jsonl","w") as f:
    for r in records[split:]: f.write(json.dumps(r)+"\n")
print(f"Test split: {len(records)-split} records")
PY

task kompress-eval MODEL=kompress-v3-local-test
```

Expected: prints keep_rate, sem_sim, exact_pct. 1-epoch local values will be worse than baseline — that's fine, just confirms the pipeline runs.

- [ ] **Step 4: Clean up local test checkpoint**

```bash
rm -rf kompress-v3-local-test
```

---

### Task 3: Launch vast.ai training run

**Files:**
- Run: `Taskfile.yml` → `kompress-vast-launch`
- Requires: `vastai` CLI installed, `HF_TOKEN` env var set

- [ ] **Step 1: Verify vastai CLI**

```bash
vastai --version
```

If not installed:
```bash
pip install vastai
vastai apikey <YOUR_API_KEY>
```

- [ ] **Step 2: Search for a cheap RTX 4090 offer**

```bash
vastai search offers 'gpu_name=RTX_4090 num_gpus=1 disk_space>=30 inet_up>=500' \
    --order dph_total --limit 5
```

Note the cheapest `ID` from the output. The Taskfile default is `37031007` — verify it is still available:

```bash
vastai show instances
# or just use the search result ID
```

- [ ] **Step 3: Export HF token**

```bash
export HF_TOKEN=$(cat ~/.huggingface/token 2>/dev/null || echo "")
# verify
echo $HF_TOKEN | head -c 10
```

- [ ] **Step 4: Launch instance**

```bash
task kompress-vast-launch OFFER_ID=<ID_FROM_STEP_2>
```

This runs:
```bash
vastai create instance <OFFER_ID> \
    --image pytorch/pytorch:2.3.0-cuda12.1-cudnn8-runtime \
    --disk 30 \
    --env "-e HF_TOKEN=$HF_TOKEN -e HF_REPO=peterlodri-sec/kompress-v3" \
    --onstart "bash /workspace/ultrawhale/scripts/vast_setup.sh && \
               bash /workspace/ultrawhale/scripts/run_training.sh"
```

Expected: instance ID printed. Training starts automatically via `--onstart`.

- [ ] **Step 5: Monitor**

```bash
vastai logs <INSTANCE_ID>
```

Watch for the 4 stage markers:
```
=== 1/4 Export training data ===
=== 2/4 Fine-tune ===
=== 3/4 Eval ===
=== 4/4 ONNX export + upload ===
=== Done. GPU time used: check vast.ai billing ===
```

Full run expected: ~30 min. Cost: ~$0.18.

- [ ] **Step 6: Destroy instance when done**

```bash
vastai destroy instance <INSTANCE_ID>
```

---

### Task 4: Collect results and update blog post

**Files:**
- Modify: `pocoo.vaked.dev/posts/2026-06-25-fine-tuning-kompress-sapir-whorf.md`

- [ ] **Step 1: Pull eval results from logs**

```bash
vastai logs <INSTANCE_ID> | grep -A 20 "=== 3/4 Eval"
```

Note the three metrics:
- `keep_rate`: target < 0.75
- `sem_sim`: target > 0.90
- `exact_pct`: target > 0.95

- [ ] **Step 2: Verify HF upload**

```bash
python3 -c "
from huggingface_hub import list_repo_files
files = list(list_repo_files('peterlodri-sec/kompress-v3'))
print(files)
"
```

Expected: `merged.pt` or `adapter_model.safetensors` + `onnx/kompress-fp32.onnx` present.

- [ ] **Step 3: Update blog post results section**

In `pocoo.vaked.dev/posts/2026-06-25-fine-tuning-kompress-sapir-whorf.md`, find the evaluation section and fill in the actual numbers. The section currently says:

> Current v2 metrics: F1=0.913, keep_rate=0.810.
> We're targeting: keep_rate < 0.75 at exact_keep_pct > 0.95.

Add after it:

```markdown
**v3 results (ultrawhale fine-tune):**

| Metric | v2 baseline | v3 fine-tuned |
|--------|------------|---------------|
| keep_rate | 0.810 | <ACTUAL> |
| exact_keep_pct | — | <ACTUAL> |
| sem_sim | — | <ACTUAL> |
| training cost | — | $<ACTUAL> |
```

- [ ] **Step 4: Commit and push blog update**

```bash
cd /Users/lodripeter/workspace/peterlodri-sec/pocoo.vaked.dev
git add posts/2026-06-25-fine-tuning-kompress-sapir-whorf.md
git commit -m "post: kompress v3 results — keep_rate <ACTUAL>, exact_keep_pct <ACTUAL>"
git push
```

---

### Task 5: Verify ONNX artifact against headroom

**Files:**
- Read: `headroom/headroom/transforms/kompress_compressor.py`
- Run: point `HEADROOM_KOMPRESS_BACKEND=pytorch` or `onnx` at the new model

- [ ] **Step 1: Download ONNX from HuggingFace**

```bash
python3 - << 'PY'
from huggingface_hub import hf_hub_download
path = hf_hub_download("peterlodri-sec/kompress-v3", "onnx/kompress-fp32.onnx")
print(f"Downloaded to: {path}")
PY
```

- [ ] **Step 2: Smoke-compress a tool output**

```bash
cd /Users/lodripeter/workspace/peterlodri-sec/headroom
HEADROOM_KOMPRESS_BACKEND=pytorch python3 - << 'PY'
import os, sys
os.environ["HEADROOM_KOMPRESS_MODEL_ID"] = "peterlodri-sec/kompress-v3"
from headroom.transforms.kompress_compressor import kompress_compress
sample = """
Error: SIGILL at 0x7fff2038 in module libsystem_kernel.dylib
Thread 0: EXC_BAD_INSTRUCTION (SIGILL)
Frame 0: libsystem_kernel + 0x2038
Frame 1: libdispatch + 0x1234
Stack trace available. Memory usage: 4.2GB RSS.
"""
result = kompress_compress(sample)
print(f"Original: {len(sample.split())} tokens")
print(f"Compressed: {len(result.compressed.split())} tokens")
print(f"Ratio: {result.compression_ratio:.2f}")
print(f"Output: {result.compressed[:200]}")
PY
```

Expected: SIGILL survives (must-keep pattern), ratio > 1.2.

- [ ] **Step 3: Commit plan as done**

```bash
cd /Users/lodripeter/workspace/peterlodri-sec/ultrawhale
git add docs/superpowers/plans/2026-06-25-kompress-finetune.md
git commit -m "docs: kompress v3 implementation plan"
```
