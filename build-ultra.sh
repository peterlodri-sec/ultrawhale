#!/usr/bin/env bash
# Ultra-performance Whale build for dev-cx53 (AMD EPYC-Rome, AVX2, BMI2, FMA)
# Uses vaked-base nix shell for Go 1.26.3
set -euo pipefail

WHALE_DIR="/home/dev/whale"
VAKED_FLAKE="/home/dev/vaked-base"
BIN="${WHALE_DIR}/bin/whale"
VERSION="${VERSION:-$(date +%Y%m%d-%H%M)-ultra}"

echo "=== Whale Ultra Build ==="
echo "Target:  $(hostname) — $(lscpu | grep 'Model name' | cut -d: -f2 | xargs)"
echo "Version: ${VERSION}"
echo ""

cd "${WHALE_DIR}"

# Clean previous build artifacts
rm -rf bin/ .gocache/

# Ultra build flags:
#   -trimpath       → reproducible, strips source paths
#   -ldflags="-s -w" → strip debug info + DWARF symbol table
#   GOAMD64=v3       → target AVX2, BMI2, FMA (EPYC-Rome sweet spot)
#   CGO_ENABLED=0    → static binary, no libc dependency
#   GOOS=linux       → explicit target
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOAMD64=v3
export GOMAXPROCS=16

LDFLAGS="-s -w -X github.com/usewhale/whale/internal/build.Version=${VERSION}"

echo "Building with:"
echo "  CGO_ENABLED=${CGO_ENABLED}"
echo "  GOAMD64=${GOAMD64}"
echo "  GOMAXPROCS=${GOMAXPROCS}"
echo "  LDFLAGS=${LDFLAGS}"
echo ""

START_NS=$(date +%s%N)

nix develop "${VAKED_FLAKE}#" --command bash -c "
  export CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 GOMAXPROCS=16
  cd ${WHALE_DIR}
  go build \
    -trimpath \
    -ldflags=\"${LDFLAGS}\" \
    -o ${BIN} \
    ./cmd/whale
"

END_NS=$(date +%s%N)
BUILD_TIME_MS=$(( (END_NS - START_NS) / 1000000 ))

echo ""
echo "=== Build Complete ==="
echo "Binary:  ${BIN}"
echo "Size:    $(du -h ${BIN} | cut -f1)"
echo "Time:    ${BUILD_TIME_MS}ms"
echo ""

# Verify
echo "=== Verification ==="
file "${BIN}"
echo ""
echo "Dynamic links:"
ldd "${BIN}" 2>&1 || echo "(static binary — no dynamic links)"
echo ""
echo "Symbol count:"
nm "${BIN}" 2>/dev/null | wc -l || echo "(fully stripped)"
echo ""
echo "Section sizes:"
size "${BIN}"
echo ""
echo "Version check:"
"${BIN}" version 2>/dev/null || "${BIN}" --version 2>/dev/null || echo "(trying help...)"
echo ""
echo "=== Ultra build ready: ${BIN} ==="
