#!/bin/bash
# ultrawhale benchmark runner — stores results in .dev/benches/
set -e
TS=$(date +%Y%m%d-%H%M%S)
DIR=".dev/benches/$TS"
mkdir -p "$DIR"

echo "=== ultrawhale benchmark — $TS ==="

# Blocks benchmarks
go test -bench=. -benchmem -benchtime=500ms ./internal/blocks/ 2>&1 | tee "$DIR/blocks.txt"

# TUI benchmark
go build -o bin/ultrawhale-bench-tui ./cmd/bench-tui/ 2>&1
bin/ultrawhale-bench-tui 2>&1 | tee "$DIR/tui.txt"

# Report
echo ""
echo "Results: $DIR"
echo "  blocks.txt — blocks engine benchmarks"
echo "  tui.txt    — TUI load test"
