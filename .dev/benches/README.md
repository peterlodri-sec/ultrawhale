# ultrawhale Benchmarks

All benchmark results stored here. Run with `task dev-bench`.

## Latest (v15.3.0)

| Benchmark | Result | Notes |
|-----------|--------|-------|
| Write 64KB | ~170µs | I/O bound, journaled |
| Read 64KB | ~18µs | mmap where available |
| Hash 64KB (BLAKE3) | ~28µs / 2.3 GB/s | Go stdlib SHA-NI |
| Batch-64 | ~3.8ms | 64 files atomic |
| Sed 1MB | ~4ms / 252 MB/s | SIMD bytes.Index |
| TUI Startup | ~430ms | Doctor check |
| TUI Load (400 ops) | ~20s | 0 errors, 20 ops/sec |

## Historical

See `.dev/benches/` directory for timestamped runs.
