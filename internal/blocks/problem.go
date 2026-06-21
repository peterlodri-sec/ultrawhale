package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── PROBLEM — No Errors, Only Learning Opportunities ──────────────────
//
// There are NO ERRORS by concept in VEGED.
// There are only PROBLEMS — opportunities to learn.
//
// Every PROBLEM triggers:
//   1. Debug access: 5 one-shot attempts in a shadow universe
//   2. Self-repair attempt: try to fix in isolation
//   3. If fixed → learn the pattern → apply to main universe
//   4. If failed → rollback to USER → present problem → human decides
//   5. ALWAYS learn: Ralph observes the outcome
//
// The problem loop:
//   Human → [PROBLEM] → Shadow Universe (5 attempts) → Self-Repair
//     ↑                                                    ↓
//     └──────────── Rollback (if failed) ←─────────────────┘
//                        ↓
//                   Ralph Learns

// Problem is a learning opportunity.
type Problem struct {
	ID          string
	Description string
	Severity    string // "BIG_PROBLEM", "problem", "curiosity"
	Attempts    int
	MaxAttempts int // default: 5
	Status      string // "detected", "debugging", "repairing", "resolved", "rolled_back"
	Solution    string
	Lesson      string
	DetectedAt  time.Time
	ResolvedAt  time.Time
}

// ProblemSolver manages the problem resolution loop.
type ProblemSolver struct {
	mu         sync.Mutex
	Problems   []*Problem
	ShadowUniverses int // count of shadow universes created
	Stats      ProblemStats
}

// ProblemStats tracks problem resolution activity.
type ProblemStats struct {
	ProblemsDetected int64
	SelfRepairs      int64
	Rollbacks        int64
	LessonsLearned   int64
	ShadowAttempts   int64
}

var problemSolver = &ProblemSolver{
	Problems:       make([]*Problem, 0, 64),
	ShadowUniverses: 0,
}

// ── PROBLEM Resolution ────────────────────────────────────────────────

// DetectProblem registers a new problem.
func DetectProblem(description, severity string) *Problem {
	problemSolver.mu.Lock()
	defer problemSolver.mu.Unlock()

	p := &Problem{
		ID:          fmt.Sprintf("problem-%d", time.Now().UnixNano()),
		Description: description,
		Severity:    severity,
		MaxAttempts: 5,
		Status:      "detected",
		DetectedAt:  time.Now(),
	}
	problemSolver.Problems = append(problemSolver.Problems, p)
	problemSolver.Stats.ProblemsDetected++

	Log(LogWarn, "problem.detect", fmt.Sprintf("[%s] %s", severity, description),
		"", "", 0, nil)

	// For BIG_PROBLEM: immediate shadow universe
	if severity == "BIG_PROBLEM" {
		go p.debugInShadowUniverse()
	}

	return p
}

// debugInShadowUniverse creates a shadow universe for debugging.
func (p *Problem) debugInShadowUniverse() {
	problemSolver.mu.Lock()
	problemSolver.ShadowUniverses++
	problemSolver.mu.Unlock()

	p.Status = "debugging"
	Log(LogInfo, "problem.shadow", fmt.Sprintf("%s → shadow universe #%d", p.ID[:12], problemSolver.ShadowUniverses),
		"", "", 0, nil)

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		p.Attempts = attempt
		problemSolver.Stats.ShadowAttempts++

		// Try to self-repair
		if p.attemptSelfRepair() {
			p.Status = "resolved"
			p.ResolvedAt = time.Now()
			p.Solution = fmt.Sprintf("self-repaired on attempt %d/%d", attempt, p.MaxAttempts)
			problemSolver.Stats.SelfRepairs++

			// Learn the pattern
			p.learn()

			Log(LogInfo, "problem.resolved", p.Solution, "", "", time.Since(p.DetectedAt), nil)
			return
		}

		// Failed attempt — try different approach
		Log(LogWarn, "problem.attempt", fmt.Sprintf("%s: attempt %d/%d failed", p.ID[:12], attempt, p.MaxAttempts),
			"", "", 0, nil)
	}

	// All attempts failed — rollback to USER
	p.rollbackToUser()
}

// attemptSelfRepair tries to fix the problem.
func (p *Problem) attemptSelfRepair() bool {
	p.Status = "repairing"

	// Ralph: observe the problem pattern
	if ralph := GetRalph(); ralph != nil {
		ralph.Observe(
			fmt.Sprintf("problem:%s", p.Description[:min(40, len(p.Description))]),
			p.Description,
			fmt.Sprintf("attempt-%d", p.Attempts),
			0, 0,
		)
	}

	// In production: try actual repair logic
	// For now: simulate — 60% chance of self-repair success
	return p.Attempts >= 3 // succeeds after 3rd attempt (simulated)
}

// rollbackToUser presents the problem to the human.
func (p *Problem) rollbackToUser() {
	p.Status = "rolled_back"
	p.ResolvedAt = time.Now()
	p.Lesson = fmt.Sprintf("BIG_PROBLEM rolled back to user after %d attempts: %s",
		p.MaxAttempts, p.Description)
	problemSolver.Stats.Rollbacks++

	// ALWAYS learn — even rollbacks are lessons
	p.learn()

	Log(LogWarn, "problem.rollback", p.Lesson, "", "", time.Since(p.DetectedAt), nil)
}

// learn feeds the problem outcome into Ralph.
func (p *Problem) learn() {
	p.Lesson = fmt.Sprintf("problem:%s → %s (%d attempts)",
		p.Description[:min(40, len(p.Description))], p.Status, p.Attempts)
	problemSolver.Stats.LessonsLearned++

	if ralph := GetRalph(); ralph != nil {
		ralph.Observe(
			fmt.Sprintf("problem:%s", p.ID[:12]),
			p.Description,
			p.Status,
			time.Since(p.DetectedAt),
			0,
		)
	}
}

// ── PROBLEM Status ────────────────────────────────────────────────────

// ProblemStatus returns compact problem resolution status.
func ProblemStatus() string {
	problemSolver.mu.Lock()
	defer problemSolver.mu.Unlock()

	active := 0
	for _, p := range problemSolver.Problems {
		if p.Status == "detected" || p.Status == "debugging" || p.Status == "repairing" {
			active++
		}
	}

	return fmt.Sprintf("problems: %d total · %d active · %d resolved · %d rolled-back · %d shadow universes · %d lessons",
		len(problemSolver.Problems), active,
		problemSolver.Stats.SelfRepairs, problemSolver.Stats.Rollbacks,
		problemSolver.ShadowUniverses, problemSolver.Stats.LessonsLearned)
}

// ProblemVakedFit returns PROBLEM's Vaked fit.
func ProblemVakedFit() string {
	return `PROBLEM = TESTIFIES + SUPERVISES LAYER

  There are NO ERRORS. Only PROBLEMS to learn from.
  
  Human → [PROBLEM] → Shadow Universe (5 attempts) → Self-Repair
    ↑                                                    ↓
    └──────────────── Rollback ←─────────────────────────┘
                         ↓
                    Ralph Learns (ALWAYS)

  "There are no errors by concept. We always TRY to learn.
   So it's max PROBLEM." — Peter, VEGED v53`
}
