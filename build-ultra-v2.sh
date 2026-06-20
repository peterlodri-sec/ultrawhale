#!/usr/bin/env bash
# Ultra-performance Whale build — dev-cx53 (AMD EPYC-Rome)
#
# Why CGO_ENABLED=0:
#   Go's internal linker is faster than mold/gold/lld for pure-Go projects.
#   CGO=1 adds glibc dynamic dependency — no static binary possible without
#   musl cross-compile (and musl + nixpkgs mimalloc = ABI mismatch).
#   mimalloc requires CGO and only helps allocation-heavy C++ workloads;
#   Go's runtime allocator is well-tuned for goroutine/TUI patterns.
#
# Flags:
#   GOAMD64=v3     — AVX2, BMI2, FMA (EPYC-Rome sweet spot)
#   -trimpath       — reproducible builds
#   -ldflags="-s -w" — strip debug + DWARF
#   CGO_ENABLED=0   — static binary, fast link
set -euo pipefail

WHALE_DIR="/home/dev/whale"
VAKED_FLAKE="/home/dev/vaked-base"
BIN="${WHALE_DIR}/bin/whale"
VERSION="${VERSION:-$(date +%Y%m%d-%H%M)-ultra}"

echo "=== Whale Ultra Build ==="
echo "Target:    $(hostname)"
echo "Version:   ${VERSION}"
echo "ISA:       x86-64-v3"
echo "Linker:    Go internal (fastest for pure Go)"
echo "Allocator: Go runtime"
echo ""

cd "${WHALE_DIR}"
rm -rf bin/ .gocache/

START_NS=$(date +%s%N)

nix develop "${VAKED_FLAKE}#" --command bash -c "
  export CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 GOMAXPROCS=16
  cd ${WHALE_DIR}
  go build \
    -trimpath \
    -ldflags=\"-s -w -X github.com/usewhale/whale/internal/build.Version=${VERSION}\" \
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
echo "Static:  $(ldd ${BIN} 2>&1)"
echo ""

# Verify ISA level
echo "=== ISA v3 check ==="
COUNT=$(objdump -d "${BIN}" 2>/dev/null | grep -c "vperm2i128\|vpbroadcast\|vpternlog\|vpmaskmov" || echo 0)
echo "v3 instructions: ${COUNT}"
echo ""

echo "Version:"
"${BIN}" version 2>&1 || "${BIN}" --version 2>&1 || true
echo ""
echo "=== Done: ${BIN} ==="
