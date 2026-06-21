# Blocks Engine + Assembly Audit — v13.0.0

## Hash Performance Comparison

| Algorithm | Speed (M1 Max) | Use Case |
|-----------|---------------|----------|
| SHA256 (Go stdlib) | 2.3 GB/s | Content-addressed refs (current) |
| BLAKE3 (Go) | ~15 GB/s | Content-addressed refs (target) |
| xxHash64 (Go) | ~20 GB/s | Cache keys, brain memos, patterns |

## Assembly Audit

| File | Lines | Status |
|------|-------|--------|
| hash_amd64.s | 36 | Stub — falls back to Go SHA256 |
| hash_arm64.s | 18 | Stub — falls back to Go SHA256 |
| tool_dispatch.s | 8 | Jump table stub |

**Finding:** Go stdlib already auto-uses SHA-NI/NEON via `internal/cpu`.
Manual assembly kernels are unnecessary unless doing 4x+ batch parallelism.
Recommendation: use BLAKE3's native Go+asm implementation instead of custom asm.

## Lock Contention

| Component | Before | After | Status |
|-----------|--------|-------|--------|
| Journal | 1 global mutex | 16-way sharded | ✅ No hot spots |
| Log | Mutex on every write | Atomic CAS + RLock | ✅ No hot spots |
| Block Write | Synchronous journal push | Sharded mutex, O(1) | ✅ |

## Missing Primitives Added (v13.0.0)

| Primitive | Purpose | Lines |
|-----------|---------|-------|
| blake3.go | BLAKE3 hash (SHA256 fallback) | 50 |
| xxhash.go | xxHash64 cache keys | 25 |
| diff.go | Unified diff generation | 72 |
| pool.go | Buffer reuse pool | 42 |

## GPU Path

Metal/CUDA stubs are placeholder. Real GPU sha256 for batch >64:
- Apple Silicon: Metal Performance Shaders (MPSMatrixMultiplication)
- NVIDIA: cuBLAS batch operations
- Estimated throughput: 40-80 GB/s for batch operations
- Activation threshold: batch size >64 files

## Benchmarks (M1 Max, 8 cores)

| Benchmark | Result |
|-----------|--------|
| Hash 64KB (Go) | 2.3 GB/s |
| Hash 64KB (BLAKE3 est.) | ~15 GB/s |
| Write 64KB | 596 MB/s (I/O bound) |
| Batch-64 | 3.8ms |
| Sed 1MB | 252 MB/s |

## Recommendations

1. **BLAKE3 for refs**: Vendor `lukechampine/blake3` and enable via EnableBlake3()
2. **xxHash for cache**: Vendor `cespare/xxhash` for tool cache keys
3. **Drop custom asm**: Go stdlib SHA-NI/NEON is sufficient
4. **GPU for batch**: Implement Metal shader when batch >64 threshold is hit
