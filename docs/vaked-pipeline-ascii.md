# Vaked Pipeline — ASCII+UI+TUI+A2UI Design

## The Pipeline

```
┌─────────────┐    ┌──────────┐    ┌──────────────┐    ┌──────────┐    ┌───────────┐    ┌──────────┐    ┌───────────┐
│ DECLARE     │───→│ ENGINE   │───→│ SUPERVISE    │───→│ ENFORCE  │───→│ TESTIFY   │───→│ INDEX    │───→│ REVEAL    │
│ schema      │    │ 60 blocks│    │ orchestrator │    │ prehooks │    │ probe     │    │ space    │    │ TUI       │
│ contract    │    │ write    │    │ ralph       │    │ sacred   │    │ predict   │    │ vfs      │    │ surface   │
│ capabilities│    │ sed      │    │ supervisor  │    │ perm     │    │ learn     │    │ crabcc   │    │ agui      │
└─────────────┘    └──────────┘    └──────────────┘    └──────────┘    └───────────┘    └──────────┘    └───────────┘
```

## TUI Layout

```
┌─────────────────────────────────────────────────────┐
│  ▸ultrawhale v30  │  baobab:8384  │  NATS:4222     │  ← InfraBar
├─────────────────────────────────────────────────────┤
│ ╔══ Vaked Layers ══╗                               │
│ ║ 📜 Declares      ║   ┌───────────────────────┐   │
│ ║ 🏗️ Materializes  ║   │  Chat View            │   │
│ ║ 🔄 Supervises    ║   │  User> build auth     │   │  ← Dashboard + Chat
│ ║ 🛡️ Enforces      ║   │  Agent> implementing  │   │
│ ║ 🔍 Testifies     ║   │  ...                  │   │
│ ║ 🗂️ Indexes       ║   └───────────────────────┘   │
│ ║ 👁️ Reveals       ║                               │
│ ╚══════════════════╝                               │
├─────────────────────────────────────────────────────┤
│  sacred: ● 1:1 · perm: GRANTED · engine: 0 ops     │  ← HUD
└─────────────────────────────────────────────────────┘
```

## A2UI Event Flow

```
Agent spawn      → A2UI toast           → ephemeral message in TUI
Probe result     → A2UI layer_update    → VakedDashboard refresh
Learn event      → A2UI layer_update    → VakedDashboard refresh
Agent complete   → A2UI layer_update    → Sidepanel update
OneShot complete → A2UI vaked_one_shot  → Chat view block
```
