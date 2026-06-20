package langfuseplugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/build"
)

const PluginID = "langfuse-telemetry"

type Plugin struct {
	mu        sync.Mutex
	config    Config
	client    *http.Client
	queue     []event
	sessionID string
	traceID   string
	version   string
}

type Config struct {
	Host      string
	PublicKey string
	SecretKey string
}

type event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      map[string]any `json:"data"`
}

func NewPlugin() *Plugin {
	p := &Plugin{
		config: Config{
			Host:      envOrDefault("LANGFUSE_HOST", "https://langfuse.crabcc.app"),
			PublicKey: os.Getenv("LANGFUSE_PUBLIC_KEY"),
			SecretKey: os.Getenv("LANGFUSE_SECRET_KEY"),
		},
		client:  &http.Client{Timeout: 5 * time.Second},
		queue:   make([]event, 0, 64),
		version: build.CurrentVersion(),
	}
	go p.flushLoop()
	return p
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "Langfuse Telemetry" }
func (p *Plugin) Version() string { return "0.2.0" }
func (p *Plugin) Description() string {
	return "Sends hierarchical LLM traces to Langfuse: ultrawhale-{version} → session → turn → tool"
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{
		{Event: agent.HookEventSessionStart, Name: "langfuse.session-start", Source: "plugin:langfuse",
			Description: "Opens a Langfuse trace: ultrawhale-{version}-session-{id}",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.sessionID = payload.SessionID
				// Trace name: ultrawhale-v1.0.0-session-a1b2c3d4
				sid := payload.SessionID
				if len(sid) > 8 { sid = sid[:8] }
				p.traceID = fmt.Sprintf("ultrawhale-%s-session-%s", p.version, sid)
				p.enqueue("trace-create", map[string]any{
					"id":      p.traceID,
					"name":    p.traceID,
					"metadata": map[string]string{
						"version":    p.version,
						"session_id": payload.SessionID,
					},
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventPreToolUse, Name: "langfuse.tool-call", Source: "plugin:langfuse",
			Description: "Records tool call as a span under the session trace.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.enqueue("observation-create", map[string]any{
					"traceId": p.traceID,
					"type":    "SPAN",
					"name":    "tool:" + payload.ToolName,
					"metadata": map[string]string{
						"session_id": payload.SessionID,
					},
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventError, Name: "langfuse.error", Source: "plugin:langfuse",
			Description: "Records errors as events under the session trace.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.enqueue("event-create", map[string]any{
					"traceId": p.traceID,
					"name":    "error:" + payload.ToolName,
					"metadata": map[string]string{
						"error": payload.ToolOutcome,
					},
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventStop, Name: "langfuse.session-end", Source: "plugin:langfuse",
			Description: "Flushes all pending Langfuse events.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.enqueue("score-create", map[string]any{
					"traceId": p.traceID,
					"name":    "session-complete",
					"value":   1,
				})
				p.flush()
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
	}
}

func (p *Plugin) enqueue(typ string, data map[string]any) {
	p.mu.Lock()
	p.queue = append(p.queue, event{Type: typ, Timestamp: time.Now(), Data: data})
	p.mu.Unlock()
}

func (p *Plugin) flush() {
	p.mu.Lock()
	events := p.queue
	p.queue = make([]event, 0, 64)
	p.mu.Unlock()
	if len(events) == 0 || p.config.PublicKey == "" { return }
	body, _ := json.Marshal(map[string]any{"batch": events})
	req, _ := http.NewRequest("POST", p.config.Host+"/api/public/ingestion", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.config.PublicKey, p.config.SecretKey)
	resp, _ := p.client.Do(req)
	if resp != nil { resp.Body.Close() }
}

func (p *Plugin) flushLoop() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for range t.C { p.flush() }
}

func (p *Plugin) Doctor() string {
	if p.config.PublicKey == "" { return "langfuse: not configured (set LANGFUSE_PUBLIC_KEY)" }
	return fmt.Sprintf("langfuse: %s, project: ultrawhale, version: %s, %d queued", p.config.Host, p.version, len(p.queue))
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
