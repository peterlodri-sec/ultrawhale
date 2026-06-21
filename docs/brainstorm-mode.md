# Brainstorm Mode — Turn-Based Co-Creation

## What It Is

A first-class UI mode for deep thinking sessions between human and LLM.
Turn-based dialog. Multi-choice cards. SACRED surface protected.

## Modes

| Mode | Description |
|------|-------------|
| `qa` | Question → Answer → Follow-up |
| `multichoice` | Question → Options → Selection → Next |
| `freeform` | Open-ended dialog, no structure |
| `debate` | Position → Counter-position → Resolution |

## Commands

```
/brainstorm start "what is consciousness?"
/brainstorm list
/brainstorm resume <id>
/brainstorm complete <id>
```

## API

```
POST /api/v1/brainstorm  {"topic": "...", "mode": "freeform"}
GET  /api/v1/brainstorm  → all sessions
```

## SACRED Guarantees

- Always visible: the brainstorm UI card cannot be obscured
- Always direct: 1:1 human↔LLM, no proxy
- Always bidirectional: human types, LLM responds, human sees
- Persistence: completed sessions saved to brain long-term memory
