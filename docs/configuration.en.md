# Configuration

## 🚀 Quick Setup

The fastest way to get started:

```bash
whale setup
```

This saves your DeepSeek API key to `~/.whale/credentials.json`.

You can also use the environment variable (takes precedence):

```bash
DEEPSEEK_API_KEY=sk-... whale
```

Run `whale doctor` anytime to confirm your current setup.

---

## Common Tasks

### Use a different model / endpoint

```toml
# .whale/config.toml (project) or ~/.whale/config.toml (global)
[model]
provider = "openai-compatible"
model = "deepseek-chat"
base_url = "https://api.deepseek.com/v1"
```

Whale is DeepSeek-native, but you can point it at any OpenAI-compatible endpoint.
Other models may not support all features (tool calling, long context).

### Set up a proxy

```toml
[model]
http_proxy = "http://127.0.0.1:7890"
https_proxy = "http://127.0.0.1:7890"
```

Whale respects `$HTTP_PROXY` and `$HTTPS_PROXY` environment variables too.

### Customize the system prompt

```toml
[settings]
prompt = "You are a coding assistant that prefers Rust over Go."
```

The prompt is injected at the start of every new session.

### Project-level settings

```toml
# .whale/config.toml — share with your team via git
[model]
model = "deepseek-chat"
```

```toml
# .whale/config.local.toml — personal overrides, do not commit
[model]
model = "deepseek-reasoner"
```

Config files are merged: `defaults < global < project shared < project local < CLI flags/env`

### Disable specific tools

```toml
[disabled_tools]
tools = ["web_search", "web_fetch"]
```

---

## Reference

### Config file locations

| Path | Scope | Commit? |
|---|---|---|
| `~/.whale/config.toml` | Global — all projects | No |
| `.whale/config.toml` | Project — shared with team | Yes |
| `.whale/config.local.toml` | Project — personal overrides | No |

On Windows, the default global directory is `%USERPROFILE%\\.whale`.
Set `WHALE_HOME` to use a custom directory on any platform.

### All settings (`config.toml`)

```toml
[model]
provider = "deepseek"                  # or "openai-compatible"
model = "deepseek-chat"                # or "deepseek-reasoner"
base_url = "https://api.deepseek.com/v1"
http_proxy = ""                        # proxy for API calls
https_proxy = ""

[settings]
prompt = ""                            # custom system prompt prefix
max_tokens = 4096                      # max response tokens

[permissions]
allowed_directories = []               # restrict file access to these dirs

[permissions.mcp]
fs = "allow"                           # "allow" | "ask" | "deny" per MCP server

[disabled_tools]
tools = []                             # hide built-in tools by name

[mcp]
config_path = ""                       # custom MCP config path

[workflows]
max_concurrency = 3                    # parallel agent limit

[skills]
disabled = []                          # skills to hide
enabled = []                           # force-enable even if project disables

[plugins]
disabled = []                          # plugins to disable
enabled = []                           # force-enable

[[hooks.PreToolUse]]
match = "shell_run"                    # optional; only PreToolUse/PostToolUse match by tool name
command = ""                           # shell command before matching tool calls

[[hooks.PostToolUse]]
match = "shell_run"
command = ""                           # shell command after matching tool calls

[logging]
level = "info"                         # debug | info | warn | error
```

### Environment variables

| Variable | Overrides |
|---|---|
| `DEEPSEEK_API_KEY` | Credential in `~/.whale/credentials.json` |
| `WHALE_HOME` | Global data directory (`~/.whale`) |
| `HTTP_PROXY` / `HTTPS_PROXY` | Proxy settings in config |
| `WHALE_MCP_CONFIG` | MCP config file path |

### Shell hooks

Hooks run shell commands on lifecycle events:

```toml
[[hooks.PreToolUse]]
match = "shell_run"
command = "echo 'about to run a tool call'"

[[hooks.PostToolUse]]
match = "shell_run"
command = "echo 'tool finished'"

[[hooks.UserPromptSubmit]]
command = "echo '{\"decision\":\"pass\"}'"
```

Supported events are `PreToolUse`, `PermissionRequest`, `PostToolUse`, `PreCompact`, `PostCompact`, `SessionStart`, `UserPromptSubmit`, `SubagentStart`, `SubagentStop`, and `Stop`. All events run at runtime and surface hook started/completed status in the TUI.

Hooks can return JSON on stdout with fields like `decision`, `reason`, `updated_input`, or `additional_context` to influence Whale's behavior. `PreToolUse`, `PermissionRequest`, and `UserPromptSubmit` can block the next action; `PreCompact` `additional_context` is added to the compact summary prompt. Project hooks are untrusted until reviewed; run `/hooks` in the TUI to inspect installed hooks and `/hooks trust all`, `/hooks trust <key>`, `/hooks disable <key>`, or `/hooks enable <key>` to manage workspace hook state.

### Worktrees

Whale supports git worktrees for isolated feature development:

```bash
whale --worktree
whale exec --worktree
```

On exit, Whale removes a clean worktree automatically. Uncommitted changes
prompt you to keep or remove.

---

## Where is local state stored?

```
~/.whale/
├── credentials.json    # API keys
├── config.toml         # global config
├── mcp.json            # MCP server config
├── sessions/           # session history
└── usage.jsonl         # usage logs
```

Do not commit these files.

---

## Need help?

```bash
whale doctor     # check your setup
whale --help     # CLI reference
```
