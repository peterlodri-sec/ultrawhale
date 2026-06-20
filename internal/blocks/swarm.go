// Package blocks — Swarm is a persistent worker with its own AgentField,
// DID identity, self-scoped memos, and child subagents.
// One level deep only. Swarms cannot spawn sub-swarms.
// Kept alive between tasks — not disposable like subagents.
package blocks

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Swarm is a persistent worker. Own identity, AgentField, memos, subagents.
type Swarm struct {
	mu sync.Mutex

	// Identity
	ID       string    // "swarm-1"
	DID      string    // did:key:swarm:{pubkey}
	Name     string    // descriptive name from first task
	Parent   string    // orchestrator universe ID
	POV      POV       // orchestrator POV + swarm:true

	// AgentField (nested)
	AFPorthttp int           // auto-assigned port (8686+)
	AFServer   *http.Server

	// Memos (self-scoped, visible to orchestrator after completion)
	Memos *MemoStore

	// Children (subagents only, NOT sub-swarms)
	Agents *AgentStore

	// State
	Status     string    // "idle", "active", "draining"
	CreatedAt  time.Time
	LastTaskAt time.Time
	TotalTasks int
	TotalTools int64

	// Budget (proportional to first task complexity)
	MaxCalls int
	MaxIters int

	active atomic.Bool
}

// SwarmStore tracks all active swarms.
type SwarmStore struct {
	mu      sync.Mutex
	swarms  map[string]*Swarm
	nextPort int // auto-assign starting at 8686
}

var swarmStore = &SwarmStore{
	swarms:   make(map[string]*Swarm),
	nextPort: 8686,
}

// SpawnSwarm creates a new persistent swarm. If one already exists and is idle, reuse it.
func SpawnSwarm(name, parent string, complexity int) *Swarm {
	swarmStore.mu.Lock()
	defer swarmStore.mu.Unlock()

	// Try to reuse an idle swarm
	for _, s := range swarmStore.swarms {
		if s.Status == "idle" {
			s.mu.Lock()
			s.Status = "active"
			s.LastTaskAt = time.Now()
			s.TotalTasks++
			s.active.Store(true)
			s.mu.Unlock()
			return s
		}
	}

	// Create new swarm
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubHex := hex.EncodeToString(pub)
	port := swarmStore.nextPort
	swarmStore.nextPort++

	home, _ := os.UserHomeDir()
	memoDir := filepath.Join(home, ".whale", "swarms", fmt.Sprintf("swarm-%d", port))
	os.MkdirAll(memoDir, 0o700)

	s := &Swarm{
		POV:      CurrentPOV(),
		ID:         fmt.Sprintf("swarm-%d", port),
		DID:        fmt.Sprintf("did:key:swarm:%s", pubHex[:40]),
		Name:       name,
		Parent:     parent,
		AFPorthttp: port,
		Memos:      &MemoStore{memos: make(map[string]Memo), dir: memoDir},
		Agents:     &AgentStore{agents: make(map[string]*Agent)},
		Status:     "active",
		CreatedAt:  time.Now(),
		LastTaskAt: time.Now(),
		TotalTasks: 1,
		MaxCalls:   proportionalBudget(complexity),
		MaxIters:   proportionalBudget(complexity) / 2,
	}
	s.active.Store(true)

	// Start nested AgentField
	s.startAgentField()

	swarmStore.swarms[s.ID] = s
	return s
}

// proportionalBudget returns budget based on task complexity.
func proportionalBudget(complexity int) int {
	base := 256
	switch {
	case complexity > 80:
		return base * 3 // 768
	case complexity > 50:
		return base * 2 // 512
	default:
		return base // 256
	}
}

func (s *Swarm) startAgentField() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"ok","swarm":"%s","did":"%s"}`, s.ID, s.DID[:40])
	})
	mux.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := s.Agents.List()
		fmt.Fprintf(w, `{"swarm":"%s","count":%d}`, s.ID, len(agents))
	})
	mux.HandleFunc("/api/v1/memos", func(w http.ResponseWriter, r *http.Request) {
		memos := s.Memos.RecallAll()
		fmt.Fprintf(w, `{"swarm":"%s","count":%d}`, s.ID, len(memos))
	})

	s.AFServer = &http.Server{Addr: fmt.Sprintf(":%d", s.AFPorthttp), Handler: mux}
	go s.AFServer.ListenAndServe()
}

// DrainSwarm marks a swarm as idle (keeps it alive for reuse).
func (s *Swarm) DrainSwarm() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "idle"
	s.active.Store(false)
}

// KillSwarm permanently destroys a swarm.
func (s *Swarm) KillSwarm() {
	swarmStore.mu.Lock()
	delete(swarmStore.swarms, s.ID)
	swarmStore.mu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.AFServer != nil {
		s.AFServer.Close()
	}
	s.Status = "dead"
	s.active.Store(false)
}

// SpawnSwarmAgent creates a subagent under this swarm.
func (s *Swarm) SpawnSwarmAgent(role, task, parent string) *Agent {
	agent := SpawnAgent(fmt.Sprintf("%s-%s-%d", s.ID, role, s.TotalTools), role, parent)
	s.TotalTools++
	return agent
}

// GetSwarm returns a swarm by ID.
func GetSwarm(id string) *Swarm {
	swarmStore.mu.Lock()
	defer swarmStore.mu.Unlock()
	return swarmStore.swarms[id]
}

// ListSwarms returns all swarms.
func ListSwarms() []*Swarm {
	swarmStore.mu.Lock()
	defer swarmStore.mu.Unlock()
	result := make([]*Swarm, 0, len(swarmStore.swarms))
	for _, s := range swarmStore.swarms {
		result = append(result, s)
	}
	return result
}

// SwarmCount returns the number of active swarms.
func SwarmCount() int {
	swarmStore.mu.Lock()
	defer swarmStore.mu.Unlock()
	return len(swarmStore.swarms)
}

// SwarmStatus returns compact status for sidepanel.
func SwarmStatus() string {
	swarmStore.mu.Lock()
	defer swarmStore.mu.Unlock()

	var parts []string
	for _, s := range swarmStore.swarms {
		parts = append(parts, fmt.Sprintf("%s:%s:%d", s.ID, s.Status, s.TotalTasks))
	}
	if len(parts) == 0 {
		return "swarms: 0"
	}
	return "swarms: " + fmt.Sprintf("%d", len(swarmStore.swarms)) + " (" + parts[0] + "...)"
}

// ComplexityScore estimates how complex a prompt is (0-100).
func ComplexityScore(prompt string) int {
	score := 0
	words := len(prompt) / 5 // rough word count
	if words > 50 { score += 30 } else if words > 20 { score += 15 }
	if containsAny(prompt, "build", "implement", "refactor", "migrate", "rewrite") { score += 40 }
	if containsAny(prompt, "module", "system", "architecture", "entire", "full") { score += 20 }
	if containsAny(prompt, "test", "fix", "review", "find", "search") { score += 10 }
	if score > 100 { score = 100 }
	return score
}

func containsAny(s string, words ...string) bool {
	for _, w := range words {
		if len(s) >= len(w) {
			for i := 0; i <= len(s)-len(w); i++ {
				if s[i:i+len(w)] == w { return true }
			}
		}
	}
	return false
}

// AgentStore.List returns all agents.
func (as *AgentStore) List() []*Agent {
	as.mu.Lock()
	defer as.mu.Unlock()
	result := make([]*Agent, 0, len(as.agents))
	for _, a := range as.agents {
		result = append(result, a)
	}
	return result
}
