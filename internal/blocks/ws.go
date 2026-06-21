package blocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var wsClients = struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]string
}{clients: make(map[*websocket.Conn]string)}

// WSA2CHandler handles A2C streaming via WebSocket.
func WSA2CHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil { return }
	defer conn.Close()

	agentID := r.URL.Query().Get("agent")
	if agentID == "" { agentID = "unknown" }

	wsClients.mu.Lock()
	wsClients.clients[conn] = agentID
	wsClients.mu.Unlock()

	defer func() {
		wsClients.mu.Lock()
		delete(wsClients.clients, conn)
		wsClients.mu.Unlock()
	}()

	Log(LogInfo, "ws.connect", agentID, "", "", 0, nil)

	// Read loop — receive messages from client
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil { break }
		// Echo back as ACK with agent context
		resp := A2CEvent{AgentID: agentID, Type: "ack", Content: string(msg), Seq: -1}
		data, _ := json.Marshal(resp)
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

// WSBroadcast sends an event to all connected WebSocket clients.
func WSBroadcast(event A2CEvent) {
	wsClients.mu.Lock()
	defer wsClients.mu.Unlock()

	data, _ := json.Marshal(event)
	for conn := range wsClients.clients {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

// WSStatus returns WebSocket status.
func WSStatus() string {
	wsClients.mu.Lock()
	defer wsClients.mu.Unlock()
	return fmt.Sprintf("ws: %d clients", len(wsClients.clients))
}
