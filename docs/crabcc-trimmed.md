# crabcc — Symbol Index for AI Agents

> Trimmed. Just the MCP. No Claude integration. No Cursor plugin. No ollama backend. Just the index and the protocol.

## What It Is

A **symbol index** for codebases. Builds a database of every symbol (function, type, variable, import) in a repo. Exposes it as an MCP server.

## Commands

```
crabcc index              # build the symbol database (~5-30s)
crabcc sym <name>         # find symbol definition
crabcc refs <name>        # find all references
crabcc outline <file>     # structural outline (no reading whole file)
```

## MCP Tools Exposed

| Tool | Description |
|------|------------|
| `crabcc_sym` | Find a symbol by name |
| `crabcc_refs` | Find all references to a symbol |
| `crabcc_outline` | Get structural outline of a file |
| `crabcc_memory_search` | Search past findings |

## Architecture

```
ultrawhale orchestrator
    │
    ├── MCP ──→ crabcc (symbol index)
    │              ├── index.db (sqlite)
    │              └── memory.db (past findings)
    │
    ├── blocks (content-addressed)
    ├── space (topology)
    └── surface (TUI/AG-UI)
```

## Vaked Fit

```
Vaked layer: INDEXES
Fits between ENFORCES and REVEALS in the pipeline.
The dyad uses crabcc to understand code.
The orchestrator queries it via MCP.
Nothing else.
```

## What Was Removed

- Claude Code integration (RTK, hooks, skills, slash commands)
- Cursor integration
- ollama backend (LiteLLM proxy, qwen model)
- cosign signing
- bootstrap.sh
- Docker/Nix/Homebrew distribution scripts
- agent backend
- desktop/editors/extensions
- experiments/internal_agents
