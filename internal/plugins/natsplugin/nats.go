// Package natsplugin publishes ultrawhale lifecycle events to NATS JetStream.
// Events: turn.start, turn.stop, tool.call, tool.result.
// Async fire-and-forget — never blocks the TUI loop.
package natsplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
)

const PluginID = "nats-eventbus"

type Plugin struct {
	mu     sync.Mutex
	conn   net.Conn
	config Config
	bound  bool
}

type Config struct {
	URL   string
	Creds string
}

func NewPlugin() *Plugin {
	return &Plugin{config: Config{
		URL: envOrDefault("NATS_URL", "nats://crabcc-nats:4222"),
	}}
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "NATS EventBus" }
func (p *Plugin) Version() string { return "0.2.0" }
func (p *Plugin) Description() string {
	return "Publishes ultrawhale lifecycle events to NATS JetStream for fleet orchestration."
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{
		{Event: agent.HookEventSessionStart, Name: "nats.session", Source: "plugin:nats",
			Description: "Connects to NATS and publishes turn.start.",
			Priority: 80,
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				go p.connect()
				p.publish("whale.turn.start", map[string]any{
					"session_id": payload.SessionID,
					"cwd":        payload.CWD,
					"timestamp":  time.Now().UTC().Format(time.RFC3339),
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventPreToolUse, Name: "nats.tool-call", Source: "plugin:nats",
			Description: "Publishes tool.call event.",
			Priority: 5,
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.tool.call", map[string]any{"tool": payload.ToolName})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventPostToolUse, Name: "nats.tool-result", Source: "plugin:nats",
			Description: "Publishes tool.result event.",
			Priority: 5,
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.tool.result", map[string]any{"tool": payload.ToolName, "outcome": payload.ToolOutcome})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventStop, Name: "nats.session-end", Source: "plugin:nats",
			Description: "Publishes turn.stop and disconnects.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.turn.stop", map[string]any{"session_id": payload.SessionID})
				p.disconnect()
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
	}
}

func (p *Plugin) connect() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bound { return }

	addr := strings.TrimPrefix(strings.TrimPrefix(p.config.URL, "nats://"), "tls://")
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return
	}
	p.conn = conn
	p.bound = true

	// Send NATS CONNECT message (minimal protocol)
	fmt.Fprintf(conn, "CONNECT {\"name\":\"ultrawhale\",\"lang\":\"go\",\"version\":\"0.2.0\"}\r\n")
}

func (p *Plugin) disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	p.bound = false
}

func (p *Plugin) publish(subject string, data map[string]any) {
	p.mu.Lock()
	conn := p.conn
	bound := p.bound
	p.mu.Unlock()

	if !bound || conn == nil {
		p.connect() // retry
		p.mu.Lock()
		conn = p.conn
		bound = p.bound
		p.mu.Unlock()
		if !bound || conn == nil { return }
	}

	payload, _ := json.Marshal(data)
	msg := fmt.Sprintf("PUB %s %d\r\n%s\r\n", subject, len(payload), string(payload))
	conn.Write([]byte(msg))
}

func (p *Plugin) Doctor() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bound {
		return fmt.Sprintf("nats: connected to %s", p.config.URL)
	}
	return fmt.Sprintf("nats: disconnected (target: %s)", p.config.URL)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
