package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── A2A Protocol — Agent-to-Agent Wire ───────────────────────────────
// Agents communicate via NATS subjects: whale.a2a.{agentID}.{action}
// Supported actions: ping, delegate, result, status

// A2AMessage is a wire-format agent-to-agent message.
type A2AMessage struct {
	From      string    `json:"from"`      // sender agent ID
	To        string    `json:"to"`        // recipient agent ID (or "*" for broadcast)
	Action    string    `json:"action"`    // "ping", "delegate", "result", "status"
	Payload   string    `json:"payload"`   // JSON-encoded task or result
	Timestamp time.Time `json:"timestamp"`
	Ref       string    `json:"ref"`       // sha256 of payload
}

// A2ARouter handles agent-to-agent message routing.
type A2ARouter struct {
	mu       sync.Mutex
	handlers map[string]A2AHandler // action → handler
	store    *AgentStore
}

// A2AHandler processes an A2A message.
type A2AHandler func(msg A2AMessage) A2AMessage

var a2aRouter = &A2ARouter{
	handlers: make(map[string]A2AHandler),
	store:    agentsStore,
}

// RegisterA2AHandler registers a handler for an A2A action.
func RegisterA2AHandler(action string, handler A2AHandler) {
	a2aRouter.mu.Lock()
	defer a2aRouter.mu.Unlock()
	a2aRouter.handlers[action] = handler
}

// RouteA2A routes an A2A message to its handler.
func RouteA2A(msg A2AMessage) (A2AMessage, error) {
	a2aRouter.mu.Lock()
	handler, ok := a2aRouter.handlers[msg.Action]
	a2aRouter.mu.Unlock()

	if !ok {
		return A2AMessage{}, fmt.Errorf("a2a: no handler for action %s", msg.Action)
	}

	return handler(msg), nil
}

// SendA2A sends an agent-to-agent message.
func SendA2A(from, to, action, payload string) A2AMessage {
	msg := A2AMessage{
		From:      from,
		To:        to,
		Action:    action,
		Payload:   payload,
		Timestamp: time.Now(),
		Ref:       Ref([]byte(payload)),
	}

	// In production: publish to NATS subject whale.a2a.{to}.{action}
	// For now: route locally
	response, _ := RouteA2A(msg)
	Log(LogInfo, "a2a."+action, fmt.Sprintf("%s → %s", from, to), msg.Ref, "", 0, nil)
	return response
}

// ── Built-in A2A handlers ─────────────────────────────────────────────

func init() {
	RegisterA2AHandler("ping", func(msg A2AMessage) A2AMessage {
		return A2AMessage{
			From: msg.To, To: msg.From,
			Action: "pong", Payload: `{"status":"alive"}`,
			Timestamp: time.Now(),
		}
	})

	RegisterA2AHandler("status", func(msg A2AMessage) A2AMessage {
		agent := GetAgent(msg.To)
		if agent == nil {
			return A2AMessage{Action: "error", Payload: "agent not found"}
		}
		return A2AMessage{
			From: msg.To, To: msg.From,
			Action: "status_response",
			Payload: fmt.Sprintf(`{"id":"%s","role":"%s","status":"%s","tools":%d}`,
				agent.ID, agent.Role, agent.Status, agent.ToolCalls),
			Timestamp: time.Now(),
		}
	})
}

// A2AStatus returns compact A2A router status.
func A2AStatus() string {
	a2aRouter.mu.Lock()
	defer a2aRouter.mu.Unlock()
	return fmt.Sprintf("a2a: %d handlers registered, %d agents in store",
		len(a2aRouter.handlers), AgentCount())
}
