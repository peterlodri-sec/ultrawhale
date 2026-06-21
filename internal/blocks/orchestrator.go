// Package blocks — Orchestrator is the single TUI universe controller.
// There is exactly ONE orchestrator per TUI session.
// It never calls the LLM directly — every prompt is delegated to a subagent.
// Identity: own DID/Ed25519 keypair, separate from AgentField.
package blocks

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Orchestrator is the TUI universe controller.
type Orchestrator struct {
	mu sync.Mutex

	// Identity
	ID       string    // "ultrawhale-orchestrator"
	Universe string    // session ID — one per TUI session
	DID      string    // own Ed25519 DID (separate from AgentField)
	PubKey   string    // hex public key
	POV      POV       // orchestrator POV

	// Fleet
	Agents    *AgentStore
	Brain     *Brain
	agentsMD  string    // loaded from AGENTS.md

	// Session
	StartedAt  time.Time
	TotalTurns int
	TotalTools int

	// Agent definitions parsed from AGENTS.md
	Definitions []AgentDef
}

// AgentDef is a parsed agent definition from AGENTS.md.
type AgentDef struct {
	Name       string   // "swe", "explore", "review"
	Role       string   // "Software engineer"
	Model      string   // "deepseek-v4-flash"
	Tools      []string // ["shell.run", "workspace.read"]
	MaxCalls   int      // 256, 128, 64
	MaxIters   int      // 128, 64, 32
}

var orchestrator *Orchestrator

// InitOrchestrator creates the single TUI universe orchestrator.
func InitOrchestrator(sessionID string) *Orchestrator {
	if orchestrator != nil {
		return orchestrator
	}

	// Generate orchestrator DID (separate from AgentField)
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubHex := hex.EncodeToString(pub)

	home, _ := os.UserHomeDir()
	keyDir := filepath.Join(home, ".whale", "orchestrator")
	os.MkdirAll(keyDir, 0o700)

	o := &Orchestrator{
		ID:       "ultrawhale-orchestrator",
		Universe: sessionID,
		DID:      fmt.Sprintf("did:key:orch:%s", pubHex[:40]),
		PubKey:   pubHex,
		POV:      CurrentPOV(),
		Agents:   agentsStore,
		Brain:    GetBrain(),
		StartedAt: time.Now(),
		Definitions: DefaultAgentDefs(),
	}

	// Load AGENTS.md if exists
	if data, err := os.ReadFile("AGENTS.md"); err == nil {
		o.agentsMD = string(data)
		o.Definitions = parseAgentsMD(o.agentsMD)
	}

	orchestrator = o

	// Initialize dyad if running on known machines
	if o.POV.Machine == "M1" || o.POV.Machine == "dev-cx53" {
		peer := "dev-cx53"
		if o.POV.Machine == "dev-cx53" { peer = "M1" }
		InitDyad(peer, "amd64", "go")
	}

	// Load previously learned Ralph patterns from brain
	ralph := GetRalph()
	loaded := ralph.LoadFromBrain()
	if loaded > 0 {
		ralph.snapshot(fmt.Sprintf("loaded %d patterns from brain", loaded))
	}

	// Auto-memo: orchestrator started
	o.Brain.memos.Remember(ScopeInternal,
		fmt.Sprintf("orchestrator started: universe=%s, did=%s", sessionID, o.DID[:45]))

	return o
}

// GetOrchestrator returns the singleton orchestrator.
func GetOrchestrator() *Orchestrator {
	if orchestrator == nil {
		return InitOrchestrator("default")
	}
	return orchestrator
}

// DelegatePrompt spawns a subagent for the given prompt.
// The orchestrator NEVER calls the LLM directly — always delegates.
func (o *Orchestrator) DelegatePrompt(prompt string) (string, string) {
	// Classify the prompt to pick the right agent
	def := o.classifyByCapability(prompt)
	if def.Name == "" {
		// Fallback: default to explore agent
		def = AgentDef{Name: "explore", Model: "deepseek-v4-flash", MaxCalls: 128, MaxIters: 64}
	}

	// Ralph: observe delegation
	ralph := GetRalph()
	_ = ralph

	// Spawn subagent
	agent := SpawnAgent(
		fmt.Sprintf("%s-%d", def.Name, o.TotalTurns),
		def.Name,
		o.Universe,
	)
	agent.MaxCalls = def.MaxCalls
	agent.MaxIters = def.MaxIters

	o.TotalTurns++

	// Dyad heartbeat
	if dyad := GetDyad(); dyad != nil {
		dyad.Ping()
	}

	// Ralph: observe this delegation
	ralph.Observe(prompt, def.Name, "delegated", 0, 0)

	return agent.ID, def.Name
}

// DelegateSwarm spawns a swarm for complex tasks.
func (o *Orchestrator) DelegateSwarm(prompt string) *Swarm {
	score := ComplexityScore(prompt)
	if score >= 40 {
		return SpawnSwarm(prompt[:min(40, len(prompt))], o.Universe, score)
	}
	return nil
}

func min(a, b int) int { if a < b { return a }; return b }

// classifyPrompt picks the best agent definition for a prompt.
func (o *Orchestrator) classifyPrompt(prompt string) AgentDef {
	lower := strings.ToLower(prompt)

	// Review triggers
	if strings.Contains(lower, "review") || strings.Contains(lower, "pr #") ||
		strings.Contains(lower, "audit") || strings.Contains(lower, "check") {
		for _, d := range o.Definitions {
			if d.Name == "review" { return d }
		}
	}

	// Exploration triggers
	if strings.Contains(lower, "find") || strings.Contains(lower, "search") ||
		strings.Contains(lower, "explore") || strings.Contains(lower, "what") ||
		strings.Contains(lower, "where") || strings.Contains(lower, "show") ||
		strings.Contains(lower, "how") {
		for _, d := range o.Definitions {
			if d.Name == "explore" { return d }
		}
	}

	// Ultracode trigger
	if strings.HasPrefix(strings.ToLower(prompt), "/ultracode") {
		for _, d := range o.Definitions {
			if d.Name == "swe" { return d }
		}
	}

	// Workflow trigger — delegate to swe agent
	// Workflow scripts are at .whale/workflows/*.js
	// Agent can spawn them via /run command
	if strings.HasPrefix(strings.ToLower(prompt), "/workflow") || strings.Contains(strings.ToLower(prompt), "workflow") {
		for _, d := range o.Definitions {
			if d.Name == "swe" { return d }
		}
	}

	// Default: swe
	for _, d := range o.Definitions {
		if d.Name == "swe" { return d }
	}

	// Fallback
	return AgentDef{Name: "explore", Model: "deepseek-v4-flash", MaxCalls: 128, MaxIters: 64}
}

// OrchestratorStatus returns a compact status for HUD/sidepanel.
func (o *Orchestrator) OrchestratorStatus() string {
	o.mu.Lock()
	defer o.mu.Unlock()

	return fmt.Sprintf("orch: %s | %d turns | %d agents | %d swarms | brain: %d memos",
		o.DID[:20],
		o.TotalTurns,
		AgentCount(),
		SwarmCount(),
		o.Brain.memos.Count(),
	)
}

// DefaultAgentDefs returns the built-in agent definitions.
func DefaultAgentDefs() []AgentDef {
	return []AgentDef{
		{Name: "swe", Role: "Software engineer", Model: "deepseek-v4-flash",
			Tools: []string{"shell.run", "workspace.read", "workspace.write"},
			MaxCalls: 256, MaxIters: 128},
		{Name: "explore", Role: "Codebase explorer", Model: "deepseek-v4-flash",
			Tools: []string{"shell.run", "workspace.read"},
			MaxCalls: 128, MaxIters: 64},
		{Name: "review", Role: "Code reviewer", Model: "deepseek-v4-flash",
			Tools: []string{"shell.run", "workspace.read"},
			MaxCalls: 64, MaxIters: 32},
	}
}

// parseAgentsMD parses AGENTS.md into AgentDefs.
func parseAgentsMD(content string) []AgentDef {
	// Simple parser: ## name sections
	defs := DefaultAgentDefs()
	// AGENTS.md overrides defaults — for now, use defaults
	_ = content
	return defs
}


func (o *Orchestrator) classifyByCapability(prompt string) AgentDef {
	lower := strings.ToLower(prompt)
	
	// Capability-based routing: what does this task NEED?
	needsWrite := strings.Contains(lower, "implement") || strings.Contains(lower, "build") || strings.Contains(lower, "write") || strings.Contains(lower, "fix") || strings.Contains(lower, "refactor")
	needsExecute := strings.Contains(lower, "run") || strings.Contains(lower, "test") || strings.Contains(lower, "deploy") || strings.Contains(lower, "bench")
	
	if needsWrite || needsExecute {
		// Needs FULL capabilities
		for _, d := range o.Definitions {
			if GetCapProfile(d.Name).Can(CapWrite) { return d }
		}
	}
	
	// Default: OBSERVE (explore, review)
	for _, d := range o.Definitions {
		if d.Name == "explore" { return d }
	}
	return o.Definitions[0]
}
