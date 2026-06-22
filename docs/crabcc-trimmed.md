# crabcc — Symbol Index for AI Agents

> Trimmed. Just the MCP. No Claude integration. No Cursor plugin. No ollama backend. Just the index and the protocol.

## What It Is

A **symbol index** for codebases. Builds a database of every symbol (function, type, variable, import) in a repo. Exposes it as an MCP server.

## Commands

```
crabcc index              # build the symbol database (~5-30s)
crabcc lookup sym <name>      # find symbol definition
crabcc lookup refs <name>     # find all references
crabcc lookup outline <file>  # structural outline (no reading whole file)
```

## MCP Tools Exposed

| Tool | Description |
|------|------------|
| `crabcc_lookup_sym` | Find a symbol by name |
| `crabcc_lookup_refs` | Find all references to a symbol |
| `crabcc_lookup_outline` | Get structural outline of a file |
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


## Integration Test — ultrawhale v100.1.0 ✅

```
crabcc index
  → 877 files · 12,530 symbols · 70,015 edges

crabcc lookup refs ASCIIBox
  → 12 references across 12 blocks

crabcc lookup outline doctor_primitive.go
  → 9 symbols: DoctorCheck, Doctor, DoctorRun, countOK...
```

## Actual CLI Surface (v6.2.0)

```
crabcc index              # build the symbol database
crabcc lookup sym <name>  # find symbol definition
crabcc lookup refs <name> # find all references
crabcc lookup outline <f> # structural outline
crabcc graph              # call-graph operations
crabcc memory             # AI memory (per-repo memory.db)
crabcc serve              # localhost call-graph viewer
```
