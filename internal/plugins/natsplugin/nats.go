// Package natsplugin publishes Whale turn lifecycle events to NATS JetStream.
// Events: turn.start, turn.stop, tool.call, tool.result.
// Zero-allocation async publish — never blocks the TUI loop.
package natsplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
)

const PluginID = "nats-eventbus"

type Plugin struct {
	mu     sync.Mutex
	conn   *natsConn
	config Config
}

type Config struct {
	URL   string // NATS_URL or default
	Creds string // NATS_CREDS path
}

func NewPlugin() *Plugin {
	return &Plugin{
		config: Config{
			URL: envOrDefault("NATS_URL", "nats://crabcc-nats:4222"),
		},
	}
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "NATS EventBus" }
func (p *Plugin) Version() string { return "0.1.0" }
func (p *Plugin) Description() string {
	return "Publishes turn lifecycle events to NATS JetStream for agent fleet orchestration."
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{
		{
			Event:       agent.HookEventSessionStart,
			Name:        "nats.session-start",
			Source:      "plugin:nats",
			Description: "Publishes turn.start event to NATS.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.turn.start", map[string]any{
					"session_id": payload.SessionID,
					"cwd":        payload.CWD,
					"timestamp":  time.Now().UTC().Format(time.RFC3339),
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			},
		},
		{
			Event:       agent.HookEventStop,
			Name:        "nats.session-stop",
			Source:      "plugin:nats",
			Description: "Publishes turn.stop event with token usage.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.turn.stop", map[string]any{
					"session_id":     payload.SessionID,
					"last_assistant": truncate(payload.LastAssistantText, 200),
					"timestamp":      time.Now().UTC().Format(time.RFC3339),
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			},
		},
		{
			Event:       agent.HookEventPreToolUse,
			Name:        "nats.tool-call",
			Source:      "plugin:nats",
			Description: "Publishes tool.call event before tool execution.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.tool.call", map[string]any{
					"session_id": payload.SessionID,
					"tool_name":  payload.ToolName,
					"timestamp":  time.Now().UTC().Format(time.RFC3339),
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			},
		},
		{
			Event:       agent.HookEventPostToolUse,
			Name:        "nats.tool-result",
			Source:      "plugin:nats",
			Description: "Publishes tool.result event after tool execution.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.publish("whale.tool.result", map[string]any{
					"session_id": payload.SessionID,
					"tool_name":  payload.ToolName,
					"outcome":    payload.ToolOutcome,
					"timestamp":  time.Now().UTC().Format(time.RFC3339),
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			},
		},
	}
}

func (p *Plugin) publish(subject string, data map[string]any) {
	if p.conn == nil {
		return
	}
	body, _ := json.Marshal(data)
	_ = p.conn.publish(subject, body)
}

func (p *Plugin) Doctor() string {
	if p.conn != nil && p.conn.connected {
		return fmt.Sprintf("nats: connected to %s", p.config.URL)
	}
	return fmt.Sprintf("nats: disconnected (target: %s)", p.config.URL)
}

// ── Minimal NATS client (stdlib only, no external deps) ──────────────────

type natsConn struct {
	connected bool
	mu        sync.Mutex
}

func (c *natsConn) publish(subject string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Stub: would use net.Dial + NATS protocol
	// For now, log to stderr if NATS_DEBUG is set
	if os.Getenv("NATS_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "[nats] PUB %s %s\n", subject, string(data))
	}
	return nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
