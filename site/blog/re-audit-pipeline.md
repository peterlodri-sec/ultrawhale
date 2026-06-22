# RE + Audit Pipeline

June 22, 2026 · Peter + CoCreator

---

**We built a complete end-to-end binary reverse engineering and audit pipeline.**
Every release is a verifiable, auditable artifact.

## The Pipeline

```
Release tag
  ↓
crabcc index ──→ 12,530 symbols, 70,015 edges
CTF Arena ─────→ "Locate ASCIIBox in the binary"
sandboxd ──────→ Namespace isolation (seccomp + cgroups)
SPACE+TIME ────→ Cryptographic proof of code state
Event Horizon ──→ Public, immutable, append-only
```

## Why This Matters

Every commit is signed. Every release is audited. Every binary is
reverse-engineerable. **Trust through transparency.**

The CTF challenge proves you CAN understand the code. The audit trail
proves you CAN verify the release. sandboxd proves you CAN contain
untrusted analysis.

## How It Works

1. **crabcc** indexes every symbol in the codebase — 877 files, 12,530
   symbols, 70,015 edges in the call graph.
2. **CTF Arena** challenges players to find specific functions in the
   binary. Solve the challenge = prove understanding.
3. **sandboxd** contains the RE execution in Linux namespaces + cgroups
   v2 + seccomp. Safe analysis of untrusted binaries.
4. **SPACE+TIME PROOF** cryptographically signs every release —
   machine, timestamp, content hash, watermark.
5. **Event Horizon** publishes everything. Public. Immutable. Honest.

## Try It

```bash
crabcc lookup sym ASCIIBox    # 12 references
crabcc lookup refs ASCIIBox   # verify the function signature
```

The flag is everywhere. The proof is in the code.

---

*"The builder becomes the built. The graph becomes the system."*
