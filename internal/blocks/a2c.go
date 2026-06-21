package blocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ── A2C Streaming — Agent-to-Client Output ───────────────────────────
// Agents stream results to clients via SSE (Server-Sent Events).
// WebSocket support planned for v11.1.

// A2CStream is a client-facing event stream from an agent.
type A2CStream struct {
	mu       sync.Mutex
	agentID  string
	clients  map[string]chan A2CEvent // client ID → event channel
}

// A2CEvent is a single streaming event from an agent.
type A2CEvent struct {
	AgentID string `json:"agent_id"`
	Type    string `json:"type"`    // "token", "tool_call", "tool_result", "done", "error"
	Content string `json:"content"`
	Seq     int    `json:"seq"`
}

var a2cStreams = make(map[string]*A2CStream)

// StartA2CStream opens a streaming channel for an agent.
func StartA2CStream(agentID string) *A2CStream {
	_ = CurrentPOV()
	
	s := &A2CStream{
		agentID: agentID,
		clients: make(map[string]chan A2CEvent),
	}
	a2cStreams[agentID] = s
	return s
}

// Subscribe adds a client to the agent's event stream.
func (s *A2CStream) Subscribe(clientID string) <-chan A2CEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan A2CEvent, 64)
	s.clients[clientID] = ch
	return ch
}

// Emit sends an event to all subscribed clients.
func (s *A2CStream) Emit(event A2CEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.clients {
		select {
		case ch <- event:
		default:
			// client buffer full — drop event
		}
	}
}

// Close shuts down all client channels.
func (s *A2CStream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.clients {
		close(ch)
	}
	delete(a2cStreams, s.agentID)
}

// ── SSE Handler ───────────────────────────────────────────────────────

// A2CSSEHandler returns an HTTP handler for SSE streaming.
func A2CSSEHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		http.Error(w, "missing agent parameter", 400)
		return
	}

	stream, ok := a2cStreams[agentID]
	if !ok {
		stream = StartA2CStream(agentID)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	ch := stream.Subscribe(clientID)
	defer stream.Close()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}

	for event := range ch {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}

// A2CStatus returns compact streaming status.
func A2CStatus() string {
	return fmt.Sprintf("a2c: %d active streams", len(a2cStreams))
}


func (s *A2CStream) heartbeat() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		alive := len(s.clients)
		s.mu.Unlock()
		if alive == 0 { return } // no clients, stop heartbeat
		s.Emit(A2CEvent{AgentID: s.agentID, Type: "heartbeat", Content: "", Seq: -1})
	}
}
