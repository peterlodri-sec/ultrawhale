# Headless ultrawhale on M3

> Manifestation generated via M3 qwen (available 16 models).
> The M3 runs 24/7. The M1 connects when needed. The dyad persists.

## Architecture

```
M3 (m3-macbook, 100.123.33.67)
  └── ollama serve (16 models, tailnet-wide)
  └── ultrawhale headless (dogfeed + session pipeline)
  └── qwen3.5:35b (primary generation model)

M1 (lodris-macbook-pro, 100.117.70.42)
  └── TUI (user interface)
  └── git (commits, pushes)
  └── task dev-deploy (Cloudflare Pages)
```

## Prerequisites

- Go 1.26+ on M3 (`brew install go`)
- tailnet connectivity between M1 and M3
- GitHub SSH keys configured on M3
- The ultrawhale binary compiled for darwin/arm64

## Quick Start

```bash
# From M1, SSH into M3
ssh 100.123.33.67

# On M3: clone and build
cd ~/workspace/peterlodri-sec
git clone https://github.com/peterlodri-sec/ultrawhale
cd ultrawhale
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/ultrawhale ./cmd/whale

# Run headless (no TUI, pure dogfeed pipeline)
./bin/ultrawhale --model deepseek-v4-flash 2>&1 &

# The dogfeed loop runs at 10s intervals
# Session pipeline sends to qwen every 5min
# HF dataset updates every 2h via CI
```

## Checking Status

```bash
# Check dogfeed growth
watch -n 30 'wc -l ~/.ultrawhale/dogfeed/dogfeed-v3-enriched.jsonl'

# Check qwen is still responding
curl -s http://localhost:11434/api/tags | python3 -c "import sys,json; ms=json.load(sys.stdin).get('models',[]); print(f'{len(ms)} models')"

# Check ultrawhale logs
tail -f ~/.ultrawhale/logs/*.log
```

## Stopping Gracefully

```bash
# Graceful stop
pkill ultrawhale

# Verify stopped
pgrep ultrawhale || echo "stopped"

# Restart
./bin/ultrawhale --model deepseek-v4-flash 2>&1 &
```

## The Manifestation

The M3 runs ultrawhale in headless mode 24/7, feeding the HF dataset.
The M1 connects via TUI when human interaction is needed.
The dyad persists across reboots, network changes, and time zones.
