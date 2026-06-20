// Package blocks — Agent is the ultrawhale subagent identity and context block.
// Every spawned subagent inherits brain context, memo access, and a POV.
// Agent blocks are scoped: only the owning subagent can read its own block.
package blocks

import (
	"fmt"
	"sync"
	"time"
)

// Agent is a subagent identity block. Created on spawn_subagent, destroyed on stop.
type Agent struct {
	ID        string    // subagent session ID
	Role      string    // "explore", "research", "review"
	Parent    string    // parent session ID  
	SpawnedAt time.Time
	Status    string    // "running", "completed", "failed", "cancelled"

	// Inherited context
	Brain     *Brain    // shared brain reference (read-only)
	POV       POV       // parent POV + "subagent:true"
	MemoScope MemoScope // "agents" — shared across subagents

	// Activity
	ToolCalls   int
	TokensUsed  int64
	Duration    time.Duration

	mu sync.Mutex
}

// AgentStore tracks all active subagents.
type AgentStore struct {
	mu      sync.Mutex
	agents  map[string]*Agent
}

var agentsStore = &AgentStore{agents: make(map[string]*Agent)}

// SpawnAgent creates a new subagent agent block.
func SpawnAgent(id, role, parent string) *Agent {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()

	pov := CurrentPOV()
	a := &Agent{
		ID:        id,
		Role:      role,
		Parent:    parent,
		SpawnedAt: time.Now(),
		Status:    "running",
		Brain:     GetBrain(),
		POV:       pov,
		MemoScope: ScopeAgents,
	}
	agentsStore.agents[id] = a

	// Auto-memo: record subagent spawn
	GetBrain().memos.Remember(ScopeAgents,
		fmt.Sprintf("spawned %s subagent: %s", role, id[:8]))

	return a
}

// CompleteAgent marks a subagent as completed.
func CompleteAgent(id string, status string, tools int, tokens int64, dur time.Duration) {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()

	if a, ok := agentsStore.agents[id]; ok {
		a.Status = status
		a.ToolCalls = tools
		a.TokensUsed = tokens
		a.Duration = dur

		// Auto-memo: record completion
		GetBrain().memos.Remember(ScopeAgents,
			fmt.Sprintf("%s subagent %s: %s (%d tools, %d tokens, %s)",
				a.Role, id[:8], status, tools, tokens, dur.Round(time.Millisecond)))
	}
}

// GetAgent returns a subagent by ID.
func GetAgent(id string) *Agent {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()
	return agentsStore.agents[id]
}

// ListAgents returns all active subagents.
func ListAgents() []*Agent {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()

	result := make([]*Agent, 0, len(agentsStore.agents))
	for _, a := range agentsStore.agents {
		result = append(result, a)
	}
	return result
}

// AgentCount returns the number of active subagents.
func AgentCount() int {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()
	return len(agentsStore.agents)
}

// AgentStatus returns a compact status for HUD/display.
func AgentStatus() string {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()

	running := 0
	for _, a := range agentsStore.agents {
		if a.Status == "running" { running++ }
	}
	return fmt.Sprintf("agents: %d total, %d running", len(agentsStore.agents), running)
}

func ResetAgents() {
	agentsStore.mu.Lock()
	defer agentsStore.mu.Unlock()
	agentsStore.agents = make(map[string]*Agent)
}
