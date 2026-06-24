# /// script
# requires-python = ">=3.11"
# dependencies = ["huggingface-hub>=0.23"]
# ///
"""
Upload local dogfeed JSONL files to HuggingFace — non-destructive, read-only.

Rules:
  - Never deletes or moves local files
  - Skips the most recently modified file (likely still being written)
  - Skips files already present on HF (checks by filename)
  - Skips empty files
  - Uploads one at a time with progress

Usage (run from the self-host-llm dir):
  uv run upload_local_dogfeed.py
  uv run upload_local_dogfeed.py --dir /path/to/dir
  uv run upload_local_dogfeed.py --dry-run
  uv run upload_local_dogfeed.py --active-grace 60   # skip files modified <60min ago
"""

import os, sys, argparse, time
from pathlib import Path
from datetime import datetime, timezone

from huggingface_hub import HfApi, CommitOperationAdd

HF_REPO  = "PeetPedro/ultrawhale-dogfood"
HF_TOKEN = os.environ.get("HF_TOKEN", "")

def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument("--dir", type=Path, default=Path("."),
                   help="Directory containing dogfeed_*.jsonl files (default: cwd)")
    p.add_argument("--dry-run", action="store_true",
                   help="Show what would be uploaded without pushing")
    p.add_argument("--active-grace", type=int, default=30,
                   help="Skip files modified within this many minutes (default: 30)")
    p.add_argument("--pattern", default="dogfeed_*.jsonl",
                   help="Glob pattern for local files (default: dogfeed_*.jsonl)")
    args = p.parse_args()

    if not args.dry_run and not HF_TOKEN:
        print("HF_TOKEN not set — use --dry-run or: export HF_TOKEN=hf_...", file=sys.stderr)
        sys.exit(1)

    target_dir = args.dir.expanduser().resolve()
    if not target_dir.is_dir():
        print(f"directory not found: {target_dir}", file=sys.stderr)
        sys.exit(1)

    # ── gather local files ──────────────────────────────────────────────────
    local_files = sorted(target_dir.glob(args.pattern))
    now = time.time()
    grace_secs = args.active_grace * 60

    eligible = []
    skipped_active = []
    skipped_empty = []

    for f in local_files:
        if f.stat().st_size == 0:
            skipped_empty.append(f.name)
            continue
        age_secs = now - f.stat().st_mtime
        if age_secs < grace_secs:
            skipped_active.append((f.name, int(age_secs // 60)))
        else:
            eligible.append(f)

    print(f"directory: {target_dir}")
    print(f"local files found: {len(local_files)}")
    if skipped_empty:
        print(f"  skip (empty): {', '.join(skipped_empty)}")
    if skipped_active:
        for name, age_min in skipped_active:
            print(f"  skip (active, {age_min}min old): {name}")
    print(f"eligible for upload: {len(eligible)}")

    if not eligible:
        print("nothing to upload.")
        return

    # ── fetch existing HF filenames ─────────────────────────────────────────
    print("\nchecking HF for already-uploaded files…")
    api = HfApi()
    try:
        hf_files = set(
            api.list_repo_files(HF_REPO, repo_type="dataset", token=HF_TOKEN)
        )
    except Exception as e:
        print(f"could not list HF files: {e}", file=sys.stderr)
        sys.exit(1)

    to_upload = [f for f in eligible if f.name not in hf_files]
    already   = [f for f in eligible if f.name in hf_files]

    if already:
        print(f"already on HF: {len(already)} files — skipping")
    print(f"to upload: {len(to_upload)} files")

    if not to_upload:
        print("all eligible files already on HF. done.")
        return

    # ── upload ──────────────────────────────────────────────────────────────
    for i, f in enumerate(to_upload, 1):
        size_kb = f.stat().st_size // 1024
        print(f"  [{i:03d}/{len(to_upload)}] {f.name}  ({size_kb} KB)", end="", flush=True)

        if args.dry_run:
            print("  [dry-run]")
            continue

        try:
            content = f.read_bytes()
            ts = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
            api.create_commit(
                repo_id=HF_REPO,
                repo_type="dataset",
                operations=[CommitOperationAdd(
                    path_in_repo=f.name,
                    path_or_fileobj=content,
                )],
                commit_message=f"upload: {f.name} (local loop) [{ts}]",
                token=HF_TOKEN,
            )
            print("  ✓")
        except Exception as e:
            print(f"  ERROR: {e}")

    mode = "dry-run" if args.dry_run else "done"
    print(f"\n{mode}: {len(to_upload)} files processed.")

if __name__ == "__main__":
    main()
