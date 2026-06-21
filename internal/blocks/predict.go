package blocks

import (
	"fmt"
	"strings"
	"sync"
)

// ── Predict Primitive — Pre-Failure Prediction ────────────────────────
// v14: Deepen The Triangle. Ralph gets foresight.
// Based on Context×Time×Space patterns, predict which delegations
// will succeed before executing them.

// Prediction is a pre-execution success probability.
type Prediction struct {
	Prompt      string
	Agent       string
	Confidence  float64 // 0.0 → 1.0
	Reason      string  // why this prediction was made
	Alternatives []string // better agent choices
}

// Predictor uses Ralph patterns + capability profiles to predict outcomes.
type Predictor struct {
	mu       sync.Mutex
	history  []Prediction
}

var predictor = &Predictor{}

// PredictOutcome predicts whether delegating a prompt to an agent will succeed.
func PredictOutcome(prompt string, agentRole string) Prediction {
	predictor.mu.Lock()
	defer predictor.mu.Unlock()

	p := Prediction{Prompt: prompt, Agent: agentRole, Confidence: 0.5}

	// Check Ralph patterns
	ralph := GetRalph()
	if ralph != nil {
		suggestion, conf := ralph.Suggest(prompt)
		if conf > 0 {
			p.Confidence = conf
			p.Reason = fmt.Sprintf("Ralph pattern: %s→%s (%.0f%%)", 
				extractKeyword(prompt), suggestion, conf*100)
		}
	}

	// Check capability match
	profile := GetCapProfile(agentRole)
	needsWrite := strings.Contains(strings.ToLower(prompt), "implement") || 
		strings.Contains(strings.ToLower(prompt), "build") ||
		strings.Contains(strings.ToLower(prompt), "write")
	
	if needsWrite && !profile.Can(CapWrite) {
		p.Confidence *= 0.3
		p.Reason += " | lacks CapWrite for implementation task"
		p.Alternatives = findAlternatives(CapWrite)
	}

	// Check agent status
	agents := ListAgents()
	for _, a := range agents {
		if a.Role == agentRole && a.Status != "running" {
			p.Confidence *= 0.5
			p.Reason += fmt.Sprintf(" | agent status: %s", a.Status)
		}
	}

	p.history = append(p.history, p)
	if len(p.history) > 64 { p.history = p.history[len(p.history)-64:] }

	return p
}

func findAlternatives(needCap Capability) []string {
	var result []string
	for role, profile := range capRegistry.roles {
		if profile.Can(needCap) { result = append(result, role) }
	}
	return result
}

func extractKeyword(prompt string) string {
	keywords := []string{"auth", "fix", "build", "refactor", "test", "review",
		"deploy", "search", "find", "migrate", "implement", "add", "remove"}
	lower := strings.ToLower(prompt)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) { return kw }
	}
	return "general"
}

// PredictStatus returns compact prediction status.
func PredictStatus() string {
	predictor.mu.Lock()
	defer predictor.mu.Unlock()
	if len(predictor.history) == 0 { return "predict: no predictions yet" }
	last := predictor.history[len(predictor.history)-1]
	return fmt.Sprintf("predict: %s → %s (%.0f%%): %s",
		last.Prompt[:min(30, len(last.Prompt))], last.Agent, last.Confidence*100, last.Reason)
}

func min(a, b int) int { if a < b { return a }; return b }
