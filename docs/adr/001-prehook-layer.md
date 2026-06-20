# ADR 001: Pre-Hook Layer

**Status:** Accepted  
**Date:** 2026-06-20  
**Version:** v7.1.0

## Context

Block operations (write, sed, git, deploy) need validation before execution.
Without pre-hooks, invalid operations fail mid-execution, leaving partial state.

## Decision

Implement a PreHook interface layer that runs BEFORE block operations.
Failures prevent the operation from executing entirely.

## Architecture

```
Operation → PreHook.Validate() → Block.Execute()
                ↓ FAIL
           return error (no state change)
```

## Pre-hooks

| Hook | Validates | Guards |
|------|-----------|--------|
| PreWrite | File size, path validity | Write() |
| PreCommit | gofmt + go vet | Git commits |
| PreSed | Pattern validity, dry-run | SedAll() |
| PreGit | Working tree dirty check | Git ops |
| PreDeploy | Binary exists + doctor | Deploy ops |

## Consequences

- Write() now calls PreWrite internally — all writes validated
- Deploy blocked if binary missing or doctor fails
- Sed operations get pattern validation + match count preview
