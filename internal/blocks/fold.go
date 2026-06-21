// LEGAL: ULTRA-RESEARCH-STATE. See LICENSE + docs/disclaimer.md.
package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Fold — Virtualized Subagent Runtime ───────────────────────────────
//
// Fold virtualizes a subagent's complete runtime into the parent.
// Tool calls execute in parent context. Output flows to parent.
// The subagent never existed as a separate entity.

// FoldContext is the virtualized execution context.
type FoldContext struct {
	mu          sync.Mutex
	ParentID    string
	AgentID     string
	Depth       int               // recursion depth (0 = parent)
	Tools       []string          // tools the folded agent used
	Output      []string          // output lines
	TokenCount  int64
	StartedAt   time.Time
	CompletedAt time.Time
	Status      string            // "folding", "completed", "error"
}

// FoldRegistry tracks all folded agents.
type FoldRegistry struct {
	mu       sync.Mutex
	contexts map[string]*FoldContext
}

var foldRegistry = &FoldRegistry{contexts: make(map[string]*FoldContext)}

// Fold executes a subagent inline in the parent's context.
func Fold(agentID, parentID, role string, depth int) (*FoldContext, error) {
	// Safety gate: max depth 5 (prevent infinite recursion)
	if depth > 5 { return nil, fmt.Errorf("fold: max recursion depth 5 exceeded") }
	// Safety gate: must be allowed
	if !IsAllowed() { return nil, fmt.Errorf("fold: permission denied") }
	foldRegistry.mu.Lock()
	defer foldRegistry.mu.Unlock()

	// Check for circular folding
	if _, exists := foldRegistry.contexts[agentID]; exists {
		return nil, fmt.Errorf("fold: agent %s already folded", agentID[:8])
	}

	ctx := &FoldContext{
		ParentID:  parentID,
		AgentID:   agentID,
		Depth:     depth,
		StartedAt: time.Now(),
		Status:    "folding",
	}

	foldRegistry.contexts[agentID] = ctx

	Log(LogInfo, "fold.start", fmt.Sprintf("%s (depth %d, role %s) → %s",
		agentID[:8], depth, role, parentID[:8]), "", "", 0, nil)

	return ctx, nil
}

// FoldToolCall records a tool call made by the folded agent.
func FoldToolCall(agentID, tool, result string) {
	foldRegistry.mu.Lock()
	defer foldRegistry.mu.Unlock()

	ctx, ok := foldRegistry.contexts[agentID]
	if !ok { return }

	ctx.Tools = append(ctx.Tools, tool)
	ctx.Output = append(ctx.Output, fmt.Sprintf("[%s] %s → %s", agentID[:8], tool, result[:min(80, len(result))]))
	ctx.TokenCount += int64(len(result))


}

// FoldUnwind completes the fold and returns output to parent.
func FoldUnwind(agentID string) []string {
	foldRegistry.mu.Lock()
	defer foldRegistry.mu.Unlock()

	ctx, ok := foldRegistry.contexts[agentID]
	if !ok { return nil }

	ctx.Status = "completed"
	ctx.CompletedAt = time.Now()

	Log(LogInfo, "fold.complete",
		fmt.Sprintf("%s → %s parent (%d tools, %d output lines, depth %d)",
			agentID[:8], ctx.ParentID[:8], len(ctx.Tools), len(ctx.Output), ctx.Depth),
		"", "", time.Since(ctx.StartedAt), nil)

	// Clean up — the context is consumed by parent
	delete(foldRegistry.contexts, agentID)

	return ctx.Output
}

// FoldDepth returns the current recursion depth for an agent.
func FoldDepth(agentID string) int {
	foldRegistry.mu.Lock()
	defer foldRegistry.mu.Unlock()
	if ctx, ok := foldRegistry.contexts[agentID]; ok {
		return ctx.Depth
	}
	return -1
}

// FoldStatus returns compact fold status.
func FoldStatus() string {
	foldRegistry.mu.Lock()
	defer foldRegistry.mu.Unlock()

	folding := 0
	for _, ctx := range foldRegistry.contexts {
		if ctx.Status == "folding" { folding++ }
	}

	return fmt.Sprintf("fold: %d agents (%d actively folding)",
		len(foldRegistry.contexts), folding)
}

// FoldVakedFit returns the fold primitive's Vaked fit.
func FoldVakedFit() string {
	return `FOLD = RECURSION THROUGH AGENTS

Full-Stop recurses through LAYERS (Declares → SACRED)
Fold recurses through AGENTS (parent → subagent → leaf)

parent >>= fold(subagent) >>= parent

The subagent's runtime virtualizes into the parent.
Tool calls execute in parent context.
Output flows up through the recursion tree.
Context wraps agent. Recursion continues.`
}

