// Package blocks — Agent is the ultrawhale subagent identity and context block.
// Every spawned subagent inherits brain context, memo access, and a POV.
// Agent blocks are scoped: only the owning subagent can read its own block.
package blocks

import (
	"fmt"
	"os"
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
	MaxCalls   int
	MaxIters   int
	Edge       *EdgeAgent // nil if not edge-deployed

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

	// Ralph: observe agent completion
	// Report to supervisor
	if status == "failed" || status == "timeout" {
		ReportAgentFailure(id)
	}

	ralph := GetRalph()
	ralph.Observe(fmt.Sprintf("agent:%s", a.Role),
		fmt.Sprintf("completed:%s", status),
		status,
		dur, tokens)
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

// DeployToEdge deploys this agent to Cloudflare edge.
// Requires CF_API_TOKEN. Returns the deployed EdgeAgent.
func (a *Agent) DeployToEdge() (*EdgeAgent, error) {
	cfg := DetectCF()
	if cfg == nil {
		return nil, fmt.Errorf("no CF credentials — run 'ultrawhale setup' or set CF_API_TOKEN")
	}
	
	edge := NewEdgeAgent(a.ID, a.Role)
	edge.Fiber.AppendLedger("prompt", fmt.Sprintf("agent spawned: %s (%s)", a.ID, a.Role))
	
	if err := edge.DeployEdgeAgent(cfg); err != nil {
		return nil, err
	}
	
	a.Edge = edge
	a.Status = "edge-active"
	return edge, nil
}

// IsEdgeDeployed returns true if this agent is running at the edge.
func (a *Agent) IsEdgeDeployed() bool {
	return a.Edge != nil && a.Edge.Status == "active"
}

// EdgeURL returns the edge deployment URL.
func (a *Agent) EdgeURL() string {
	if a.Edge == nil { return "" }
	return fmt.Sprintf("https://%s.%s.workers.dev", a.Edge.ID, os.Getenv("CF_ACCOUNT_ID"))
}
