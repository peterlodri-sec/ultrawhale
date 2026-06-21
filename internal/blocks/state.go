package blocks

import (
	"fmt"
	"sync"
)

// ── State Primitive — Context State Machine ───────────────────────────
// Vaked dimension: Context. State machines for block lifecycle.

// StateMachine tracks state transitions for any block.
type StateMachine struct {
	mu          sync.Mutex
	Current     string
	Transitions map[string][]string // state → allowed next states
	History     []StateTransition
	MaxHistory  int
}

// StateTransition is one state change.
type StateTransition struct {
	From   string
	To     string
	Reason string
	Lamport int64
}

// NewStateMachine creates a state machine.
func NewStateMachine(initial string, transitions map[string][]string) *StateMachine {
	return &StateMachine{
		Current:     initial,
		Transitions: transitions,
		History:     make([]StateTransition, 0, 64),
		MaxHistory:  64,
	}
}

// Transition attempts to change state.
func (sm *StateMachine) Transition(to, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	allowed, ok := sm.Transitions[sm.Current]
	if !ok { return fmt.Errorf("state: no transitions from %s", sm.Current) }

	for _, a := range allowed {
		if a == to {
			sm.History = append(sm.History, StateTransition{
				From: sm.Current, To: to, Reason: reason, Lamport: TickLamport(),
			})
			if len(sm.History) > sm.MaxHistory { sm.History = sm.History[1:] }
			sm.Current = to
			Log(LogInfo, "state.transition", fmt.Sprintf("%s → %s: %s", sm.Current, to, reason), "", "", 0, nil)
			return nil
		}
	}
	return fmt.Errorf("state: %s → %s not allowed (from %s)", sm.Current, to, sm.Current)
}

// DefaultAgentStates returns the standard agent state machine.
func DefaultAgentStates() *StateMachine {
	return NewStateMachine("idle", map[string][]string{
		"idle":      {"running"},
		"running":   {"completed", "failed", "paused"},
		"paused":    {"running", "failed"},
		"completed": {"idle"},
		"failed":    {"idle", "running"},
	})
}

// DefaultWorkflowStates returns the standard workflow state machine.
func DefaultWorkflowStates() *StateMachine {
	return NewStateMachine("draft", map[string][]string{
		"draft":    {"active"},
		"active":   {"completed", "failed", "paused"},
		"paused":   {"active"},
		"completed": {"draft"},
		"failed":    {"draft", "active"},
	})
}
