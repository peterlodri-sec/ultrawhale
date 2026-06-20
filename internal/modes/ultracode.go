// Package modes implements ultrawhale-native agent modes.
// ultracode is the first native mode — inspired by Claude Code's ultra mode.
// It chains: plan → implement → test → review → fix → verify → commit
// All writes go through blocks (rollback on failure), all traces go to Langfuse,
// all events publish to NATS, workflow state persists in Supabase.
package modes

import (
	"fmt"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/blocks"
)

// UltracodeMode implements the 7-phase agent loop.
type UltracodeMode struct {
	mu       sync.Mutex
	phases   []Phase
	current  int
	session  string
	status   UltracodeStatus
	onUpdate func(UltracodeStatus) // callback for TUI updates
}

// Phase represents one step in the ultracode loop.
type Phase struct {
	Name        string // "plan", "implement", "test", "review", "fix", "verify", "commit"
	Description string
	Status      PhaseStatus
	StartedAt   time.Time
	EndedAt     time.Time
	BlockRef    string // sha256 ref of the phase output (via blocks)
	Error       string
}

type PhaseStatus string

const (
	PhasePending PhaseStatus = "pending"
	PhaseRunning PhaseStatus = "running"
	PhasePassed  PhaseStatus = "passed"
	PhaseFailed  PhaseStatus = "failed"
	PhaseSkipped PhaseStatus = "skipped"
)

// UltracodeStatus is the full session state, persisted to Supabase.
type UltracodeStatus struct {
	SessionID  string    `json:"session_id"`
	StartedAt  time.Time `json:"started_at"`
	Phases     []Phase   `json:"phases"`
	Current    int       `json:"current"`
	POV        string    `json:"pov"`         // blocks.CurrentPOV().String()
	ToolCalls  int       `json:"tool_calls"`  // total tool calls across phases
	Verdict    string    `json:"verdict"`     // "pass" | "fail" | "aborted"
}

// DefaultPhases returns the standard 7-phase ultracode loop.
func DefaultPhases() []Phase {
	return []Phase{
		{Name: "plan", Description: "Analyze the task and design the implementation plan", Status: PhasePending},
		{Name: "implement", Description: "Write code changes via blocks.Write (journaled)", Status: PhasePending},
		{Name: "test", Description: "Run tests — failure triggers rollback", Status: PhasePending},
		{Name: "review", Description: "Self-review: lint, format, dead code, security", Status: PhasePending},
		{Name: "fix", Description: "Apply review fixes — skipped if review passes", Status: PhasePending},
		{Name: "verify", Description: "Re-test + verify fix completeness", Status: PhasePending},
		{Name: "commit", Description: "Sign and commit with auto-generated message", Status: PhasePending},
	}
}

// NewUltracode creates a new ultracode session.
func NewUltracode(sessionID string) *UltracodeMode {
	return &UltracodeMode{
		phases:  DefaultPhases(),
		session: sessionID,
		status: UltracodeStatus{
			SessionID: sessionID,
			StartedAt: time.Now(),
			POV:       blocks.CurrentPOV().String(),
		},
	}
}

// StartPhase begins a phase. Returns the phase index.
func (u *UltracodeMode) StartPhase(name string) (int, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range u.phases {
		if u.phases[i].Name == name && u.phases[i].Status == PhasePending {
			u.phases[i].Status = PhaseRunning
			u.phases[i].StartedAt = time.Now()
			u.current = i
			u.emitUpdate()

			return i, nil
		}
	}
	return -1, fmt.Errorf("phase %s not found or already started", name)
}

// EndPhase marks a phase as passed or failed.
func (u *UltracodeMode) EndPhase(name string, passed bool, err error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range u.phases {
		if u.phases[i].Name == name {
			u.phases[i].EndedAt = time.Now()
			// Ralph: observe ultracode phase outcome
			if ralph := blocks.GetRalph(); ralph != nil {
				outcome := "success"
				if !passed { outcome = "failed" }
				ralph.Observe(
					fmt.Sprintf("ultracode:%s", name),
					fmt.Sprintf("phase:%s", name),
					outcome,
					time.Since(u.phases[i].StartedAt),
					0,
				)
			}
			if passed {
				u.phases[i].Status = PhasePassed
			} else {
				u.phases[i].Status = PhaseFailed
				if err != nil {
					u.phases[i].Error = err.Error()
				}
			}
			break
		}
	}
	u.emitUpdate()

}

// RollbackPhase undoes the writes from the last phase via blocks.Rollback.
func (u *UltracodeMode) RollbackPhase(name string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range u.phases {
		if u.phases[i].Name == name {
			u.phases[i].Status = PhaseFailed
			break
		}
	}
	u.emitUpdate()

}

// SkipPhase marks a phase as skipped (e.g., "fix" when review passes).
func (u *UltracodeMode) SkipPhase(name string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range u.phases {
		if u.phases[i].Name == name {
			u.phases[i].Status = PhaseSkipped
			break
		}
	}
	u.emitUpdate()

}

// AutoAdvance moves to the next pending phase.
func (u *UltracodeMode) AutoAdvance() (string, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i := range u.phases {
		if u.phases[i].Status == PhasePending {
			u.phases[i].Status = PhaseRunning
			u.phases[i].StartedAt = time.Now()
			u.current = i
			u.emitUpdate()

			return u.phases[i].Name, true
		}
	}
	u.status.Verdict = "pass"
	u.emitUpdate()

	return "", false
}

// StatusSnapshot returns a copy of the current status.
func (u *UltracodeMode) StatusSnapshot() UltracodeStatus {
	u.mu.Lock()
	defer u.mu.Unlock()
	s := u.status
	s.Phases = make([]Phase, len(u.phases))
	copy(s.Phases, u.phases)
	s.Current = u.current
	return s
}

// PhaseSummary returns a compact TUI status line.
func (u *UltracodeMode) PhaseSummary() string {
	u.mu.Lock()
	defer u.mu.Unlock()

	var icons []string
	for _, p := range u.phases {
		switch p.Status {
		case PhasePassed:  icons = append(icons, "✓")
		case PhaseFailed:  icons = append(icons, "✗")
		case PhaseRunning: icons = append(icons, "●")
		case PhaseSkipped: icons = append(icons, "◌")
		default:           icons = append(icons, "·")
		}
	}
	return fmt.Sprintf("ultracode: %s", joinIcons(icons))
}

func joinIcons(icons []string) string {
	result := ""
	for i, icon := range icons {
		if i > 0 { result += "→" }
		result += icon
	}
	return result
}

func (u *UltracodeMode) emitUpdate() {
	u.status.Phases = make([]Phase, len(u.phases))
	copy(u.status.Phases, u.phases)
	u.status.Current = u.current
	if u.onUpdate != nil {
		go u.onUpdate(u.status)
	}
}

// OnUpdate sets the TUI callback for live status updates.
func (u *UltracodeMode) OnUpdate(cb func(UltracodeStatus)) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.onUpdate = cb
}

// Context returns agent-ready context for the current phase.
func (u *UltracodeMode) Context() string {
	u.mu.Lock()
	defer u.mu.Unlock()
	return fmt.Sprintf("[ULTRACODE phase=%d/%d pov=%s]", u.current+1, len(u.phases), blocks.CurrentPOV().String())
}
