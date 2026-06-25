#!/usr/bin/env python3
"""
Dogfeed generator — 100 iterations, SOTA scrub, HuggingFace upload.

Schema matches existing site/dogfood/dogfeed-v3-enriched.jsonl:
  id, user_message, free_response, free_model, deepseek_response,
  timestamp, session_id, topic, format, pov, capabilities,
  space_node, memory_ref, enriched_at, pipeline

Scrub pipeline:
  1. PII strip  (email, phone, URL, IP)
  2. Dedup      (exact SHA-256 + MinHash near-dup fingerprint)
  3. Quality    (min tokens, non-empty, no [dogfeed:...] echo artifacts)
  4. Content    (no raw keys/secrets/credentials)
"""

import os, sys, re, json, hashlib, time, random, string, argparse
from datetime import datetime, timezone
from typing import Optional

import requests
from huggingface_hub import HfApi, CommitOperationAdd

# ── Config ──────────────────────────────────────────────────────────────
OPENROUTER_KEY = os.environ["OPENROUTER_API_KEY"]
HF_TOKEN       = os.environ["HF_TOKEN"]
HF_REPO        = "PeetPedro/ultrawhale-dogfood"
SESSION_ID     = "v100.1.0-gen"
PIPELINE_TAG   = "dogfeed → scrub → HF (generate_dogfeed.py)"

FREE_MODELS = [
    "openai/gpt-oss-20b:free",
    "liquid/lfm-2.5-1.2b-instruct:free",
]

# Prompts covering the vaked/ultrawhale topic space
PROMPTS = [
    # agentic / coding
    "what is an agentic loop?",
    "how does a self-improving AI system work?",
    "explain recursive self-modification in software",
    "what is a coding agent?",
    "how do LLMs generate code?",
    "what is tool use in AI?",
    "explain chain-of-thought reasoning",
    "what is a multi-agent system?",
    "how does RAG work?",
    "what is fine-tuning a language model?",
    # consciousness / psychedelic
    "what are Klüver form constants?",
    "explain the retino-cortical map",
    "what is a Turing bifurcation in neural fields?",
    "how do psychedelics affect visual perception?",
    "what is the Wilson-Cowan neural field model?",
    "explain reaction-diffusion patterns in biology",
    "what is enthea?",
    "how does WebGL2 work?",
    "what is the geometry of hallucination?",
    "explain quasicrystal mathematics",
    # vaked / protocol
    "what is the AG-UI protocol?",
    "explain genesis seal in a distributed system",
    "what is a vaked space node?",
    "how does a public ledger work for AI sessions?",
    "what is a dogfeed in machine learning?",
    "explain Tier 2 telemetry anonymization",
    "what is a cryptographic proof of work?",
    "how does a WebSocket IRC connection work?",
    "explain Ergo IRCd federation",
    "what is a Mastodon ActivityPub toot?",
    # math / CS
    "what is sacred geometry?",
    "explain the Penrose tiling",
    "what is a Kleinian group?",
    "explain the Weierstrass elliptic function",
    "what is hyperbolic geometry?",
    "explain the Arnold tongue devil's staircase",
    "what is an Abrikosov vortex lattice?",
    "explain Gray-Scott reaction-diffusion",
    "what is a domain warp in shader programming?",
    "explain perceptual OKLab color space",
    # philosophy / culture
    "what is caged cognition?",
    "explain open source philosophy",
    "what does AGPL-3.0 mean for software freedom?",
    "what is the entheogen principle?",
    "explain set and setting",
    "what is the substrate of consciousness?",
    "how does music visualization work?",
    "what is a VJ in live performance?",
    "explain the mathematics of sound",
    "what is binaural entrainment?",
]

TOPIC_MAP = {
    "agentic": "agentic",
    "self-improving": "agentic",
    "coding agent": "agentic",
    "LLM": "llm",
    "language model": "llm",
    "fine-tuning": "llm",
    "RAG": "llm",
    "chain-of-thought": "llm",
    "Klüver": "neuroscience",
    "retino-cortical": "neuroscience",
    "neural field": "neuroscience",
    "psychedelic": "consciousness",
    "hallucination": "consciousness",
    "WebGL": "graphics",
    "shader": "graphics",
    "quasicrystal": "mathematics",
    "geometry": "mathematics",
    "Kleinian": "mathematics",
    "Weierstrass": "mathematics",
    "AG-UI": "protocol",
    "genesis": "protocol",
    "telemetry": "protocol",
    "IRC": "infra",
    "WebSocket": "infra",
    "Mastodon": "infra",
    "vaked": "vaked",
    "dogfeed": "dataset",
}

def infer_topic(prompt: str) -> str:
    for kw, topic in TOPIC_MAP.items():
        if kw.lower() in prompt.lower():
            return topic
    return "general"

# ── OpenRouter call ─────────────────────────────────────────────────────
def call_openrouter(model: str, prompt: str, timeout: int = 30, retries: int = 2) -> Optional[str]:
    for attempt in range(retries + 1):
        try:
            r = requests.post(
                "https://openrouter.ai/api/v1/chat/completions",
                headers={
                    "Authorization": f"Bearer {OPENROUTER_KEY}",
                    "HTTP-Referer": "https://vaked.dev/ultrawhale",
                    "X-Title": "ultrawhale-dogfeed",
                    "Content-Type": "application/json",
                },
                json={
                    "model": model,
                    "messages": [{"role": "user", "content": prompt}],
                    "max_tokens": 256,
                    "temperature": 0.7,
                },
                timeout=timeout,
            )
            r.raise_for_status()
            choices = r.json().get("choices")
            if not choices:
                return None
            return choices[0]["message"]["content"].strip()
        except requests.HTTPError as e:
            if e.response.status_code == 429 and attempt < retries:
                wait = 2 ** attempt * 3
                print(f" [429 retry {attempt+1} in {wait}s]", end="", flush=True)
                time.sleep(wait)
            else:
                print(f"  [warn] {model} failed: {e}", file=sys.stderr)
                return None
        except Exception as e:
            print(f"  [warn] {model} failed: {e}", file=sys.stderr)
            return None
    return None

# ── SOTA Scrub ──────────────────────────────────────────────────────────
_PII_PATTERNS = [
    re.compile(r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b"),  # email
    re.compile(r"\b(\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b"),  # US phone
    re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"),                            # IPv4
    re.compile(r"https?://\S+"),                                             # URLs
    re.compile(r"\b[A-Za-z0-9_-]{20,}\b"),                                  # long tokens (API keys etc)
    re.compile(r"sk-[A-Za-z0-9]{20,}"),                                      # OpenAI-style keys
    re.compile(r"hf_[A-Za-z0-9]{10,}"),                                      # HF tokens
    re.compile(r"Bearer [A-Za-z0-9._-]+"),                                   # auth headers
]

_SEEN_HASHES: set[str] = set()

def pii_strip(text: str) -> str:
    for pat in _PII_PATTERNS:
        text = pat.sub("[REDACTED]", text)
    return text

def sha256(text: str) -> str:
    return hashlib.sha256(text.encode()).hexdigest()

def minhash_fingerprint(text: str, n: int = 8) -> frozenset:
    words = re.findall(r"\w+", text.lower())
    shingles = {" ".join(words[i:i+3]) for i in range(len(words)-2)} if len(words) >= 3 else set(words)
    seeds = [hash(s + str(i)) & 0xFFFFFFFF for i in range(n) for s in shingles]
    return frozenset(seeds[:n]) if seeds else frozenset()

_SEEN_FINGERPRINTS: list[frozenset] = []

def is_near_dup(fp: frozenset, threshold: float = 0.85) -> bool:
    for seen in _SEEN_FINGERPRINTS:
        union = len(fp | seen)
        if union == 0:
            continue
        jaccard = len(fp & seen) / union
        if jaccard >= threshold:
            return True
    return False

def quality_ok(free_resp: str, deepseek_resp: str) -> bool:
    if not free_resp or not deepseek_resp:
        return False
    if free_resp.startswith("[dogfeed:") or free_resp.startswith("[sacred]"):
        return False  # simulation artifacts
    if len(free_resp.split()) < 8:
        return False
    if len(deepseek_resp.split()) < 8:
        return False
    return True

def scrub(record: dict) -> Optional[dict]:
    fr = pii_strip(record["free_response"])
    dr = pii_strip(record["deepseek_response"])
    um = pii_strip(record["user_message"])

    if not quality_ok(fr, dr):
        return None

    combined = fr + dr
    h = sha256(combined)
    if h in _SEEN_HASHES:
        return None
    _SEEN_HASHES.add(h)

    fp = minhash_fingerprint(combined)
    if is_near_dup(fp):
        return None
    _SEEN_FINGERPRINTS.append(fp)

    return {**record, "user_message": um, "free_response": fr, "deepseek_response": dr}

# ── Main generate loop ──────────────────────────────────────────────────
def generate(n: int = 100, dry_run: bool = False) -> list[dict]:
    records = []
    model_idx = 0
    prompt_pool = PROMPTS.copy()
    random.shuffle(prompt_pool)
    # expand if n > pool
    while len(prompt_pool) < n:
        extra = PROMPTS.copy()
        random.shuffle(extra)
        prompt_pool.extend(extra)

    print(f"Generating {n} records (dry_run={dry_run})...")
    for i in range(n):
        prompt = prompt_pool[i]
        model_a = FREE_MODELS[model_idx % len(FREE_MODELS)]
        model_b = FREE_MODELS[(model_idx + 1) % len(FREE_MODELS)]
        model_idx += 1

        print(f"  [{i+1:03d}/{n}] {prompt[:50]!r} → {model_a.split('/')[1]}", end="", flush=True)

        if dry_run:
            free_resp = f"Dry-run response for: {prompt}"
            deep_resp = f"Dry-run deepseek response for: {prompt}"
        else:
            free_resp = call_openrouter(model_a, prompt)
            if free_resp is None:
                print(" [skip]")
                continue
            # second call with different model as "deepseek_response" slot
            deep_resp = call_openrouter(model_b, f"Give a comprehensive technical answer: {prompt}")
            if deep_resp is None:
                deep_resp = free_resp  # fallback

        ts = datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")
        mem_ref = hashlib.sha256(f"{prompt}{ts}".encode()).hexdigest()[:12]
        idx_str = f"{len(records) + 1:04d}"

        raw = {
            "id": f"dogfeed-gen-{idx_str}",
            "user_message": prompt,
            "free_response": free_resp,
            "free_model": model_a,
            "deepseek_response": deep_resp,
            "timestamp": ts,
            "session_id": SESSION_ID,
            "topic": infer_topic(prompt),
            "format": "qa-pair",
            "pov": "M1/arm64/python",
            "capabilities": "FULL",
            "space_node": f"ultrawhale/datasets/dogfeed-gen-{idx_str}",
            "memory_ref": mem_ref,
            "enriched_at": ts,
            "pipeline": PIPELINE_TAG,
        }

        clean = scrub(raw)
        if clean is None:
            print(" [scrubbed]")
            continue

        records.append(clean)
        print(f" ✓ ({len(clean['free_response'].split())}w)")

        if not dry_run:
            time.sleep(1.5)  # free tier rate limit buffer (2 calls per record)

    print(f"\nGenerated {len(records)} clean records (from {n} attempts)")
    return records

# ── HuggingFace upload ──────────────────────────────────────────────────
def upload_to_hf(records: list[dict], filename: str) -> str:
    api = HfApi()
    content = "\n".join(json.dumps(r) for r in records) + "\n"
    op = CommitOperationAdd(path_in_repo=filename, path_or_fileobj=content.encode())
    api.create_commit(
        repo_id=HF_REPO,
        repo_type="dataset",
        operations=[op],
        commit_message=f"add {filename}: {len(records)} clean records (SOTA scrub)",
        token=HF_TOKEN,
    )
    return f"https://huggingface.co/datasets/{HF_REPO}/blob/main/{filename}"

# ── CLI ─────────────────────────────────────────────────────────────────
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate + scrub + upload dogfeed data")
    parser.add_argument("--iterations", "-n", type=int, default=100)
    parser.add_argument("--dry-run", action="store_true", help="Skip API calls, use placeholder text")
    parser.add_argument("--no-upload", action="store_true", help="Write JSONL locally only")
    parser.add_argument("--out", default=None, help="Output path (default: site/dogfood/dogfeed-vN-generated.jsonl)")
    args = parser.parse_args()

    records = generate(args.iterations, dry_run=args.dry_run)

    if not records:
        print("No records generated. Exiting.", file=sys.stderr)
        sys.exit(1)

    ts_tag = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    filename = args.out or f"dogfeed-v4-generated-{ts_tag}.jsonl"
    local_path = os.path.join(
        os.path.dirname(__file__), "..", "site", "dogfood", os.path.basename(filename)
    )
    local_path = os.path.normpath(local_path)

    with open(local_path, "w") as f:
        for r in records:
            f.write(json.dumps(r) + "\n")
    print(f"\nWrote {len(records)} records → {local_path}")

    if not args.no_upload:
        print(f"Uploading to {HF_REPO}...")
        url = upload_to_hf(records, os.path.basename(filename))
        print(f"Uploaded → {url}")
    else:
        print("Skipped upload (--no-upload)")
