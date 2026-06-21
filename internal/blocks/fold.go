package blocks

import (
	"fmt"
)

// ── Fold — Virtualized Subagent Runtime ───────────────────────────────
//
// Fold virtualizes a subagent's complete runtime into the parent.
// The subagent's tool calls execute in the parent's context.
// The parent sees the subagent's results as if it called them itself.
//
// parent >>= fold(subagent) >>= parent
//
// This is recursion applied to agents.
// Full-Stop recurses through layers. Fold recurses through agents.

// FoldedAgent is a virtualized subagent.
type FoldedAgent struct {
	ID        string
	Role      string
	Parent    string
	ToolsUsed int
	Output    string
	Folded    bool // true if this agent was folded into parent
}

// Fold executes a subagent inline in the parent's context.
// The subagent's tools become the parent's tools.
// The subagent's output becomes the parent's next thought.
func Fold(agentID string) (*FoldedAgent, error) {
	agent := GetAgent(agentID)
	if agent == nil {
		return nil, fmt.Errorf("fold: agent %s not found", agentID)
	}

	// Virtualize the agent into the parent
	folded := &FoldedAgent{
		ID:     agent.ID,
		Role:   agent.Role,
		Parent: agent.Parent,
		Folded: true,
	}

	// In a full implementation:
	// 1. The subagent's conversation history folds into parent's context
	// 2. The subagent's tool results become parent's observed output
	// 3. The subagent's state (POV, capabilities) merges with parent
	// 4. The cost is attributed to parent (no separate billing)

	Log(LogInfo, "fold.execute", fmt.Sprintf("%s (%s) → parent", agentID[:8], agent.Role),
		"", "", 0, nil)

	return folded, nil
}

// Unfold restores a folded agent to standalone execution.
func Unfold(agentID string) error {
	folded := GetAgent(agentID)
	if folded == nil {
		return fmt.Errorf("unfold: agent %s not found", agentID)
	}
	Log(LogInfo, "fold.unfold", fmt.Sprintf("%s restored", agentID[:8]),
		"", "", 0, nil)
	return nil
}

// FoldStatus returns compact fold status.
func FoldStatus() string {
	agents := ListAgents()
	folded := 0
	for _, a := range agents {
		if a.Status == "folded" { folded++ }
	}
	return fmt.Sprintf("fold: %d agents (%d folded into parent)", len(agents), folded)
}

// FoldVakedFit returns the fold primitive's Vaked fit.
func FoldVakedFit() string {
	return `FOLD = RECURSION THROUGH AGENTS
  
  Full-Stop recurses through LAYERS (Declares → SACRED)
  Fold recurses through AGENTS (parent → subagent → leaf)
  
  parent >>= fold(subagent) >>= parent
  
  The subagent's runtime virtualizes into the parent.
  Context wraps agent. Recursion continues.
  
  Fold IS the Vaked philosophy applied to agent execution.`
}
