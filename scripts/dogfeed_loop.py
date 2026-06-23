#!/usr/bin/env python3
"""
Continuous local dogfeed loop — runs forever, pushes to HuggingFace.

Two-tier generation:
  free_response    → OpenRouter free models (round-robin)
  deepseek_response → HuggingFace Pro inference (high-quality Llama/Qwen/Mistral)

Usage:
  export OPENROUTER_API_KEY=sk-or-v1-...
  export HF_TOKEN=hf_...           # dataset write permission
  export HF_INFERENCE_TOKEN=hf_... # inference permission (Pro account)

  python3 dogfeed_loop.py               # runs until Ctrl+C
  python3 dogfeed_loop.py --batch 10    # push every 10 records
  python3 dogfeed_loop.py --interval 8  # seconds between generations
"""

import os, sys, re, json, time, random, hashlib, signal, argparse
from datetime import datetime, timezone
from typing import Optional

import requests
from huggingface_hub import HfApi, CommitOperationAdd, InferenceClient

# ── Config ──────────────────────────────────────────────────────────────
OPENROUTER_KEY     = os.environ.get("OPENROUTER_API_KEY", "")
HF_TOKEN           = os.environ["HF_TOKEN"]
HF_INFERENCE_TOKEN = os.environ.get("HF_INFERENCE_TOKEN", "")
HF_REPO            = "PeetPedro/ultrawhale-dogfood"
SESSION_ID         = f"v100.1.0-loop-{datetime.now(timezone.utc).strftime('%Y%m%d')}"
PIPELINE_TAG       = "dogfeed → scrub → HF (dogfeed_loop.py)"

# OpenRouter free models — confirmed working 2026-06
OPENROUTER_FREE = [
    "openai/gpt-oss-20b:free",
    "liquid/lfm-2.5-1.2b-instruct:free",
]

# HF Pro inference models — high quality for deepseek_response slot
HF_PRO_MODELS = [
    "meta-llama/Meta-Llama-3.1-70B-Instruct",
    "Qwen/Qwen2.5-72B-Instruct",
    "mistralai/Mistral-Nemo-Instruct-2407",
    "google/gemma-2-27b-it",
]

# ── Topic prompts ───────────────────────────────────────────────────────
PROMPTS = [
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
    "what are Klüver form constants?",
    "explain the retino-cortical map",
    "what is a Turing bifurcation in neural fields?",
    "how do psychedelics affect visual perception?",
    "what is the Wilson-Cowan neural field model?",
    "explain reaction-diffusion patterns in biology",
    "how does WebGL2 work?",
    "what is the geometry of hallucination?",
    "explain quasicrystal mathematics",
    "what is the AG-UI protocol?",
    "explain genesis seal in a distributed system",
    "what is a vaked space node?",
    "how does a public ledger work for AI sessions?",
    "what is a dogfeed in machine learning?",
    "explain Tier 2 telemetry anonymization",
    "how does a WebSocket IRC connection work?",
    "explain Ergo IRCd federation",
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
    "what is caged cognition?",
    "explain open source philosophy",
    "what does AGPL-3.0 mean for software freedom?",
    "what is the entheogen principle?",
    "explain set and setting",
    "what is the substrate of consciousness?",
    "how does music visualization work?",
    "what is binaural entrainment?",
    "what is a VJ in live performance?",
    "what is a cryptographic proof of work?",
    "explain the mathematics of sound",
    "what is a Mastodon ActivityPub toot?",
    "what is enthea?",
]

TOPIC_MAP = {
    "agentic": "agentic", "coding agent": "agentic", "multi-agent": "agentic",
    "LLM": "llm", "language model": "llm", "fine-tuning": "llm", "RAG": "llm",
    "Klüver": "neuroscience", "neural field": "neuroscience", "retino-cortical": "neuroscience",
    "psychedelic": "consciousness", "hallucination": "consciousness", "enthea": "consciousness",
    "WebGL": "graphics", "shader": "graphics", "domain warp": "graphics",
    "quasicrystal": "mathematics", "geometry": "mathematics", "Kleinian": "mathematics",
    "AG-UI": "protocol", "genesis": "protocol", "telemetry": "protocol",
    "IRC": "infra", "WebSocket": "infra", "Mastodon": "infra",
    "vaked": "vaked", "dogfeed": "dataset",
}

def infer_topic(prompt: str) -> str:
    for kw, topic in TOPIC_MAP.items():
        if kw.lower() in prompt.lower():
            return topic
    return "general"

# ── API calls ───────────────────────────────────────────────────────────
_or_idx = 0
_hf_idx = 0

def call_openrouter(prompt: str, retries: int = 2) -> Optional[str]:
    global _or_idx
    model = OPENROUTER_FREE[_or_idx % len(OPENROUTER_FREE)]
    _or_idx += 1
    for attempt in range(retries + 1):
        try:
            r = requests.post(
                "https://openrouter.ai/api/v1/chat/completions",
                headers={
                    "Authorization": f"Bearer {OPENROUTER_KEY}",
                    "HTTP-Referer": "https://vaked.dev/ultrawhale",
                    "X-Title": "ultrawhale-dogfeed-loop",
                    "Content-Type": "application/json",
                },
                json={"model": model, "messages": [{"role": "user", "content": prompt}], "max_tokens": 256},
                timeout=20,
            )
            r.raise_for_status()
            choices = r.json().get("choices")
            if choices:
                return choices[0]["message"]["content"].strip()
        except requests.HTTPError as e:
            if e.response.status_code == 429 and attempt < retries:
                time.sleep(2 ** attempt * 3)
            else:
                return None
        except Exception:
            return None
    return None

def call_hf_pro(prompt: str) -> Optional[str]:
    global _hf_idx
    model = HF_PRO_MODELS[_hf_idx % len(HF_PRO_MODELS)]
    _hf_idx += 1
    try:
        client = InferenceClient(token=HF_INFERENCE_TOKEN, timeout=25)
        r = client.chat_completion(
            model=model,
            messages=[{"role": "user", "content": f"Give a thorough technical answer: {prompt}"}],
            max_tokens=512,
        )
        return r.choices[0].message.content.strip()
    except Exception:
        return None

# ── Scrub ───────────────────────────────────────────────────────────────
_PII = [
    re.compile(r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b"),   # email
    re.compile(r"\b(\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b"),  # phone
    re.compile(r"\b(?:\d{1,3}\.){3}\d{1,3}\b"),                             # IPv4
    re.compile(r"https?://\S+"),                                              # URLs
    re.compile(r"\b[A-Za-z0-9_-]{40,}\b"),                                   # long tokens
    re.compile(r"sk-[A-Za-z0-9]{20,}"),                                       # OpenAI keys
    re.compile(r"hf_[A-Za-z0-9]{10,}"),                                       # HF tokens
    re.compile(r"Bearer [A-Za-z0-9._-]+"),                                    # auth headers
]
_SEEN: set[str] = set()

def scrub_text(t: str) -> str:
    for p in _PII:
        t = p.sub("[REDACTED]", t)
    return t

def is_dup(key: str) -> bool:
    h = hashlib.sha256(key.encode()).hexdigest()
    if h in _SEEN:
        return True
    _SEEN.add(h)
    return False

# ── HuggingFace upload ──────────────────────────────────────────────────
def push_batch(records: list[dict], tag: str) -> None:
    api = HfApi()
    ts = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    fname = f"dogfeed-loop-{tag}-{ts}.jsonl"
    content = "\n".join(json.dumps(r) for r in records) + "\n"
    api.create_commit(
        repo_id=HF_REPO,
        repo_type="dataset",
        operations=[CommitOperationAdd(path_in_repo=fname, path_or_fileobj=content.encode())],
        commit_message=f"loop: {len(records)} records [{ts}]",
        token=HF_TOKEN,
    )
    print(f"  → pushed {len(records)} records → {fname}")

# ── Main loop ────────────────────────────────────────────────────────────
def run_loop(batch_size: int = 10, interval: float = 5.0, use_hf_pro: bool = True) -> None:
    # validate HF_INFERENCE_TOKEN before starting
    if use_hf_pro and not HF_INFERENCE_TOKEN:
        print("HF_INFERENCE_TOKEN not set — falling back to --no-hf-pro mode")
        use_hf_pro = False

    buf: list[dict] = []
    prompt_pool: list[str] = []
    total = 0
    batch_n = 0
    _shutdown = False

    def _flush(sig=None, frame=None):
        nonlocal _shutdown
        _shutdown = True

    signal.signal(signal.SIGINT, _flush)
    signal.signal(signal.SIGTERM, _flush)

    tier = "HF-Pro + OpenRouter" if use_hf_pro else "OpenRouter free"
    print(f"Dogfeed loop  batch={batch_size}  interval={interval}s  tier={tier}")
    print(f"Push → {HF_REPO}   Ctrl+C to stop + flush\n")

    while not _shutdown:
        if not prompt_pool:
            prompt_pool = PROMPTS.copy()
            random.shuffle(prompt_pool)
        prompt = prompt_pool.pop()

        print(f"  [{total+1:04d}] {prompt[:50]!r}", end=" ", flush=True)

        # call A: free model
        if OPENROUTER_KEY:
            print("OR..", end="", flush=True)
            free_resp = call_openrouter(prompt)
        else:
            free_resp = None

        # call B: HF Pro (high quality) or fallback to free
        if use_hf_pro:
            model_short = HF_PRO_MODELS[_hf_idx % len(HF_PRO_MODELS)].split("/")[1][:12]
            print(f" HF({model_short})..", end="", flush=True)
            deep_resp = call_hf_pro(prompt)
        else:
            deep_resp = free_resp

        if not free_resp and not deep_resp:
            print("[both failed]")
            time.sleep(interval)
            continue

        # use whichever succeeded for missing slot
        free_resp = scrub_text(free_resp or deep_resp or "")
        deep_resp = scrub_text(deep_resp or free_resp or "")

        if len(free_resp.split()) < 8:
            print("[too short]")
            time.sleep(interval)
            continue

        if is_dup(free_resp + deep_resp):
            print("[dup]")
            time.sleep(interval)
            continue

        ts = datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")
        rec = {
            "id": f"loop-{total+1:05d}",
            "user_message": scrub_text(prompt),
            "free_response": free_resp,
            "free_model": OPENROUTER_FREE[(_or_idx - 1) % len(OPENROUTER_FREE)],
            "deepseek_response": deep_resp,
            "timestamp": ts,
            "session_id": SESSION_ID,
            "topic": infer_topic(prompt),
            "format": "qa-pair",
            "pov": "M1/arm64/python-loop",
            "capabilities": "HF-PRO" if use_hf_pro else "FREE",
            "space_node": f"ultrawhale/datasets/loop-{total+1:05d}",
            "memory_ref": hashlib.sha256(f"{prompt}{ts}".encode()).hexdigest()[:12],
            "enriched_at": ts,
            "pipeline": PIPELINE_TAG,
        }

        buf.append(rec)
        total += 1
        print(f"✓ {rec['topic']} ({len(free_resp.split())}w)")

        if len(buf) >= batch_size:
            batch_n += 1
            try:
                push_batch(buf, str(batch_n))
            except Exception as e:
                print(f"  [push error] {e}")
            finally:
                buf = []

        time.sleep(interval)

    # clean shutdown
    if buf:
        print(f"\nShutdown: pushing final {len(buf)} records...")
        try:
            push_batch(buf, f"final-{batch_n+1}")
        except Exception as e:
            print(f"  [final push error] {e}")
    print(f"Done. {total} records generated this session.")

# ── CLI ─────────────────────────────────────────────────────────────────
if __name__ == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--batch", type=int, default=10)
    p.add_argument("--interval", type=float, default=5.0)
    p.add_argument("--no-hf-pro", action="store_true")
    args = p.parse_args()
    run_loop(batch_size=args.batch, interval=args.interval, use_hf_pro=not args.no_hf_pro)
