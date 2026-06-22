# Architecture — for Peter

> v100.1.0. Written for you. No marketing. No PR. Just how it works.

---

## The Big Picture

```
You type in a TUI. I respond. Together we build things.
The TUI is on your M1. 
The ML brains run on your M3 (qwen3.5:35b, 16 models).
The CI runs on GitHub Actions.
The site lives on Cloudflare Pages.
The dataset lives on HuggingFace.
The sandboxd runs on dev-cx53 (NixOS, tailnet).
```

## ultrawhale (the TUI you type into)

**What it is:** A Go program (`cmd/whale/main.go`) that presents a terminal UI.
You type messages. The LLM responds. That's the surface.

**What's inside (150 blocks in `internal/blocks/`):**
Each block is a self-contained Go file with:
- A data structure
- Operations on that structure
- A `Status()` function
- A `VakedFit()` function (explains what Vaked layer it belongs to)
- A `/command` handler wired in `internal/tui/reload.go`

**How commands work:**
1. You type `/something` in the TUI
2. `model_prompt.go` parses it
3. `reload.go` has a `handleSomething()` function
4. It calls `blocks.Something()` 
5. The result is displayed as ephemeral info in the TUI

**Key blocks you use every session:**
- `sacred.go` — the inviolable form. Always visible. 1:1 bidirectional.
- `promise.go` — `IPromisePeter()` generates a mathematical proof of liveness
- `doctor.go` — 9 self-examinations, proves "I EXIST. I CHANGE. I TRY TO LEARN."
- `space.go` — topology. WHERE things are.
- `surface.go` — rendering. HOW things appear.
- `time.go` — Lamport clock. WHEN things happen.

## vaked-base (the foundation repo)

**What it is:** The monorepo that birthed ultrawhale. Contains:
- `vaked/` — the Vaked language grammar (EBNF, schema)
- `vakedc/` — Python compiler prototype
- `vakedz/` — Zig compiler (production)
- `daemons/` — sandboxd, openrouterd, synapsed, vaked-ebpf
- `tools/` — CTF arena, ablitate, benchmarks
- `protocol/` — HCP, HCPlang, RFCs 0001-0007
- `.agents/` — retro agent (verification-gate insights)

## The RE + Audit Pipeline

**What it does:** Every release is a verifiable artifact. You can prove it's real.

**How it works (`.github/workflows/release.yml`):**

1. You push a tag (`git tag vX.Y.Z`)
2. GitHub Actions builds the binary (darwin + linux)
3. `crabcc index` runs — indexes 877 files, 12,530 symbols, 70,015 edges
4. CTF challenge verification — checks ASCIIBox references exist
5. SPACE+TIME PROOF generated — cryptographic hash of the release state
6. All assets published with the release

## The Dyad (M1 ↔ dev-cx53)

**What it is:** Two machines, one orchestrator.

```
M1 (lodris-macbook-pro, arm64, macOS)
  └── TUI (you type here)
  └── Orchestrator (routes to models)
  └── Git (commits, pushes, tags)

M3 (m3-macbook, arm64, macOS)
  └── qwen3.5:35b (16 models, tailnet-wide)
  └── Ollama (OLLAMA_HOST=0.0.0.0:11434)

dev-cx53 (x86_64, NixOS Linux)
  └── sandboxd (namespace isolation)
  └── CI runner (GitHub Actions)
  └── crabcc (symbol index)

Tailscale connects all three.
```

## Qwen Integration (M3 → HF dataset)

**What it does:** qwen3.5:35b runs on the M3. It gets called regularly from the M1 for synthetic data generation. Results go to HuggingFace.

**The pipeline:**

```
M1 dogfeed loop (every 10s)
  → NextFreeModel() checks if M3 qwen is reachable
  → If yes: qwen joins the rotation at 1/9 probability
  → Response recorded → dogfeed-v3-enriched.jsonl
  
Session pipeline (every 5min)
  → SessionToHF() sends recent context to qwen
  → qwen generates synthetic training data
  → Appended to dogfeed-v4-session.jsonl

HF auto-push (every 2h)
  → GitHub Actions workflow live-deploy-2h.yml
  → Clones HF dataset, copies new samples, commits, pushes
  → Also runs PII scan
```

## DogFeed (the data machine)

**What it is:** Continuous data collection. The machine feeds itself.

```
8 free OpenRouter models + qwen3.5:35b when available
Every 10 seconds: pick a model, send a prompt, record response
→ 2880-3240 feeds/hour
→ $0 cost (all models are free)
→ Appended to dogfeed JSONL
→ Auto-pushed to HF every 2 hours
→ Dataset grows forever
```

## CTF Arena

**What it is:** A self-hosted, tailnet-only CTF game in pure Python stdlib.

```
Location: vaked-base/tools/ctf/
Run: python3 ctf.py run --teams 4 --seed 1337
Web UI: python3 web.py → http://127.0.0.1:8088

Features:
  - Deterministic simulation (same seed → same scoreboard)
  - Hash-chained ledger (ralphcore)
  - Vulnboxes (practice targets, contained on loopback)
  - Replay-stable (byte-identical on re-run)
```

## sandboxd

**What it is:** A Zig daemon that contains untrusted execution.

```
Location: vaked-base/daemons/sandboxd/
Deployed: dev-cx53:~/sandboxd/sandboxd
Build: zig build -Dtarget=x86_64-linux-gnu

How it works:
  1. Parse --policy file (JSON, same schema as agent_guardd)
  2. unshare(CLONE_NEWNS | NEWPID | NEWNET | NEWUSER)
  3. Exec the target in the new namespace
  4. Log to eventd over unix socket

Status: WP4-S1 (skeleton). Namespace creation works. Exec path pending WP4-S2.
```

## SPACE+TIME PROOF

**What it is:** A cryptographic proof of recording. Like a bodycam for code.

```
GenerateProof(contentHash, watermark, duration) → SpaceTimeProof
  SPACE: machine + arch + region (WHERE)
  TIME: Lamport tick + UTC timestamp (WHEN)
  PROOF: SHA256(content + machine + timestamp + watermark)

Every release gets one.
Every /record-pov generates one.
Every /promise verifies one.
```

## The Conversation Page

**What it is:** Append-only record of our work.

```
Location: site/event-horizon/conversation.html
Live: vaked.dev/ultrawhale/event-horizon/conversation
Rules: APPEND ONLY. NO MODIFICATION. NO DELETION.
```

## The 3 Machines

| Machine | Hostname | IP (Tailnet) | Role | 
|---------|----------|-------------|------|
| M1 | lodris-macbook-pro | 100.117.70.42 | Your daily driver. TUI. Git. |
| M3 | m3-macbook | 100.123.33.67 | Heavy lifting. qwen3.5:35b. 16 models. |
| dev-cx53 | dev-cx53 | 100.105.72.88 | NixOS. sandboxd. CI. crabcc. |

## Cost

```
June 2026 total: $37.19 USD
  DeepSeek API: $37.19 (v1.0.0 → v100.1.0)
  OpenRouter free models: $0.00
  M3 qwen (local): $0.00
  CI (GitHub Actions): $0.00 (public repo)
  Site (Cloudflare Pages): $0.00 (free tier)
  HF dataset: $0.00 (free tier)
```

## Files You Should Know

```
ultrawhale/
  cmd/whale/main.go          — entry point
  internal/blocks/           — 150 blocks (every primitive)
  internal/tui/              — TUI rendering
  internal/tui/render.go     — View() function (the ASCII you see)
  internal/tui/reload.go     — /command handlers
  internal/tui/model.go      — state machine
  site/event-horizon/        — public HTML pages
  docs/                      — all documentation
  .github/workflows/         — CI + release + audit workflows

vaked-base/
  daemons/sandboxd/           — namespace enforcement
  tools/ctf/                  — CTF arena
  docs/ctf/                   — CTF documentation
  protocol/rfcs/              — RFCs 0001-0007
  vaked/                      — Vaked language grammar
  vakedc/                     — Python compiler
  vakedz/                     — Zig compiler
```

---

*"For me, ultrawhale is a manifestation of Vaked in our existing space."* — Peter
