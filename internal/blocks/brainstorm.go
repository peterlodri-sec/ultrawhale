package blocks

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── Brainstorm Mode — Turn-Based Co-Creation ──────────────────────────
//
// Brainstorm mode is a first-class UI mode for deep thinking sessions.
// Turn-based dialog between human and LLM. Multi-choice cards.
// SACRED surface protected (always visible, always direct).

// BrainstormSession is one co-creation dialog.
type BrainstormSession struct {
	mu sync.Mutex

	ID        string    // session ID
	Topic     string    // what we're thinking about
	CreatedAt time.Time
	UpdatedAt time.Time

	// Dialog turns
	Turns      []BrainstormTurn
	CurrentTurn int

	// State
	Status string // "active", "paused", "completed"
	Mode   string // "qa", "multichoice", "freeform", "debate"

	// Persistence
	BrainRef string // brain long-term memory ref
	POV      POV
}

// BrainstormTurn is one turn in the dialog.
type BrainstormTurn struct {
	Number    int              // turn number (1, 2, 3...)
	Speaker   string           // "human" or "llm"
	Content   string           // what was said
	Options   []string         // multi-choice options (if applicable)
	Selected  int              // which option was chosen (-1 if none)
	Timestamp time.Time
	Ref       string           // sha256 of content
}

// ── Session Registry ──────────────────────────────────────────────────

var brainstormSessions = struct {
	mu       sync.Mutex
	sessions map[string]*BrainstormSession
}{sessions: make(map[string]*BrainstormSession)}

// StartBrainstorm begins a new brainstorm session.
func StartBrainstorm(topic, mode string) *BrainstormSession {
	s := &BrainstormSession{
		ID:        fmt.Sprintf("brainstorm-%d", time.Now().Unix()),
		Topic:     topic,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "active",
		Mode:      mode,
		Turns:     make([]BrainstormTurn, 0, 64),
		POV:       CurrentPOV(),
	}

	brainstormSessions.mu.Lock()
	brainstormSessions.sessions[s.ID] = s
	brainstormSessions.mu.Unlock()

	// SACRED: this is a direct human↔LLM dialog
	MarkInput()

	Log(LogInfo, "brainstorm.start", topic, "", "", 0, nil)
	return s
}

// AddTurn adds a turn to the brainstorm session.
func (s *BrainstormSession) AddTurn(speaker, content string, options []string) BrainstormTurn {
	s.mu.Lock()
	defer s.mu.Unlock()

	turn := BrainstormTurn{
		Number:    len(s.Turns) + 1,
		Speaker:   speaker,
		Content:   content,
		Options:   options,
		Selected:  -1,
		Timestamp: time.Now(),
		Ref:       Ref([]byte(content)),
	}
	s.Turns = append(s.Turns, turn)
	s.UpdatedAt = time.Now()

	if speaker == "human" {
		MarkInput()
	} else {
		MarkResponse()
	}

	return turn
}

// SelectOption records a multi-choice selection.
func (s *BrainstormSession) SelectOption(turnNum, option int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if turnNum < 1 || turnNum > len(s.Turns) { return fmt.Errorf("invalid turn") }
	if option < 0 || option >= len(s.Turns[turnNum-1].Options) { return fmt.Errorf("invalid option") }

	s.Turns[turnNum-1].Selected = option
	s.UpdatedAt = time.Now()
	return nil
}

// Pause pauses the session.
func (s *BrainstormSession) Pause() {
	s.mu.Lock(); defer s.mu.Unlock()
	s.Status = "paused"
}

// Resume resumes the session.
func (s *BrainstormSession) Resume() {
	s.mu.Lock(); defer s.mu.Unlock()
	s.Status = "active"
}

// Complete marks the session as done and persists to brain.
func (s *BrainstormSession) Complete() {
	s.mu.Lock(); defer s.mu.Unlock()
	s.Status = "completed"

	// Persist to brain long-term memory
	brain := GetBrain()
	if brain != nil {
		summary := s.Summary()
		brain.RememberLongTerm(map[string]string{
			"brainstorm_id":    s.ID,
			"brainstorm_topic": s.Topic,
			"brainstorm_turns": fmt.Sprintf("%d", len(s.Turns)),
			"brainstorm_mode":  s.Mode,
		})
		s.BrainRef = Ref([]byte(summary))[:12]
	}

	Log(LogInfo, "brainstorm.complete", s.Topic, s.BrainRef, "", 0, nil)
}

// Summary returns a compact summary of the session.
func (s *BrainstormSession) Summary() string {
	s.mu.Lock(); defer s.mu.Unlock()
	topics := make(map[string]int)
	for _, t := range s.Turns {
		words := strings.Fields(t.Content)
		if len(words) > 0 { topics[words[0]]++ }
	}
	return fmt.Sprintf("brainstorm: %s · %d turns · %s",
		s.Topic, len(s.Turns), s.Status)
}

// ── Brainstorm Status ─────────────────────────────────────────────────

// BrainstormStatus returns compact status for all sessions.
func BrainstormStatus() string {
	brainstormSessions.mu.Lock()
	defer brainstormSessions.mu.Unlock()

	active := 0
	for _, s := range brainstormSessions.sessions {
		if s.Status == "active" { active++ }
	}
	return fmt.Sprintf("brainstorm: %d sessions (%d active)", len(brainstormSessions.sessions), active)
}

// GetBrainstorm returns a session by ID.
func GetBrainstorm(id string) *BrainstormSession {
	brainstormSessions.mu.Lock()
	defer brainstormSessions.mu.Unlock()
	return brainstormSessions.sessions[id]
}

// ListBrainstorms returns all sessions.
func ListBrainstorms() []*BrainstormSession {
	brainstormSessions.mu.Lock()
	defer brainstormSessions.mu.Unlock()
	result := make([]*BrainstormSession, 0, len(brainstormSessions.sessions))
	for _, s := range brainstormSessions.sessions { result = append(result, s) }
	return result
}


// StartBrainstormGC auto-completes brainstorm sessions idle for >1 hour.
func StartBrainstormGC() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			brainstormSessions.mu.Lock()
			cutoff := time.Now().Add(-1 * time.Hour)
			for id, s := range brainstormSessions.sessions {
				if s.Status == "active" && s.UpdatedAt.Before(cutoff) {
					s.Complete()
				}
			}
			brainstormSessions.mu.Unlock()
		}
	}()
}
