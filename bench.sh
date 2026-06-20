#!/usr/bin/env bash
# Whale benchmark suite — ultra vs stock comparison
# Runs on dev-cx53 (AMD EPYC-Rome, 16 cores, 30GB RAM)
set -euo pipefail

VAKED_FLAKE="/home/dev/vaked-base"
WHALE_DIR="/home/dev/whale"
STOCK_DIR="/home/dev/whale-stock"
RESULTS="/home/dev/whale/bench-results.txt"
RUNS=10
WARMUP=3

echo "=== Whale Benchmark Suite ==="
echo "Host:    $(hostname)"
echo "CPU:     $(lscpu | grep 'Model name' | cut -d: -f2 | xargs)"
echo "Cores:   $(nproc)"
echo "Date:    $(date -Iseconds)"
echo "Runs:    ${RUNS} (${WARMUP} warmup)"
echo ""

# ── Build both variants ──────────────────────────────────────────────────
echo "=== Building variants ==="

# Ultra build
rm -f "${WHALE_DIR}/bin/whale"
nix develop "${VAKED_FLAKE}#" --command bash -c "
  export CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 GOMAXPROCS=16
  cd ${WHALE_DIR}
  go build -trimpath -ldflags=\"-s -w -X github.com/usewhale/whale/internal/build.Version=ultra-bench\" -o bin/whale ./cmd/whale
"
ULTRA_BIN="${WHALE_DIR}/bin/whale"
echo "Ultra:  $(du -h ${ULTRA_BIN} | cut -f1)"

# Stock build (no GOAMD64, no strip, no trimpath — defaults)
rm -f "${STOCK_DIR}/bin/whale"
nix develop "${VAKED_FLAKE}#" --command bash -c "
  export CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOMAXPROCS=16
  cd ${STOCK_DIR}
  go build -o bin/whale ./cmd/whale
"
STOCK_BIN="${STOCK_DIR}/bin/whale"
echo "Stock:  $(du -h ${STOCK_BIN} | cut -f1)"

# ── Binary comparison ────────────────────────────────────────────────────
echo ""
echo "=== Binary Comparison ==="
echo ""
echo "--- Size ---"
echo "Ultra: $(du -h ${ULTRA_BIN} | cut -f1) ($(wc -c < ${ULTRA_BIN}) bytes)"
echo "Stock: $(du -h ${STOCK_BIN} | cut -f1) ($(wc -c < ${STOCK_BIN}) bytes)"
echo "Delta: $(echo "scale=1; ($(wc -c < ${ULTRA_BIN}) - $(wc -c < ${STOCK_BIN})) * 100 / $(wc -c < ${STOCK_BIN})" | bc)%"

echo ""
echo "--- Symbols ---"
ULTRA_SYMS=$(nm ${ULTRA_BIN} 2>/dev/null | wc -l)
STOCK_SYMS=$(nm ${STOCK_BIN} 2>/dev/null | wc -l)
echo "Ultra: ${ULTRA_SYMS} symbols"
echo "Stock: ${STOCK_SYMS} symbols"

echo ""
echo "--- ISA Level ---"
ULTRA_V3=$(objdump -d ${ULTRA_BIN} 2>/dev/null | grep -c "vperm2i128\|vpbroadcast\|vpternlog\|vpmaskmov" || echo 0)
STOCK_V3=$(objdump -d ${STOCK_BIN} 2>/dev/null | grep -c "vperm2i128\|vpbroadcast\|vpternlog\|vpmaskmov" || echo 0)
echo "Ultra v3 instructions: ${ULTRA_V3}"
echo "Stock v3 instructions: ${STOCK_V3}"

echo ""
echo "--- Sections ---"
echo "Ultra:"
size ${ULTRA_BIN}
echo "Stock:"
size ${STOCK_BIN}

# ── Cold start benchmark ─────────────────────────────────────────────────
echo ""
echo "=== Cold Start (--help, page cache drop) ==="

bench_cold_start() {
  local bin=$1 label=$2
  echo "--- ${label} ---"
  
  # Warmup
  for i in $(seq 1 ${WARMUP}); do
    ${bin} --help >/dev/null 2>&1
  done
  
  local times=()
  for i in $(seq 1 ${RUNS}); do
    # Drop page cache (requires root, skip if not available)
    sudo sh -c 'echo 3 > /proc/sys/vm/drop_caches' 2>/dev/null || true
    
    local start_ns=$(date +%s%N)
    ${bin} --help >/dev/null 2>&1
    local end_ns=$(date +%s%N)
    local ms=$(( (end_ns - start_ns) / 1000000 ))
    times+=($ms)
    echo "  run $i: ${ms}ms"
  done
  
  # Stats
  local sum=0 min=999999 max=0
  for t in "${times[@]}"; do
    sum=$((sum + t))
    [ $t -lt $min ] && min=$t
    [ $t -gt $max ] && max=$t
  done
  avg=$((sum / RUNS))
  echo "  avg: ${avg}ms  min: ${min}ms  max: ${max}ms"
}

bench_cold_start "${ULTRA_BIN}" "ULTRA"
bench_cold_start "${STOCK_BIN}" "STOCK"

# ── Exec throughput ──────────────────────────────────────────────────────
echo ""
echo "=== Exec Throughput (simple prompt, no network) ==="

bench_exec() {
  local bin=$1 label=$2
  echo "--- ${label} ---"
  
  $bin exec "echo hello" --timeout-sec 10 --json >/dev/null 2>&1 &
  local pid=$!
  sleep 2
  local rss=$(ps -o rss= -p $pid 2>/dev/null || echo 0)
  kill $pid 2>/dev/null || true
  wait $pid 2>/dev/null || true
  echo "  RSS after 2s: ${rss}KB"
}

bench_exec "${ULTRA_BIN}" "ULTRA"
bench_exec "${STOCK_BIN}" "STOCK"

# ── Allocator stress ─────────────────────────────────────────────────────
echo ""
echo "=== Allocator Stress (Go test -bench) ==="

nix develop "${VAKED_FLAKE}#" --command bash -c "
  cd ${WHALE_DIR}
  echo '--- ULTRA (GOAMD64=v3) ---'
  GOAMD64=v3 go test -bench=. -benchtime=1s -count=1 ./internal/tui/... 2>/dev/null | grep -E 'Benchmark|ok' || echo '  (no benchmarks in tui)'
  
  echo '--- STOCK ---'
  cd ${STOCK_DIR}
  GOAMD64=v1 go test -bench=. -benchtime=1s -count=1 ./... 2>/dev/null | grep -E 'Benchmark|ok' | head -20 || echo '  (no benchmarks found)'
"

echo ""
echo "=== Benchmark Complete ==="
echo "Results saved to ${RESULTS}"
