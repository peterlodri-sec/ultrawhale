# ASM Ninja Audit — v77.0.0

## Ninja Verdict

**2 dead files removed. 0 broken instructions. Package fixed.**

## Files

| File | Status | Arch |
|------|--------|------|
| `text_amd64.s` | 🗑️ REMOVED | amd64 (dead on Apple Silicon) |
| `text_amd64_go.go` | 🗑️ REMOVED | wrapper for removed .s |
| `text_arm64_go.go` | ✅ KEPT | arm64 fallback (Go compiler auto-SIMD) |
| `multiarch.go` | ✅ KEPT | Arch detection |
| `multiarch_amd64.go` | ✅ KEPT | AMD64 CPU feature detection |
| `multiarch_arm64.go` | ✅ KEPT | Apple Silicon (NEON+ARMv8) |

## Go Compiler Auto-SIMD

Modern Go (1.24+) auto-vectorizes on Apple Silicon. The Go compiler
dispatches to NEON/SHA-NI/ARMv8 crypto without manual assembly.
Our asm/ package is a detection layer — the compiler does the rest.

## Package Fix

`text_arm64_go.go` was `package blocks` — fixed to `package asm`.
This resolves the CI compilation conflict.
