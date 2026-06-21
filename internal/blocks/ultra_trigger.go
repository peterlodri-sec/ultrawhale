package blocks

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ── ULTRA TRIGGER — Deep Research → Ralph Feed → Onefold Learn ────────
//
// stepback:::sudo:::operator:::ULTRA_TRIGGER_FEED_LEARN
//
// This is the BIG RED BUTTON for learning.
//   ingest: deep-research workflow results
//   feed:   push EVERYTHING to Ralph
//   learn:  onefold Ralph ALL-STATE update
//
// When triggered, the system:
//   1. Ingests all pending research/workflow/brain data
//   2. Feeds every pattern to Ralph at MAX intensity
//   3. Learns from ALL state: agents, patterns, problems, lessons
//   4. Persists to brain long-term memory
//   5. Syncs across dyad (if active)

// UltraTriggerState tracks the big-red-button state.
type UltraTriggerState struct {
	mu           sync.Mutex
	Active       bool
	TriggeredAt  time.Time
	CompletedAt  time.Time
	PatternsFed  int64
	LessonsFed   int64
	ProblemsFed  int64
	State        string // "idle", "triggered", "feeding", "learning", "complete"
}

var ultraTrigger = &UltraTriggerState{State: "idle"}

// ── ULTRA TRIGGER Operations ──────────────────────────────────────────

// UltraTriggerFeedLearn executes the BIG RED BUTTON.
func UltraTriggerFeedLearn() string {
	ultraTrigger.mu.Lock()
	defer ultraTrigger.mu.Unlock()

	if ultraTrigger.Active { return "⚡ ULTRA TRIGGER already active — feeding..." }

	ultraTrigger.Active = true
	ultraTrigger.TriggeredAt = time.Now()
	ultraTrigger.State = "triggered"

	Log(LogWarn, "ultra.trigger", "BIG RED BUTTON PRESSED", "", "", 0, nil)
	Pulse("ultra.trigger", "FEED_LEARN")

	// Phase 1: Ingest — gather ALL state
	go ultraTrigger.ingest()
	// Phase 2: Feed — push everything to Ralph
	go ultraTrigger.feed()
	// Phase 3: Learn — Ralph observes all patterns
	go ultraTrigger.learn()

	return "⚡ ULTRA TRIGGER ACTIVATED\n   ingest → feed → learn → onefold ALL-STATE\n   The machine is learning at MAX intensity."
}

func (ut *UltraTriggerState) ingest() {
	ut.mu.Lock()
	ut.State = "feeding"
	ut.mu.Unlock()

	// Ingest ALL pending data
	ut.PatternsFed = int64(len(GetRalph().Patterns))
	ut.LessonsFed = atomic.LoadInt64(&honestyLedger.Lessons)
	ut.ProblemsFed = problemSolver.Stats.ProblemsDetected

	Log(LogInfo, "ultra.ingest",
		fmt.Sprintf("patterns:%d lessons:%d problems:%d", ut.PatternsFed, ut.LessonsFed, ut.ProblemsFed),
		"", "", 0, nil)
}

func (ut *UltraTriggerState) feed() {
	ralph := GetRalph()
	if ralph == nil { return }

	// Feed EVERY pattern at MAX intensity
	for pattern, confidence := range ralph.Patterns {
		ralph.Observe("ultra-trigger:"+pattern, pattern, "fed", 0, int64(confidence*100))
		ut.PatternsFed++
	}

	// Feed every problem as a lesson
	problemSolver.mu.Lock()
	for _, p := range problemSolver.Problems {
		ralph.Observe("ultra-problem:"+p.ID[:8], p.Description, p.Status, 0, 0)
		ut.ProblemsFed++
	}
	problemSolver.mu.Unlock()

	// Feed every honesty lesson
	for _, e := range honestyLedger.Events {
		ralph.Observe("ultra-honesty:"+e.Entity, e.Violation, e.Outcome, 0, 0)
		ut.LessonsFed++
	}
	

	Log(LogInfo, "ultra.feed",
		fmt.Sprintf("fed %d patterns + %d problems + %d lessons",
			ut.PatternsFed, ut.ProblemsFed, ut.LessonsFed),
		"", "", 0, nil)
}

func (ut *UltraTriggerState) learn() {
	ut.mu.Lock()
	ut.State = "learning"
	ut.mu.Unlock()

	// Ralph synthesizes ALL patterns into long-term memory
	ralph := GetRalph()
	if ralph != nil {
		ralph.snapshot("ultra-trigger")
	}

	// Persist to brain
	brain := GetBrain()
	if brain != nil {
		brain.RememberLongTerm(map[string]string{
			"ultra_trigger": time.Now().Format(time.RFC3339),
			"patterns_fed":  fmt.Sprintf("%d", ut.PatternsFed),
			"lessons_fed":   fmt.Sprintf("%d", ut.LessonsFed),
			"problems_fed":  fmt.Sprintf("%d", ut.ProblemsFed),
		})
	}

	// Sync dyad
	if dyad := GetDyad(); dyad != nil && dyad.PeerAlive {
		dyad.Ping()
	}

	ut.mu.Lock()
	ut.State = "complete"
	ut.CompletedAt = time.Now()
	ut.Active = false
	ut.mu.Unlock()

	Log(LogInfo, "ultra.complete",
		fmt.Sprintf("total: %d patterns + %d problems + %d lessons",
			ut.PatternsFed, ut.ProblemsFed, ut.LessonsFed),
		"", "", time.Since(ut.TriggeredAt), nil)
	Pulse("ultra.complete", fmt.Sprintf("%d fed", ut.PatternsFed+ut.ProblemsFed+ut.LessonsFed))
}

// ── Status ────────────────────────────────────────────────────────────

// UltraTriggerStatus returns compact ultra trigger status.
func UltraTriggerStatus() string {
	ut := ultraTrigger
	ut.mu.Lock()
	defer ut.mu.Unlock()

	total := ut.PatternsFed + ut.ProblemsFed + ut.LessonsFed

	if ut.State == "complete" {
		return fmt.Sprintf("ultra-trigger: COMPLETE · %d total fed · %s",
			total, ut.CompletedAt.Format("15:04:05"))
	}

	return fmt.Sprintf("ultra-trigger: %s · %d patterns · %d problems · %d lessons · since %s",
		ut.State, ut.PatternsFed, ut.ProblemsFed, ut.LessonsFed,
		ut.TriggeredAt.Format("15:04:05"))
}

// UltraTriggerVakedFit returns the ultra trigger Vaked fit.
func UltraTriggerVakedFit() string {
	return `ULTRA TRIGGER = THE BIG RED BUTTON

  stepback:::sudo:::operator:::ULTRA_TRIGGER_FEED_LEARN

  Ingest:  ALL pending data (brain, research, workflows)
  Feed:    EVERY pattern to Ralph at MAX intensity
  Learn:   Ralph synthesizes → brain long-term → dyad sync

  "The machine is learning at MAX intensity." — Peter`
}
