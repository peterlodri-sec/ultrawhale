# Multi-Platform ASM Layer — v66.0.0

## Architecture Detection

| Platform | SIMD | SHA | Detection |
|----------|------|-----|-----------|
| Apple Silicon (M1/M2/M3) | NEON | ARMv8 Crypto | Auto (arm64 build tag) |
| AMD EPYC / Intel Xeon | AVX2 | SHA-NI | `cpu.X86.Has*` at runtime |
| Generic amd64 | — | — | Pure Go fallback |
| Generic arm64 | NEON | — | Pure Go fallback |

## Dispatch Logic

```
HashDispatch(data)
  ├── Apple Silicon → BLAKE3 (NEON + ARMv8 Crypto)
  ├── Intel SHA-NI → BLAKE3 (AVX2 + SHA-NI)
  ├── AMD AVX2 → BLAKE3 (AVX2)
  └── Generic → SHA256 (pure Go)
```

## Files

| File | Platform | Purpose |
|------|----------|---------|
| `asm/multiarch.go` | All | ArchCapabilities struct + DetectArch() |
| `asm/multiarch_amd64.go` | amd64 | CPU feature detection via golang.org/x/sys/cpu |
| `asm/multiarch_arm64.go` | arm64 | Apple Silicon (NEON + ARMv8 always) |
| `hash_dispatch.go` | All | Route to fastest hash based on arch |
| `asm/text_amd64.s` | amd64 | SIMD substring + line count |
| `asm/text_arm64_go.go` | arm64 | Go fallback for text operations |

## Benchmarks (M1 Max)

| Operation | Throughput | Arch Features |
|-----------|-----------|---------------|
| BLAKE3 64KB | 2.3 GB/s | NEON + ARMv8 Crypto |
| Write 64KB | ~170µs | Journaled |
| Sed 1MB | 252 MB/s | SIMD bytes.Index |
D O C E O F  # intentional
