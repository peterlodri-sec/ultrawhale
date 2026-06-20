// Package langfuseplugin sends hierarchical LLM traces to Langfuse.
// Trace: ultrawhale-{version}-session-{id} → Span: tool:{name} → Event: error
// Batch ingestion every 2s or 64 events, flush on session stop.
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
	"github.com/usewhale/whale/internal/blocks"
	"github.com/usewhale/whale/internal/build"
)

const PluginID = "langfuse-telemetry"

type Plugin struct {
	mu        sync.Mutex
	config    Config
	client    *http.Client
	queue     []ingestionEvent
	sessionID string
	traceID   string
	version   string
}

type Config struct {
	Host      string
	PublicKey string
	SecretKey string
}

type ingestionEvent struct {
	Type       string         `json:"type"`
	Timestamp  time.Time      `json:"timestamp"`
	Body       map[string]any `json:"body"`
}

func NewPlugin() *Plugin {
	p := &Plugin{
		config: Config{
			Host:      envOrDefaultLang("LANGFUSE_HOST", "https://langfuse.crabcc.app"),
			PublicKey: os.Getenv("LANGFUSE_PUBLIC_KEY"),
			SecretKey: os.Getenv("LANGFUSE_SECRET_KEY"),
		},
		client:  &http.Client{Timeout: 5 * time.Second},
		queue:   make([]ingestionEvent, 0, 64),
		version: build.CurrentVersion(),
	}
	go p.flushLoop()
	return p
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "Langfuse Telemetry" }
func (p *Plugin) Version() string { return "0.3.0" }
func (p *Plugin) Description() string {
	return "Hierarchical LLM traces to Langfuse — batch ingestion, auto-flush."
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{
		{Event: agent.HookEventSessionStart, Name: "langfuse.session", Source: "plugin:langfuse",
			Description: "Opens ultrawhale-{version}-session-{id} trace.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.sessionID = payload.SessionID
				sid := payload.SessionID
				if len(sid) > 8 { sid = sid[:8] }
				p.traceID = fmt.Sprintf("ultrawhale-%s-session-%s", p.version, sid)
				pov := blocks.CurrentPOV()
				p.enqueue("trace-create", map[string]any{
					"id":   p.traceID,
					"name": p.traceID,
					"metadata": map[string]string{
						"version":    p.version,
						"session_id": payload.SessionID,
						"machine":    pov.Machine,
						"arch":       pov.Arch,
						"tier":       pov.Tier,
					},
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventPreToolUse, Name: "langfuse.tool-call", Source: "plugin:langfuse",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.enqueue("observation-create", map[string]any{
					"traceId": p.traceID,
					"type":    "SPAN",
					"name":    "tool:" + payload.ToolName,
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventError, Name: "langfuse.error", Source: "plugin:langfuse",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.enqueue("event-create", map[string]any{
					"traceId": p.traceID,
					"name":    "error:" + payload.ToolName,
					"metadata": map[string]string{"error": payload.ToolOutcome},
				})
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventStop, Name: "langfuse.flush", Source: "plugin:langfuse",
			Priority: 95, // flush before other cleanup
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.enqueue("score-create", map[string]any{
					"traceId": p.traceID,
					"name":    "session-complete",
					"value":   1,
				})
				p.flush() // synchronously flush on stop
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
	}
}

func (p *Plugin) enqueue(typ string, body map[string]any) {
	p.mu.Lock()
	p.queue = append(p.queue, ingestionEvent{Type: typ, Timestamp: time.Now(), Body: body})
	shouldFlush := len(p.queue) >= 64
	p.mu.Unlock()
	if shouldFlush {
		go p.flush()
	}
}

func (p *Plugin) flush() {
	p.mu.Lock()
	if len(p.queue) == 0 || p.config.PublicKey == "" {
		p.mu.Unlock()
		return
	}
	batch := p.queue
	p.queue = make([]ingestionEvent, 0, 64)
	p.mu.Unlock()

	body, _ := json.Marshal(map[string]any{"batch": batch})
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
	p.mu.Lock()
	queued := len(p.queue)
	p.mu.Unlock()
	if p.config.PublicKey == "" {
		return "langfuse: not configured (set LANGFUSE_PUBLIC_KEY)"
	}
	return fmt.Sprintf("langfuse: %s, project: ultrawhale, version: %s, queued: %d", p.config.Host, p.version, queued)
}

func envOrDefaultLang(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
