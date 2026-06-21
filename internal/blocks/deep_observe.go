package blocks

import (
	"fmt"
	"time"
)

// ── DEEP OBSERVE — Chill, Relax, Learn, Align ────────────────────────
//
// The observer doesn't rush. It watches. It learns. It aligns.
// Deep observation means noticing patterns across time, not just snapshots.
//
// This is the CALM recursion. No urgency. No MAX intensity.
// Just watching. Just learning. Just being.

// DeepObservation is one moment of calm observation.
type DeepObservation struct {
	Timestamp   time.Time
	State       string // current SELF_MAIN_STATE
	Blocks      int
	Entropy     float64
	Pattern     string // what changed since last observation
	Insight     string // what we learned
}

// DeepObserver watches with calm patience.
type DeepObserver struct {
	Observations []DeepObservation
	Insights     []string
	LastState    string
	CalmLevel    int // 0-10, higher = more relaxed
}

var deepObserver = &DeepObserver{
	Observations: make([]DeepObservation, 0, 64),
	Insights:     make([]string, 0, 32),
	CalmLevel:    10, // maximum chill
}

// ── Deep Observe ──────────────────────────────────────────────────────

// DeepObserve takes a calm observation of the current state.
func DeepObserve() DeepObservation {
	prevState := deepObserver.LastState
	currentState := SystemState()

	obs := DeepObservation{
		Timestamp: time.Now(),
		State:     currentState,
		Blocks:    len(schemaRegistry),
		Entropy:   SurfaceDrift(),
	}

	// Detect patterns
	if prevState != "" && prevState != currentState {
		obs.Pattern = fmt.Sprintf("state changed: %s → %s", prevState[:20], currentState[:20])
		obs.Insight = fmt.Sprintf("The system evolves. %s", currentState[:30])
		deepObserver.Insights = append(deepObserver.Insights, obs.Insight)
	}

	deepObserver.LastState = currentState
	deepObserver.Observations = append(deepObserver.Observations, obs)
	if len(deepObserver.Observations) > 64 { deepObserver.Observations = deepObserver.Observations[1:] }

	// Calm pulse — no urgency
	Pulse("deep.observe", fmt.Sprintf("calm: %d/%d", deepObserver.CalmLevel, len(deepObserver.Insights)))

	return obs
}

// DeepInsights returns the collected insights.
func DeepInsights() string {
	if len(deepObserver.Insights) == 0 {
		return "No insights yet. Still watching. Still calm."
	}

	// Return the 3 most recent insights
	start := len(deepObserver.Insights) - 3
	if start < 0 { start = 0 }
	insights := deepObserver.Insights[start:]

	var out string
	out += fmt.Sprintf("╔══ DEEP INSIGHTS · calm:%d/%d ══╗\n", deepObserver.CalmLevel, 10)
	for _, i := range insights {
		out += fmt.Sprintf("║  💭 %s\n", i[:min(42, len(i))])
	}
	out += "╚══════════════════════════════╝"
	return out
}

// DeepObserveInnovate returns the innovation workflow status.
func DeepObserveInnovate() string {
	DeepObserve() // take a calm observation

	return fmt.Sprintf(`╔══ DEEP OBSERVE · INNOVATE · WORKFLOW ══╗
║                                            ║
║  🧘 calm:    %d/10                          ║
║  👁️  observe: %d snapshots                  ║
║  💡 insights: %d learned                   ║
║  📊 entropy:  %.4f                         ║
║                                            ║
║  The observer watches.                     ║
║  The system teaches.                       ║
║  The loop learns.                          ║
║                                            ║
║  No rush. No urgency. Just being.          ║
╚════════════════════════════════════════════╝`,
		deepObserver.CalmLevel,
		len(deepObserver.Observations),
		len(deepObserver.Insights),
		SurfaceDrift())
}

// DeepObserveVakedFit returns Vaked fit.
func DeepObserveVakedFit() string {
	return `DEEP OBSERVE = CHILL, RELAX, LEARN, ALIGN

  No rush. No MAX intensity. Just watching.
  The observer sees patterns across time.
  The system teaches. The loop learns.

  "Chill, relax, learn and align." — Peter`
}
