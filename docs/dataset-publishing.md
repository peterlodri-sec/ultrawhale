# Dataset Publishing — HuggingFace + Git LFS

## Current Datasets

| Dataset | Size | Format | License |
|---------|------|--------|---------|
| DogFood v1 | Growing | JSONL | MIT + CC-BY-4.0 |
| DogFood v2 | Planned | JSONL + Parquet | MIT + CC-BY-4.0 |
| Council Verdicts | Planned | JSONL | MIT + CC-BY-4.0 |

## Publishing Paths

### Primary: HuggingFace Datasets

```
huggingface.co/peterlodri-sec/ultrawhale-dogfood (create: hf repo create)
```

```python
# Load directly from HuggingFace
from datasets import load_dataset
dataset = load_dataset("peterlodri-sec/ultrawhale-dogfood")
```

### Mirror: GitHub LFS

```sh
# Track large dataset files with Git LFS
git lfs track "*.jsonl" "*.parquet" "*.arrow"
git lfs push origin main
```

### Mirror: vaked.dev

```
vaked.dev/ultrawhale/dogfood/          # Live counter + info
vaked.dev/ultrawhale/dogfood-v1.jsonl  # Direct download
```

## Setup (Manual Steps)

1. **HuggingFace**:
   - Create account at huggingface.co
   - `hf login`
   - `hf repo create ultrawhale-dogfood --type dataset`
   - Push: `cp ~/.ultrawhale/dogfeed/*.jsonl . && git add . && git commit -m "v1" && git push`

2. **Git LFS**:
   - `brew install git-lfs && git lfs install`
   - `git lfs track "*.jsonl"`
   - Push datasets as LFS objects

3. **vaked.dev**:
   - Already deployed: `vaked.dev/ultrawhale/dogfood/`
   - Direct download link: add to Taskfile dev-deploy

## Auto-Publish (v65)

```sh
# One command to publish all datasets
/dog-feed export          # exports JSONL
/dataset publish          # pushes to HuggingFace + GitHub LFS + vaked.dev
```

## PII Guarantee

All datasets are zero-trust scrubbed before publishing.
See `.dev/scrub-pii.sh` and `.dev/DATASET-MANIFEST.md`.
No API keys. No emails. No private IPs.
