#!/bin/bash
# Pre-push CI check — validate BEFORE push
set -e

echo "▸ Pre-push CI check"

# Validate YAML
echo "  YAML..."
python3 -c "import yaml, glob; [yaml.safe_load(open(f)) for f in glob.glob('.github/workflows/*.yml')]" 2>/dev/null && echo "    ✅ valid" || echo "    ❌ invalid"

# Test blocks
echo "  Test..."
go test -count=1 -run TestReadWrite ./internal/blocks/ > /dev/null 2>&1 && echo "    ✅ PASS" || echo "    ❌ FAIL"

# Build
echo "  Build..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /dev/null ./cmd/whale > /dev/null 2>&1 && echo "    ✅ OK" || echo "    ❌ FAIL"

echo "✅ Pre-push OK"
