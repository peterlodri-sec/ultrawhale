# Contributing to ultrawhale

## Good First Issues

- Add a new block primitive (see `docs/primitive-mapping.md`)
- Add a new Vaked .vaked declaration example
- Improve AG-UI theme colors
- Add benchmarks for a block

## Architecture

See `docs/primitive-mapping.md` for the 53-block → 7 Vaked layer mapping.

## Development

```sh
git clone https://github.com/peterlodri-sec/ultrawhale.git
cd ultrawhale
go build ./cmd/whale
task dev-deploy
```

## Code Style

- All blocks must carry POV context (`_ = CurrentPOV()`)
- Operations must be journaled via `blocks.Write()`
- New commands: add to `internal/tui/reload.go` + `model_prompt.go`

## RFC Process

Use `.dev/agents/meta-architect.md` for architectural thinking.
Submit RFCs as GitHub issues with label `rfc`.
