# Whale

<p align="center">
  <img src="docs/logo.svg" alt="Whale — AI coding agent for DeepSeek, in any environment" width="640">
</p>

<p align="center">
  <a href="./README.zh.md">简体中文</a> · <strong>English</strong>
</p>

<p align="center">
  <a href="https://github.com/usewhale/DeepSeek-Code-Whale/releases"><img src="https://img.shields.io/github/v/release/usewhale/DeepSeek-Code-Whale?label=release" alt="release"></a>
  <a href="https://www.npmjs.com/package/@usewhale/whale"><img src="https://img.shields.io/npm/v/@usewhale/whale" alt="npm"></a>
  <a href="https://github.com/usewhale/DeepSeek-Code-Whale/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/usewhale/DeepSeek-Code-Whale/ci.yml?label=CI" alt="CI"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/github/license/usewhale/DeepSeek-Code-Whale" alt="license"></a>
  <a href="https://github.com/usewhale/DeepSeek-Code-Whale/stargazers"><img src="https://img.shields.io/github/stars/usewhale/DeepSeek-Code-Whale?style=flat&logo=github&label=stars" alt="GitHub stars"></a>
  <img src="https://img.shields.io/badge/prompt%20cache-98%25-brightgreen" alt="98% prompt cache hit">
</p>

<p align="center">
  Blazingly fast · ~98% prompt cache hit · Zero bloat
</p>

<p align="center">
  <b>Whale — AI coding agent for DeepSeek, in any environment.</b><br>
  Long context, tools, and programmable workflows —<br>
  start in the terminal, scale to desktop and beyond.
</p>

---

## 🚀 Quick Start

Any platform:

```bash
npm install -g @usewhale/whale
```

macOS:

```bash
brew install usewhale/tap/whale
```

Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/usewhale/DeepSeek-Code-Whale/main/scripts/install.sh | sh
```

Windows PowerShell:

Requires Windows 10 or Windows Server 2016 or later.

```powershell
irm https://raw.githubusercontent.com/usewhale/DeepSeek-Code-Whale/main/scripts/install.ps1 | iex
```

Windows CMD:

```cmd
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/usewhale/DeepSeek-Code-Whale/main/scripts/install.ps1 | iex"
```

```bash
# Set your DeepSeek API key
whale setup

# Launch the interactive TUI
whale
```

That's it. Type your question and Whale starts working — reading files, running commands,
editing code, searching the web.

> Need a different model provider, proxy, or custom config? See [Configuration](docs/configuration.en.md).

---

## ✨ At a Glance

| What | Why it matters |
|---|---|
| 💰 **~98% prompt cache hit** | Whale reuses cached context aggressively — most prompts hit cache, slashing costs to pennies per session. DeepSeek pricing × Whale caching = AI-assisted coding at scale. |
| 🐋 **DeepSeek-native** | Built for DeepSeek's long context (1M tokens), tool calling, and cost efficiency — no generic multi-model wrapper |
| 🔁 **Dynamic Workflows** | Write JavaScript scripts that orchestrate multiple agents — fan-out research, multi-perspective review, pipelines. Claude Code compatible. |
| 🔌 **MCP** | Plug in 1,000+ MCP servers for extended tooling — file ops, shell, git, web, and more |
| 🧩 **Skills + Plugins** | Install community skills (code review, git workflows, etc.) or write your own |

---

## 🔁 Dynamic Workflows

Whale's **Dynamic Workflows** let you script multi-agent orchestration in JavaScript:

```js
// .whale/workflows/research.js
const results = await parallel([
  () => agent("Search for best practices in Go error handling"),
  () => agent("Find common Go error handling mistakes"),
]);
return agent("Synthesize both findings into a concise guide");
```

**Fan-out research · Multi-perspective review · Pipeline processing · Adversarial validation**

> ✅ **Claude Code compatible** — workflow scripts written for Claude Code work as-is in Whale.

> **⚠️ Disabled by default** — run `/config` in the TUI and enable `Dynamic workflows`, or add `[workflows] enabled = true` to `.whale/config.local.toml`.

Learn more: [Workflow Overview](docs/workflows.en.md) · [Custom Workflow Guide](docs/custom-workflows.en.md)

---

## 🧰 MCP, Skills & Plugins

| Extension | What it does | Get started |
|---|---|---|
| **MCP Servers** | Connect to 1,000+ tools (databases, APIs, browser automation) | [docs/mcp.en.md](docs/mcp.en.md) |
| **Skills** | Load domain expertise — code review, git-worktree, and more | [docs/skills.en.md](docs/skills.en.md) |
| **Subagents** | Define focused child-agent roles such as reviewers or researchers | [docs/agents.en.md](docs/agents.en.md) |
| **Plugins** | Extend Whale's runtime with custom logic | [docs/plugins.en.md](docs/plugins.en.md) |
| **Hooks** | Run scripts on lifecycle events | [docs/hooks.en.md](docs/hooks.en.md) |

---

## 📸 How It Works

Whale currently offers three interfaces — with more environments on the way:

| Interface | When to use |
|---|---|
| **`whale`** (TUI) | Interactive coding sessions — chat, review, iterate with full context |
| **`whale ask "..."`** (CLI) | One-shot questions, quick code reviews, single commands |
| **`whale --headless`** | CI/CD, automated PR reviews, scheduled tasks |

---

## 🎯 Non-goals

- **Multi-model shell.** Whale is DeepSeek-first — optimized for DeepSeek's caching, tools, and pricing.
- **IDE replacement.** Whale is not an IDE — it's an agent that meets you wherever you code: terminal, desktop, or CI.

## 📦 Project Status

Whale is in active development. Best suited for personal projects, experimental repositories,
and workflows where changes can be reviewed and rolled back.

> **Disclaimer:** This project is not affiliated with DeepSeek Inc. It is an independent open-source community project.

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for local development, testing, issues, and PRs.

Current direction and available tasks: [ROADMAP.md](ROADMAP.md).

Security issues: [SECURITY.md](SECURITY.md).

---

## Star History

<a href="https://www.star-history.com/?repos=usewhale%2FDeepSeek-Code-Whale&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=usewhale/DeepSeek-Code-Whale&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=usewhale/DeepSeek-Code-Whale&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=usewhale/DeepSeek-Code-Whale&type=date&legend=top-left" />
 </picture>
</a>

---

## 🙏 Credits

Whale stands on the shoulders of giants:

- [Charmbracelet](https://charm.sh) — Bubble Tea, Lip Gloss, Glamour, and the entire TUI ecosystem
- [fastschema/qjs](https://github.com/fastschema/qjs) — QuickJS Go bindings for workflow scripting
- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- [alecthomas/chroma](https://github.com/alecthomas/chroma) — Syntax highlighting
- [yuin/goldmark](https://github.com/yuin/goldmark) — Markdown parsing
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) — MCP protocol support
- [tetratelabs/wazero](https://github.com/tetratelabs/wazero) — Pure-Go WebAssembly runtime

And the many open-source libraries we depend on — thank you.

## Blocks Engine Benchmarks (v1.8.0)

| Benchmark | Result | Notes |
|-----------|--------|-------|
| Hash 64KB (Go) | 1,464 MB/s | `crypto/sha256` stdlib, auto-SHA-NI |
| Hash 64KB (Asm) | 1,524 MB/s | AVX2+SHA-NI assembly |
| Hash 64KB (GPU) | 1,548 MB/s | Metal stub (Apple Silicon) |
| Write 64KB | 596 MB/s | I/O bound by filesystem |
| Read 64KB | 11.7µs | In-memory after first read |
| Batch-64 files | 3.8ms | 64 files atomic |
| Batch-256 files | 13.9ms | Parallel goroutine dispatch |
| Sed 1KB | 3,972ns | 257 MB/s, 896 B alloc |
| Sed 1MB | 4.17ms | 252 MB/s, 877 KB alloc |
| SedFile | 7.25µs | Read→sed→write journaled |
| BlockLifecycle | 547µs | Write→Write→Rollback→Read cycle |
| Concurrent writes | 3,200 w/0 err | 32 workers × 100 ops |

### macOS ARM64 Assembly

Apple Silicon (M1+) uses ARMv8 crypto extensions via `hash_arm64.s`:
- SHA256H + SHA256H2 instructions for hardware sha256
- NEON SIMD for parallel block processing
- Same tier as Linux AVX2+SHA-NI — ~1.5 GB/s
- Metal Performance Shaders stub for GPU offload (batch >64 files)
