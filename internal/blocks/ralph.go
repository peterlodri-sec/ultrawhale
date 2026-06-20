package blocks

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── Ralph Loop — Self-Improving Agent Cycle ──────────────────────────
// Versioned per orchestrator session. Rollback-able.
// Bad adjustments never persist — each session starts clean or
// can be rolled back to any previous version.

// RalphLoop is the self-improving observation→adjustment cycle.
type RalphLoop struct {
	mu        sync.RWMutex
	SessionID string   // orchestrator session this ralph belongs to
	POV       POV     // current execution context

	// Versioned state
	Version   int               // monotonic — incremented on each adjustment
	Snapshots []RalphSnapshot   // rollback history (last 16 versions)

	// Current state
	Cycles      []RalphCycle              // last 64 observation cycles
	Patterns    map[string]float64        // "auth→explore" → 0.85 confidence
	Adjustments map[string]RalphAdjustment // active adjustments

	// Stats
	TotalObservations int64
	TotalAdjustments  int64
	Rollbacks         int64
	ConsecutiveFailures int
}

// RalphCycle is one observe→learn→adjust cycle.
type RalphCycle struct {
	Seq        int           `json:"seq"`
	Timestamp  time.Time     `json:"ts"`
	Prompt     string        `json:"prompt"`
	Decision   string        `json:"decision"`   // "delegated to swe", "cached"
	Outcome    string        `json:"outcome"`    // "success", "failed", "timeout"
	Latency    time.Duration `json:"latency_ms"`
	TokensUsed int64         `json:"tokens"`

	// Learned
	Pattern    string  `json:"pattern,omitempty"`    // extracted pattern
	Confidence float64 `json:"confidence,omitempty"` // 0.0→1.0

	// Applied
	Adjustment string `json:"adjustment,omitempty"` // what was changed
	Applied    bool   `json:"applied"`
}

// RalphAdjustment is an active behavioral change.
type RalphAdjustment struct {
	Key        string    `json:"key"`        // "classify:auth" or "tool:grep:cache_ttl"
	Value      string    `json:"value"`      // "explore" or "10m"
	Confidence float64   `json:"confidence"`
	AppliedAt  time.Time `json:"applied_at"`
	RevertedAt time.Time `json:"reverted_at,omitempty"`
}

// RalphSnapshot is a point-in-time state for rollback.
type RalphSnapshot struct {
	Version     int
	At          time.Time
	Patterns    map[string]float64
	Adjustments map[string]RalphAdjustment
	Description string
}

// NewRalphLoop creates a ralph loop for a session.
func NewRalphLoop(sessionID string) *RalphLoop {
	r := &RalphLoop{
		POV:       CurrentPOV(),
		SessionID:   sessionID,
		Version:     1,
		Cycles:      make([]RalphCycle, 0, 64),
		Patterns:    make(map[string]float64),
		Adjustments: make(map[string]RalphAdjustment),
		Snapshots:   make([]RalphSnapshot, 0, 16),
	}
	r.snapshot("initial")
	return r
}

// ── Observe → Learn → Adjust ──────────────────────────────────────────

// Observe records a decision and its outcome.
func (r *RalphLoop) Observe(prompt, decision, outcome string, latency time.Duration, tokens int64) RalphCycle {
	r.mu.Lock()
	defer r.mu.Unlock()

	cycle := RalphCycle{
		Seq:        len(r.Cycles),
		Timestamp:  time.Now(),
		Prompt:     truncate(prompt, 80),
		Decision:   decision,
		Outcome:    outcome,
		Latency:    latency,
		TokensUsed: tokens,
	}

	// Learn from outcome
	if outcome == "failed" || outcome == "timeout" {
		r.ConsecutiveFailures++
		cycle.Pattern = extractPattern(prompt, decision)
		cycle.Confidence = r.bumpPattern(cycle.Pattern, -0.1)

		// Auto-rollback on 3 consecutive failures
		if r.ConsecutiveFailures >= 3 {
			r.Rollback(r.Version - 1)
			cycle.Adjustment = fmt.Sprintf("auto-rollback to v%d after %d failures", r.Version, r.ConsecutiveFailures)
			r.ConsecutiveFailures = 0
		}

		// Suggest adjustment
		if cycle.Confidence < 0.3 {
			cycle.Adjustment = r.suggestAdjustment(cycle.Pattern, decision)
		}
	} else {
		r.ConsecutiveFailures = 0
		cycle.Pattern = extractPattern(prompt, decision)
		cycle.Confidence = r.bumpPattern(cycle.Pattern, +0.05)
	}

	r.Cycles = append(r.Cycles, cycle)
	if len(r.Cycles) > 64 {
		r.Cycles = r.Cycles[len(r.Cycles)-64:]
	}
	r.TotalObservations++

	return cycle
}

// Apply commits an adjustment.
func (r *RalphLoop) Apply(key, value string, confidence float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Version++
	r.Adjustments[key] = RalphAdjustment{
		Key: key, Value: value, Confidence: confidence, AppliedAt: time.Now(),
	}
	r.TotalAdjustments++

	// Persist to brain long-term memory
	brain := GetBrain()
	if brain != nil {
		brain.RememberLongTerm(map[string]string{
			"ralph_key":   key,
			"ralph_value": value,
			"confidence":  fmt.Sprintf("%.2f", confidence),
		})
	}

	r.snapshot(fmt.Sprintf("applied %s=%s", key, value))
}

// ── Suggest ────────────────────────────────────────────────────────────

// Suggest returns the best action for a prompt based on learned patterns.
func (r *RalphLoop) Suggest(prompt string) (string, float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	best := ""
	bestConf := 0.0
	for pattern, conf := range r.Patterns {
		if strings.Contains(prompt, pattern) && conf > bestConf {
			best = r.patternToAction(pattern)
			bestConf = conf
		}
	}
	return best, bestConf
}

// ── Rollback ───────────────────────────────────────────────────────────

// Rollback restores to a previous version.
func (r *RalphLoop) Rollback(targetVersion int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := len(r.Snapshots) - 1; i >= 0; i-- {
		if r.Snapshots[i].Version <= targetVersion {
			snap := r.Snapshots[i]
			r.Patterns = snap.Patterns
			r.Adjustments = snap.Adjustments
			r.Version = snap.Version
			r.Rollbacks++
			r.snapshot(fmt.Sprintf("rolled back to v%d", targetVersion))
			return true
		}
	}
	return false
}

// ── Status ─────────────────────────────────────────────────────────────

// RalphStatus returns a multi-line status for AG-UI display.
func (r *RalphLoop) RalphStatus() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var lines []string
	lines = append(lines, fmt.Sprintf("ralph v%d · %d cycles · %d patterns · %d adjustments · %d rollbacks",
		r.Version, len(r.Cycles), len(r.Patterns), len(r.Adjustments), r.Rollbacks))

	if len(r.Adjustments) > 0 {
		lines = append(lines, "active adjustments:")
		for k, a := range r.Adjustments {
			if a.RevertedAt.IsZero() {
				lines = append(lines, fmt.Sprintf("  %s → %s (%.0f%%)", k, a.Value, a.Confidence*100))
			}
		}
	}

	// Show top patterns
	if len(r.Patterns) > 0 {
		type kv struct {
			k string
			v float64
		}
		var sorted []kv
		for k, v := range r.Patterns {
			sorted = append(sorted, kv{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })

		lines = append(lines, "top patterns:")
		for i, p := range sorted {
			if i >= 3 { break }
			lines = append(lines, fmt.Sprintf("  %s: %.0f%%", p.k, p.v*100))
		}
	}

	// Last cycle
	if len(r.Cycles) > 0 {
		last := r.Cycles[len(r.Cycles)-1]
		lines = append(lines, fmt.Sprintf("last: %s → %s (%s, %s)",
			truncate(last.Prompt, 40), last.Decision, last.Outcome, last.Latency.Round(time.Millisecond)))
	}

	return strings.Join(lines, "\n")
}

// ── Internals ──────────────────────────────────────────────────────────

func (r *RalphLoop) bumpPattern(pattern string, delta float64) float64 {
	v := r.Patterns[pattern] + delta
	if v < 0 { v = 0 }
	if v > 1 { v = 1 }
	r.Patterns[pattern] = v
	return v
}

func (r *RalphLoop) suggestAdjustment(pattern, decision string) string {
	// Pattern: "auth→swe". Failed. Suggest: "auth→explore"
	parts := strings.Split(pattern, "→")
	if len(parts) == 2 && parts[1] == "swe" {
		key := fmt.Sprintf("classify:%s", parts[0])
		value := "explore"
		r.Adjustments[key] = RalphAdjustment{
			Key: key, Value: value,
			Confidence: r.Patterns[pattern],
			AppliedAt:  time.Now(),
		}
		return fmt.Sprintf("%s → %s", key, value)
	}
	return ""
}

func (r *RalphLoop) patternToAction(pattern string) string {
	parts := strings.Split(pattern, "→")
	if len(parts) == 2 { return parts[1] }
	return pattern
}

func (r *RalphLoop) snapshot(desc string) {
	snap := RalphSnapshot{
		Version:     r.Version,
		At:          time.Now(),
		Patterns:    copyMap(r.Patterns),
		Adjustments: copyAdj(r.Adjustments),
		Description: desc,
	}
	r.Snapshots = append(r.Snapshots, snap)
	if len(r.Snapshots) > 16 {
		r.Snapshots = r.Snapshots[len(r.Snapshots)-16:]
	}
}

// ── Helpers ────────────────────────────────────────────────────────────

func extractPattern(prompt, decision string) string {
	// Extract a keyword from prompt + what we delegated to
	word := firstKeyword(prompt)
	return fmt.Sprintf("%s→%s", word, decision)
}

func firstKeyword(prompt string) string {
	keywords := []string{"auth", "fix", "build", "refactor", "test", "review",
		"deploy", "search", "find", "migrate", "implement", "add", "remove"}
	lower := strings.ToLower(prompt)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) { return kw }
	}
	return "general"
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n-3] + "..."
}

func copyMap(m map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for k, v := range m { out[k] = v }
	return out
}

func copyAdj(m map[string]RalphAdjustment) map[string]RalphAdjustment {
	out := make(map[string]RalphAdjustment, len(m))
	for k, v := range m { out[k] = v }
	return out
}

// ── Session Ralph ──────────────────────────────────────────────────────

var sessionRalph *RalphLoop

func GetRalph() *RalphLoop {
	if sessionRalph == nil {
		sessionRalph = NewRalphLoop("default")
	}
	return sessionRalph
}

func InitRalph(sessionID string) *RalphLoop {
	sessionRalph = NewRalphLoop(sessionID)
	return sessionRalph
}
