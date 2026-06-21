package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── EVOLVE — The 4th Recursion ──────────────────────────────────────
//
// EVOLVE recurses through VERSIONS.
// Each release learns from the previous.
// Ralph patterns persist across versions.
// The capability graph grows without bound.
//
//   v39 → EVOLVE → v40 → EVOLVE → ... → v50
//     ↑                                          ↓
//     └────────── patterns persist ──────────────┘

// Evolution is one step in the version recursion.
type Evolution struct {
	FromVersion string
	ToVersion   string
	Changes     []string // what changed
	Lessons     []string // what we learned
	AppliedAt   time.Time
}

// EvolveEngine manages version evolution.
type EvolveEngine struct {
	mu        sync.Mutex
	History   []Evolution
	Version   string
	Stats     EvolveStats
}

type EvolveStats struct {
	Evolutions   int64
	LessonsKept  int64
	PatternsKept int64
}

var evolveEngine = &EvolveEngine{
	History: make([]Evolution, 0, 64),
	Version: CurrentVersion(),
}

// Evolve records an evolution step.
func Evolve(fromVersion, toVersion string, changes, lessons []string) Evolution {
	evolveEngine.mu.Lock()
	defer evolveEngine.mu.Unlock()

	ev := Evolution{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Changes:     changes,
		Lessons:     lessons,
		AppliedAt:   time.Now(),
	}

	evolveEngine.History = append(evolveEngine.History, ev)
	evolveEngine.Stats.Evolutions++
	evolveEngine.Stats.LessonsKept += int64(len(lessons))
	evolveEngine.Version = toVersion

	// Feed all lessons into Ralph for cross-version pattern persistence
	if ralph := GetRalph(); ralph != nil {
		for _, lesson := range lessons {
			ralph.Observe(fmt.Sprintf("evolve:%s", lesson), lesson, "persisted", 0, 0)
			evolveEngine.Stats.PatternsKept++
		}
	}

	Log(LogInfo, "evolve.step", fmt.Sprintf("%s → %s (%d changes, %d lessons)",
		fromVersion, toVersion, len(changes), len(lessons)), "", "", 0, nil)

	return ev
}

// EvolveStatus returns compact evolve status.
func EvolveStatus() string {
	evolveEngine.mu.Lock()
	defer evolveEngine.mu.Unlock()
	return fmt.Sprintf("evolve: %d steps · %d lessons · %d patterns · v%s",
		evolveEngine.Stats.Evolutions, evolveEngine.Stats.LessonsKept,
		evolveEngine.Stats.PatternsKept, evolveEngine.Version)
}

// EvolveVakedFit returns EVOLVE's Vaked fit.
func EvolveVakedFit() string {
	return `EVOLVE = THE 4TH RECURSION (through VERSIONS)

  Full-Stop → layers  → SACRED
  Fold      → agents  → leaf
  Heal      → checks  → resolved
  EVOLVE    → versions → v50

  Each release learns from the previous.
  Ralph patterns persist across versions.
  The capability graph grows without bound.

  v39 → EVOLVE → v40 → ... → v50 → ...`
}
