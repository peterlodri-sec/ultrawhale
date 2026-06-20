package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Supervisor Primitive — OTP-like Agent Restart Tree ───────────────
// Vaked layer: Supervises. Watches agents, restarts failed ones.
// One-shot agents (subagents) vs persistent agents (swarms).

// Supervisor watches agents and restarts them on failure.
type Supervisor struct {
	mu       sync.Mutex
	children map[string]*SupervisedChild
	strategy RestartStrategy
	pov      POV
}

// SupervisedChild is an agent under supervision.
type SupervisedChild struct {
	ID          string
	Role        string
	Status      string // "running", "failed", "restarting", "stopped"
	Restarts    int
	MaxRestarts int
	LastFailure time.Time
	LastRestart time.Time
	Agent       *Agent // reference to the underlying agent
}

// RestartStrategy defines how the supervisor handles failures.
type RestartStrategy struct {
	MaxRestarts     int           // max restarts before giving up (default: 3)
	RestartDelay    time.Duration // delay before restart (default: 1s)
	BackoffFactor   float64       // multiply delay on each restart (default: 2.0)
	MaxRestartDelay time.Duration // cap on restart delay (default: 30s)
}

// DefaultRestartStrategy returns the default supervision policy.
func DefaultRestartStrategy() RestartStrategy {
	return RestartStrategy{
		MaxRestarts:     3,
		RestartDelay:    1 * time.Second,
		BackoffFactor:   2.0,
		MaxRestartDelay: 30 * time.Second,
	}
}

// NewSupervisor creates a new agent supervisor.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		children: make(map[string]*SupervisedChild),
		strategy: DefaultRestartStrategy(),
		pov:      CurrentPOV(),
	}
}

// Supervise adds an agent to the supervision tree.
func (s *Supervisor) Supervise(agent *Agent) *SupervisedChild {
	s.mu.Lock()
	defer s.mu.Unlock()

	child := &SupervisedChild{
		ID:          agent.ID,
		Role:        agent.Role,
		Status:      "running",
		MaxRestarts: s.strategy.MaxRestarts,
		Agent:       agent,
	}
	s.children[agent.ID] = child

	Log(LogInfo, "supervisor.add", fmt.Sprintf("%s (%s)", child.ID[:8], child.Role), "", "", 0, nil)
	return child
}

// ReportFailure notifies the supervisor that an agent has failed.
func (s *Supervisor) ReportFailure(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	child, ok := s.children[agentID]
	if !ok { return }

	child.Status = "failed"
	child.LastFailure = time.Now()

	// Auto-restart if under max
	if child.Restarts < child.MaxRestarts {
		delay := time.Duration(float64(s.strategy.RestartDelay) * powFloat(s.strategy.BackoffFactor, float64(child.Restarts)))
		if delay > s.strategy.MaxRestartDelay { delay = s.strategy.MaxRestartDelay }

		child.Restarts++
		child.Status = "restarting"
		child.LastRestart = time.Now()

		go func() {
			time.Sleep(delay)
			s.mu.Lock()
			if c, ok := s.children[agentID]; ok {
				c.Status = "running"
			}
			s.mu.Unlock()
			Log(LogInfo, "supervisor.restart", fmt.Sprintf("%s (%s) — attempt %d/%d after %s",
				child.ID[:8], child.Role, child.Restarts, child.MaxRestarts, delay.Round(time.Millisecond)),
				"", "", 0, nil)
		}()
	} else {
		child.Status = "stopped"
		Log(LogWarn, "supervisor.gaveup", fmt.Sprintf("%s (%s) — %d restarts, giving up",
			child.ID[:8], child.Role, child.Restarts),
			"", "", 0, nil)
	}
}

// SupervisorStatus returns compact supervision tree status.
func (s *Supervisor) SupervisorStatus() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	running := 0
	failed := 0
	for _, c := range s.children {
		if c.Status == "running" { running++ }
		if c.Status == "failed" || c.Status == "stopped" { failed++ }
	}

	return fmt.Sprintf("supervisor: %d children (%d running, %d failed/gave-up)",
		len(s.children), running, failed)
}

func powFloat(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ { result *= base }
	return result
}

// ── Global supervisor ──────────────────────────────────────────────────

var globalSupervisor = NewSupervisor()

// GetSupervisor returns the global supervisor.
func GetSupervisor() *Supervisor { return globalSupervisor }

// SuperviseAgent adds an agent to the global supervision tree.
func SuperviseAgent(agent *Agent) *SupervisedChild {
	return globalSupervisor.Supervise(agent)
}

// ReportAgentFailure notifies the global supervisor of a failure.
func ReportAgentFailure(agentID string) {
	globalSupervisor.ReportFailure(agentID)
}
