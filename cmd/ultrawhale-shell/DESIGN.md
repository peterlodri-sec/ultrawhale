# ultrawhale-shell — Remote Execution Daemon

## Design

A tiny (~5MB) Go static binary deployed to remote machines. Communicates
via localhost HTTP API, journaled via blocks engine, POV-tracked.

## Architecture

```
Local ultrawhale                    Remote machine
     │                                    │
     │ SSH (one-time bootstrap)           │
     ├───────────────────────────────────→│ deploy ultrawhale-shell binary
     │                                    │ start daemon on :9797
     │                                    │
     │ HTTP (localhost tunnel via SSH)    │
     ├───────────────────────────────────→│ POST /exec {command, pov, ref}
     │←───────────────────────────────────┤ {result, exit_code, journal_ref}
     │                                    │
     │ blocks.Log + journal.Push          │
```

## Deployment

```sh
# One-time bootstrap
ssh dev-cx53 'mkdir -p ~/.local/bin'
scp bin/ultrawhale-shell dev-cx53:~/.local/bin/
ssh dev-cx53 'ultrawhale-shell --daemon --port 9797 &'
```

## API

POST /exec
  Body: {"command": "docker ps", "pov": {...}, "timeout_sec": 30}
  Response: {"result": "...", "exit_code": 0, "ref": "sha256...", "duration_ms": 123}

GET /health
  Response: {"status": "ok", "version": "v4.10.0", "uptime": "2h34m"}

GET /journal
  Response: [...last 64 command entries...]

## Build

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -trimpath -ldflags="-s -w" -o bin/ultrawhale-shell ./cmd/ultrawhale-shell/
# Result: ~5MB static binary
```
