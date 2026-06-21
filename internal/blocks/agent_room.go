package blocks

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── Agent Room — Multi-Agent Conversation Space ──────────────────────
//
// Inspired by agentpipe (kevinelliott/agentpipe): multiple AI agents
// conversing in a shared "room". Folded into Vaked philosophy.
//
// Each room is a SPACE node. Each agent is a FOLDED participant.
// The conversation is a TIME-ordered sequence. The room IS the context.

// AgentRoom is a shared conversation space for multiple agents.
type AgentRoom struct {
	ID        string
	Topic     string
	Agents    []string          // agent IDs in the room
	Messages  []RoomMessage     // conversation history
	CreatedAt time.Time
	Active    bool
}

// RoomMessage is one message in an agent room conversation.
type RoomMessage struct {
	From      string    // agent ID or "human"
	Content   string
	Timestamp time.Time
	Ref       string    // SHA256 of content
}

// AgentRoomManager manages all agent rooms.
type AgentRoomManager struct {
	mu    sync.Mutex
	Rooms map[string]*AgentRoom
	Stats AgentRoomStats
}

// AgentRoomStats tracks room activity.
type AgentRoomStats struct {
	RoomsCreated   int64
	MessagesSent   int64
	AgentsJoined   int64
	FoldDepth      int // max fold depth across all rooms
}

var agentRoomManager = &AgentRoomManager{
	Rooms: make(map[string]*AgentRoom),
}

// ── Room Operations ──────────────────────────────────────────────────

// CreateRoom creates a new agent conversation room.
func CreateRoom(topic string) *AgentRoom {
	agentRoomManager.mu.Lock()
	defer agentRoomManager.mu.Unlock()

	room := &AgentRoom{
		ID:        fmt.Sprintf("room-%d", time.Now().Unix()),
		Topic:     topic,
		Agents:    make([]string, 0),
		Messages:  make([]RoomMessage, 0, 256),
		CreatedAt: time.Now(),
		Active:    true,
	}

	agentRoomManager.Rooms[room.ID] = room
	agentRoomManager.Stats.RoomsCreated++

	// Place room in space topology
	PlaceNode(room.ID, "room",
		SpacePosition{Depth: 1, Layer: "conversation", Machine: CurrentPOV().Machine, Region: "eu"},
		CapOBSERVE)

	Log(LogInfo, "room.create", topic, "", "", 0, nil)
	Pulse("room.create", topic)

	return room
}

// JoinRoom adds an agent to a room.
func JoinRoom(roomID, agentID string) error {
	agentRoomManager.mu.Lock()
	defer agentRoomManager.mu.Unlock()

	room, ok := agentRoomManager.Rooms[roomID]
	if !ok { return fmt.Errorf("room %s not found", roomID[:8]) }

	// Fold: the agent becomes part of the room (light fold)
	room.Agents = append(room.Agents, agentID)
	agentRoomManager.Stats.AgentsJoined++

	// Space edge: agent → room
	ConnectNodes(agentID, roomID, "joins", 0)

	Log(LogInfo, "room.join", fmt.Sprintf("%s → %s", agentID[:8], roomID[:8]),
		"", "", 0, nil)

	return nil
}

// SendMessage sends a message in a room.
func SendRoomMessage(roomID, from, content string) error {
	agentRoomManager.mu.Lock()
	defer agentRoomManager.mu.Unlock()

	room, ok := agentRoomManager.Rooms[roomID]
	if !ok { return fmt.Errorf("room %s not found", roomID[:8]) }

	msg := RoomMessage{
		From:      from,
		Content:   content,
		Timestamp: time.Now(),
		Ref:       Ref([]byte(content)),
	}

	room.Messages = append(room.Messages, msg)
	if len(room.Messages) > 256 { room.Messages = room.Messages[1:] }
	agentRoomManager.Stats.MessagesSent++

	// If fold depth exists, track it
	if depth := FoldDepth(from); depth > agentRoomManager.Stats.FoldDepth {
		agentRoomManager.Stats.FoldDepth = depth
	}

	Pulse("room.message", fmt.Sprintf("%s: %s", from[:8], content[:min(40, len(content))]))

	return nil
}

// ── Room Status ──────────────────────────────────────────────────────

// AgentRoomStatus returns compact room status.
func AgentRoomStatus() string {
	agentRoomManager.mu.Lock()
	defer agentRoomManager.mu.Unlock()

	active := 0
	for _, r := range agentRoomManager.Rooms {
		if r.Active { active++ }
	}

	return fmt.Sprintf("rooms: %d total · %d active · %d messages · max fold: %d",
		len(agentRoomManager.Rooms), active,
		agentRoomManager.Stats.MessagesSent, agentRoomManager.Stats.FoldDepth)
}

// AgentRoomVakedFit returns the room's Vaked fit.
func AgentRoomVakedFit() string {
	return `AGENT ROOM = FOLD-LOWERED MULTI-AGENT SPACE

  Inspired by agentpipe (kevinelliott/agentpipe).
  Multiple AI agents conversing in a shared room.
  Folded into Vaked: room = space node, agents = folded participants.

  Originally envisioned pre-ultrawhale.
  Now native to vaked. Lowered. Aligned.

  "Multi-agent conversation, folded into the Vaked graph."`
}
