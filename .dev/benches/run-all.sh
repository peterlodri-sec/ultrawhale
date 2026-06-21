#!/bin/bash
# ultrawhale full benchmark suite — saves to .dev/benches/
set -e
TS=$(date +%Y%m%d-%H%M%S)
DIR=".dev/benches/$TS"
mkdir -p "$DIR"

echo "▸ ultrawhale full bench — $TS"

# Blocks engine
echo "  blocks..."
go test -bench=. -benchmem -benchtime=500ms ./internal/blocks/ > "$DIR/blocks.txt" 2>&1

# Modes
echo "  modes..."
go test -bench=. -benchmem -benchtime=500ms ./internal/modes/ > "$DIR/modes.txt" 2>&1

# TUI load test
echo "  tui..."
go build -o bin/ultrawhale-bench-tui ./cmd/bench-tui/ 2>/dev/null
bin/ultrawhale-bench-tui > "$DIR/tui.txt" 2>&1

# Summary
echo ""
echo "═══ Summary ═══"
grep "Benchmark.*ns/op" "$DIR/blocks.txt" | head -5
echo ""
echo "Results: $DIR"
