# Protocol Mapping — ultrawhale v66.0.0

All protocols, their status functions, VakedFit functions, and commands.

| Protocol | Status Function | VakedFit Function | /cmd | Wired? |
|----------|----------------|-------------------|------|--------|
| SSH | `SSHStatus()` (ssh.go) | `SSHVakedFit()` | — | ✅ |
| GPG | — (external) | — | — | ✅ key 2B2495 |
| HF Webhook | `HFWebhookStatus()` | `HFWebhookVakedFit()` | `/hf` | ✅ |
| RADIO | `RadioStatus()` | `RadioVakedFit()` | `/radio` | ✅ |
| A2A | `A2AStatus()` | — | `/a2a` | ✅ |
| A2C | `A2CStatus()` | — | `/a2c` | ✅ |
| A2UI | `A2UIStatus()` | — | — | ✅ |
| MCP | `MCPStatus()` | — | — | ✅ |
| VFS | `VFSStatus()` | — | `/vfs` | ✅ |
| Git | `GitPrimitiveStatus()` | `GitPrimitiveVakedFit()` | `/git` | ✅ |
| Live Session | `LiveSessionStatus()` | `LiveSessionVakedFit()` | `/session`, `/who` | ✅ |
| WebSocket | `WSStatus()` | — | — | ✅ |

**12 protocols. All wired. All have status. All have commands.**
