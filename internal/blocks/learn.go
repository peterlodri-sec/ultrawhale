package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Learn Primitive — Pattern Learning Engine ─────────────────────────
// v14: Deepen The Triangle. Dedicated learning from patterns + outcomes.
// Feeds back into capability profiles, orchestration, Ralph, and dyad.

// Lesson is a learned pattern.
type Lesson struct {
	Pattern    string    // "auth→explore" — what we learned
	Confidence float64   // 0.0 → 1.0
	Source     string    // "ralph", "probe", "predict", "ultracode"
	AppliedAt  time.Time
	Outcome    string    // "success", "failure", "neutral"
	Adjustment string    // what was changed
}

// Learner accumulates lessons and applies them.
type Learner struct {
	mu       sync.Mutex
	lessons  []Lesson
	pov      POV
}

var learner = &Learner{pov: CurrentPOV()}

// Learn records a new lesson from any source.
func Learn(pattern string, confidence float64, source, outcome, adjustment string) Lesson {
	learner.mu.Lock()
	defer learner.mu.Unlock()

	lesson := Lesson{
		Pattern:    pattern,
		Confidence: confidence,
		Source:     source,
		AppliedAt:  time.Now(),
		Outcome:    outcome,
		Adjustment: adjustment,
	}
	learner.lessons = append(learner.lessons, lesson)
	if len(learner.lessons) > 256 { learner.lessons = learner.lessons[len(learner.lessons)-256:] }

	// Feed back into Ralph
	if ralph := GetRalph(); ralph != nil {
		ralph.Apply(pattern, adjustment, confidence)
	}

	Log(LogInfo, "learn."+source, pattern, "", "", 0, nil)
	return lesson
}

// LearnFromProbe learns from probe results.
func LearnFromProbe(probe Probe) {
	outcome := "success"
	if probe.Result != "pass" { outcome = "failure" }
	Learn(
		fmt.Sprintf("probe:%s:%s", probe.AgentID[:8], capName(probe.Capability)),
		boolToConf(probe.Result == "pass"),
		"probe", outcome,
		fmt.Sprintf("capability %s: %s", capName(probe.Capability), probe.Result),
	)
}

// LearnFromPrediction learns from prediction accuracy.
func LearnFromPrediction(pred Prediction, actualOutcome string) {
	Learn(
		fmt.Sprintf("predict:%s→%s", extractKeyword(pred.Prompt), pred.Agent),
		boolToConf(actualOutcome == "success"),
		"predict", actualOutcome,
		fmt.Sprintf("predicted %.0f%% confidence", pred.Confidence*100),
	)
}

func boolToConf(b bool) float64 { if b { return 0.9 }; return 0.1 }

// LearnStatus returns compact learning status.
func LearnStatus() string {
	learner.mu.Lock()
	defer learner.mu.Unlock()
	return fmt.Sprintf("learn: %d lessons from %d sources",
		len(learner.lessons), countSources(learner.lessons))
}

func countSources(lessons []Lesson) int {
	sources := make(map[string]bool)
	for _, l := range lessons { sources[l.Source] = true }
	return len(sources)
}
