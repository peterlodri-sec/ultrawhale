package blocks

import (
	"fmt"
	"sync"
)

// ── Agent Mesh Discovery ──────────────────────────────────────────────
// Agents announce their capabilities on the mesh.
// Other agents discover peers via CapRegistry.

// AgentAnnouncement is published to whale.mesh.agent.{id}
type AgentAnnouncement struct {
	ID           string      `json:"id"`
	Role         string      `json:"role"`
	Capabilities CapProfile  `json:"capabilities"`
	Status       string      `json:"status"`
	POV          POV         `json:"pov"`
}

var meshAnnouncements = struct {
	mu    sync.Mutex
	peers map[string]AgentAnnouncement
}{peers: make(map[string]AgentAnnouncement)}

// AnnounceAgent publishes an agent's capabilities to the mesh.
func AnnounceAgent(agent *Agent) {
	meshAnnouncements.mu.Lock()
	defer meshAnnouncements.mu.Unlock()

	meshAnnouncements.peers[agent.ID] = AgentAnnouncement{
		ID:           agent.ID,
		Role:         agent.Role,
		Capabilities: GetCapProfile(agent.Role),
		Status:       agent.Status,
		POV:          agent.POV,
	}

	Log(LogInfo, "mesh.agent.announce",
		fmt.Sprintf("%s (%s) — %s", agent.ID[:8], agent.Role, GetCapProfile(agent.Role).Name),
		"", "", 0, nil)
}

// DiscoverAgents returns agents matching a capability requirement.
func DiscoverAgents(need Capability) []AgentAnnouncement {
	meshAnnouncements.mu.Lock()
	defer meshAnnouncements.mu.Unlock()

	var result []AgentAnnouncement
	for _, a := range meshAnnouncements.peers {
		if a.Capabilities.Can(need) {
			result = append(result, a)
		}
	}
	return result
}

// MeshAgentStatus returns compact mesh agent status.
func MeshAgentStatus() string {
	meshAnnouncements.mu.Lock()
	defer meshAnnouncements.mu.Unlock()
	return fmt.Sprintf("mesh: %d agents announced", len(meshAnnouncements.peers))
}
