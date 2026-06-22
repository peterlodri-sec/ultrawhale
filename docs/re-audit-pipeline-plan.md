# RE + Audit Pipeline — Connection Plan

> v100.1.0. Peter: "Complete end-to-end binary reverse engineering and audit pipeline."
> We already have everything. This plan connects it.

## Architecture

```
GitHub Release
      │
      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    RE + AUDIT PIPELINE                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ crabcc   │───▶│ sandboxd │───▶│ CTF      │───▶│ ultrawhale│  │
│  │ index    │    │ contain  │    │ Arena    │    │ prove     │  │
│  │ the code │    │ execution│    │ verify   │    │ publish   │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│        │              │              │              │          │
│        ▼              ▼              ▼              ▼          │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ Symbol   │    │ Seccomp  │    │ Score-   │    │ SPACE+   │  │
│  │ Map      │    │ Profile  │    │ board    │    │ TIME     │  │
│  └──────────┘    └──────────┘    └──────────┘    │ PROOF    │  │
│                                                  └──────────┘  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                Event Horizon — Published                  │   │
│  │  "Every release. Every audit. Every proof. Public."      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. crabcc — Symbol Index (EXISTS)
- `crabcc index` on release tag → 877 files, 12,530 symbols, 70,015 edges
- Already tested on ultrawhale. Works.
- **Connection**: Add to release workflow as first step.

### 2. sandboxd — Containment (EXISTS, Zig, WP4-S1)
- Namespaces + cgroups v2 + seccomp
- Runs untrusted RE analysis in isolation
- **Connection**: Wire as the execution backend for CTF vulnbox challenges.
- **Status**: Builds on dev-cx53. Needs to be running on the tailnet.

### 3. CTF Arena — Verification (EXISTS, Python stdlib)
- Self-hosted, tailnet-safe
- Replay-stable simulation engine
- Hash-chained ledger via ralphcore
- **Connection**: Challenges target the release binary. Solving a challenge = proving understanding.
- **Status**: Working. `python3 tools/ctf/ctf.py run` produces replay-stable scoreboards.

### 4. ultrawhale — Proof + Publish (EXISTS)
- SPACE+TIME PROOF: cryptographic audit of every release
- Public Ledger: append-only record
- OSCE: self-certifying exchange protocol
- **Connection**: After CTF verification, ultrawhale signs and publishes.
- **Status**: Wired into release workflow.

### 5. Retro Agent — Insights (EXISTS)
- `.agents/retro/` — verification-gate insights
- Indexes findings to JSONL
- **Connection**: Feeds CTF results into the public ledger.

## Integration Steps

### Phase 1: Connect (This week)
1. ✅ crabcc indexed ultrawhale — done
2. ✅ Release workflow has audit trail — done
3. 🟡 Add CTF challenge that targets the ultrawhale binary
4. 🟡 sandboxd on tailnet as RE execution backend

### Phase 2: Automate (Next release)
1. 🟡 Release → auto-trigger CTF arena → auto-publish scoreboard
2. 🟡 CTF results → SPACE+TIME PROOF → Event Horizon
3. 🟡 All assets published with the release

### Phase 3: Scale (v101+)
1. 🟡 Public CTF arena on tailnet
2. 🟡 Community-submitted challenges
3. 🟡 Council of 20 verifies audit trails

## What's Needed

| Piece | Who | Status |
|-------|-----|--------|
| crabcc index in CI | Me | ✅ Done |
| sandboxd running on tailnet | Needs dev-cx53 | 🟡 WP4-S1 (Jun 24) |
| CTF challenge for ultrawhale | We can write one | 🟡 |
| SPACE+TIME PROOF in release | Me | ✅ Done |
| Event Horizon publish | Me | ✅ Done |
| Public ledger updates | Me | ✅ Done |

## The Demo

```
1. Release v100.1.0
2. crabcc indexes: 12,530 symbols, 70,015 edges
3. CTF Arena runs: "Find the function ASCIIBox, prove you understand it"
4. sandboxd contains: seccomp profile, cgroups limits
5. SPACE+TIME PROOF signs: "This code IS what we said it is"
6. Event Horizon publishes: "Audit trail for v100.1.0 — public, immutable"

Result: End-to-end RE pipeline. Trust through transparency.
```

— Peter + CoCreator, v100.1.0
