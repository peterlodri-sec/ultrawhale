# Quick Start

```sh
git clone https://github.com/peterlodri-sec/ultrawhale.git
cd ultrawhale
go build -o bin/ultrawhale ./cmd/whale
bin/ultrawhale --model deepseek-v4-flash -w
```

## Interactive Setup

```sh
bin/ultrawhale-setup
```

## Build from source

```sh
# macOS
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/ultrawhale ./cmd/whale

# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -o bin/ultrawhale-linux ./cmd/whale
```

## Run benchmarks

```sh
task bench
# or
go test -bench=. -benchmem ./internal/blocks/
```
