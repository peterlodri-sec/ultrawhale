package blocks

import (
	"fmt"
	"sync"
)

// ── A2UI Protocol — Agent-to-UI Streaming ────────────────────────────
// v20: SOTA protocol. Agents stream UI events directly to the TUI surface.
// Enables live ChatBlock updates, card renders, layer status changes.

// A2UIEvent is a UI event from an agent to the TUI.
type A2UIEvent struct {
	AgentID  string `json:"agent_id"`
	Type     string `json:"type"`     // "chat_block", "card", "layer_update", "toast", "vaked_one_shot"
	Content  string `json:"content"`
	Layer    string `json:"layer,omitempty"`    // Vaked layer (if layer_update)
	BlockID  string `json:"block_id,omitempty"` // AG-UI block ID (if chat_block)
	Streaming bool  `json:"streaming"`          // true if content is still arriving
}

// A2UIRouter routes UI events from agents to the TUI.
type A2UIRouter struct {
	mu       sync.Mutex
	handlers map[string]A2UIHandler
}

// A2UIHandler processes an A2UI event.
type A2UIHandler func(event A2UIEvent)

var a2uiRouter = &A2UIRouter{handlers: make(map[string]A2UIHandler)}

// RegisterA2UIHandler registers a handler for an A2UI event type.
func RegisterA2UIHandler(eventType string, handler A2UIHandler) {
	a2uiRouter.mu.Lock()
	defer a2uiRouter.mu.Unlock()
	a2uiRouter.handlers[eventType] = handler
}

// EmitA2UI sends a UI event from an agent to the TUI.
func EmitA2UI(event A2UIEvent) {
	a2uiRouter.mu.Lock()
	handler, ok := a2uiRouter.handlers[event.Type]
	a2uiRouter.mu.Unlock()

	if ok {
		handler(event)
	}
	Log(LogInfo, "a2ui."+event.Type, fmt.Sprintf("%s → UI", event.AgentID[:8]),
		"", "", 0, nil)
}

// A2UIStatus returns compact A2UI router status.
func A2UIStatus() string {
	a2uiRouter.mu.Lock()
	defer a2uiRouter.mu.Unlock()
	return fmt.Sprintf("a2ui: %d event handlers", len(a2uiRouter.handlers))
}
