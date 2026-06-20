# ultrawhale dev artifacts

- `benches/` — benchmark results (gitignored)
- `logs/` — session logs
- `profiles/` — pprof CPU/memory profiles
- `tmp/` — temporary build artifacts

## Quick start

```sh
# Run benchmarks
.dev/bench.sh

# Run tests with race detection
go test -race -count=1 ./...

# Build
go build -o bin/ultrawhale ./cmd/whale

# Run
bin/ultrawhale --model deepseek-v4-flash -w
```
