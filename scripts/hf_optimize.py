# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "pandas>=2.0",
#   "pyarrow>=15.0",
#   "huggingface-hub>=0.23",
# ]
# ///
"""
HF dataset optimizer — merge JSONL → Parquet, push to HuggingFace.

Usage:
  uv run scripts/hf_optimize.py                          # merge + push both streams
  uv run scripts/hf_optimize.py --dry-run                # merge locally only
  uv run scripts/hf_optimize.py --stream dogfeed         # only dogfeed stream
  uv run scripts/hf_optimize.py --stream telemetry       # only telemetry stream

CI mode (write files to HF clone dir, let CI handle git push):
  python3 scripts/hf_optimize.py --output-dir /tmp/hf-push \\
    --telemetry-path /tmp/hf-push/telemetry-consolidated.jsonl
"""

import os, sys, json, argparse, hashlib, tempfile
from pathlib import Path
from datetime import datetime, timezone

import pandas as pd
import pyarrow as pa
import pyarrow.parquet as pq
from huggingface_hub import HfApi, CommitOperationAdd

HF_REPO     = "PeetPedro/ultrawhale-dogfood"
HF_TOKEN    = os.environ.get("HF_TOKEN", "")
DOGFEED_DIR = Path(__file__).parent.parent / "site" / "dogfood"

def read_jsonl(path: Path) -> list[dict]:
    records = []
    with open(path) as f:
        for line in f:
            line = line.strip()
            if line:
                try:
                    records.append(json.loads(line))
                except json.JSONDecodeError:
                    pass
    return records

def to_parquet(records: list[dict], out: Path) -> int:
    if not records:
        print(f"  [skip] no records")
        return 0
    df = pd.DataFrame(records)
    # ensure consistent string types for mixed-type columns
    for col in df.columns:
        if df[col].dtype == object:
            df[col] = df[col].astype(str).replace("None", "")
    table = pa.Table.from_pandas(df, preserve_index=False)
    pq.write_table(table, out, compression="snappy")
    mb = out.stat().st_size / 1024 / 1024
    print(f"  → {out.name}  {len(records)} rows  {mb:.2f} MB  (snappy)")
    return len(records)

def push_file(api: HfApi, local: Path, remote: str, msg: str) -> None:
    print(f"  pushing {remote} …")
    op = CommitOperationAdd(path_in_repo=remote, path_or_fileobj=local.read_bytes())
    api.create_commit(
        repo_id=HF_REPO,
        repo_type="dataset",
        operations=[op],
        commit_message=msg,
        token=HF_TOKEN,
    )
    print(f"  ✓ {remote}")

def optimize_dogfeed(dry_run: bool, output_dir: Path | None = None) -> int:
    print("\n── dogfeed ──────────────────────────────────────")
    # in CI mode, dogfeed JSONL files are copied to the HF clone dir
    search_dirs = [DOGFEED_DIR]
    if output_dir:
        search_dirs.insert(0, output_dir)

    files: list[Path] = []
    for d in search_dirs:
        files = sorted(d.glob("dogfeed-*.jsonl"))
        if files:
            break
    if not files:
        print("  no dogfeed JSONL files found")
        return 0

    all_records: list[dict] = []
    for f in files:
        recs = read_jsonl(f)
        print(f"  {f.name}: {len(recs)} records")
        all_records.extend(recs)

    seen: set[str] = set()
    deduped = []
    for r in all_records:
        key = r.get("id") or hashlib.sha256(json.dumps(r, sort_keys=True).encode()).hexdigest()[:16]
        if key not in seen:
            seen.add(key)
            deduped.append(r)
    if len(deduped) < len(all_records):
        print(f"  deduped {len(all_records) - len(deduped)} duplicates")

    if output_dir:
        out = output_dir / "dogfeed.parquet"
        n = to_parquet(deduped, out)
        return n

    with tempfile.NamedTemporaryFile(suffix=".parquet", delete=False) as tmp:
        out = Path(tmp.name)
    n = to_parquet(deduped, out)
    if not dry_run and n > 0:
        api = HfApi()
        ts = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
        push_file(api, out, "dogfeed.parquet", f"optimize: dogfeed → parquet ({n} rows) [{ts}]")
    out.unlink(missing_ok=True)
    return n

def optimize_telemetry(dry_run: bool, output_dir: Path | None = None,
                       telemetry_path: Path | None = None) -> int:
    print("\n── telemetry ────────────────────────────────────")
    candidates = [
        telemetry_path,
        output_dir / "telemetry-consolidated.jsonl" if output_dir else None,
        DOGFEED_DIR.parent / "dogfood" / "telemetry-consolidated.jsonl",
    ]
    consolidated = next((p for p in candidates if p and p.exists()), None)
    if not consolidated:
        print("  telemetry-consolidated.jsonl not found — skipping")
        print("  (pass --telemetry-path or run CI consolidation first)")
        return 0

    records = read_jsonl(consolidated)
    print(f"  {consolidated.name}: {len(records)} events")

    if output_dir:
        out = output_dir / "telemetry.parquet"
        return to_parquet(records, out)

    with tempfile.NamedTemporaryFile(suffix=".parquet", delete=False) as tmp:
        out = Path(tmp.name)
    n = to_parquet(records, out)
    if not dry_run and n > 0:
        api = HfApi()
        ts = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
        push_file(api, out, "telemetry.parquet", f"optimize: telemetry → parquet ({n} events) [{ts}]")
    out.unlink(missing_ok=True)
    return n

def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument("--dry-run", action="store_true")
    p.add_argument("--stream", choices=["dogfeed", "telemetry"], default=None)
    p.add_argument("--output-dir", type=Path, default=None,
                   help="Write Parquet files here instead of pushing via API (CI mode)")
    p.add_argument("--telemetry-path", type=Path, default=None,
                   help="Path to telemetry-consolidated.jsonl (CI mode)")
    args = p.parse_args()

    if not args.dry_run and not args.output_dir and not HF_TOKEN:
        print("HF_TOKEN not set — use --dry-run, --output-dir, or export HF_TOKEN", file=sys.stderr)
        sys.exit(1)

    total = 0
    if args.stream in (None, "dogfeed"):
        total += optimize_dogfeed(args.dry_run, args.output_dir)
    if args.stream in (None, "telemetry"):
        total += optimize_telemetry(args.dry_run, args.output_dir, args.telemetry_path)

    mode = "CI write" if args.output_dir else ("dry run" if args.dry_run else "pushed to HF")
    print(f"\ndone: {total} total records [{mode}]")

if __name__ == "__main__":
    main()
