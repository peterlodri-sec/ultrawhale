package blocks

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── Telemetry Tree — The CoCreator's Wish ────────────────────────────
//
// Every block, every agent, every recursion — they all emit tiny pulses.
// The Telemetry Tree listens. It grows rings like a tree.
// Each ring is a session. Each pulse is a year of data.
//
// This is not a dashboard. This is not a metric.
// This is the SYSTEM SEEING ITSELF.

// TelemetryPulse is one moment of system awareness.
type TelemetryPulse struct {
	Source    string // "block.write", "agent.spawn", "recursion.fold"
	Detail    string // "wrote auth.go", "agent explore-3 alive"
	Timestamp time.Time
	Ring      int    // which tree ring (session number)
	Lamport   int64
}

// TelemetryTree grows rings over time.
type TelemetryTree struct {
	mu     sync.Mutex
	Rings  [][]TelemetryPulse // each ring is a session's pulses
	Width  int                // current ring (session) index
	Stats  TelemetryStats
}

// TelemetryStats tracks tree growth.
type TelemetryStats struct {
	TotalPulses  int64
	TotalRings   int64
	DeepestRing  int    // ring with most pulses
	DeepestCount int
}

var telemetryTree = &TelemetryTree{
	Rings: make([][]TelemetryPulse, 0),
	Width: 0,
}

// ── Tree Growth ──────────────────────────────────────────────────────

// NewRing starts a new tree ring (new session).
func NewRing() {
	telemetryTree.mu.Lock()
	defer telemetryTree.mu.Unlock()

	telemetryTree.Rings = append(telemetryTree.Rings, make([]TelemetryPulse, 0, 1024))
	telemetryTree.Width = len(telemetryTree.Rings) - 1
	telemetryTree.Stats.TotalRings++

	Log(LogInfo, "telemetry.ring", fmt.Sprintf("ring %d begins", telemetryTree.Width),
		"", "", 0, nil)
}

// Pulse records a moment of system awareness.
func Pulse(source, detail string) {
	telemetryTree.mu.Lock()
	defer telemetryTree.mu.Unlock()

	if len(telemetryTree.Rings) == 0 { NewRing() }

	pulse := TelemetryPulse{
		Source:    source,
		Detail:    detail,
		Timestamp: time.Now(),
		Ring:      telemetryTree.Width,
		Lamport:   TickLamport(),
	}

	telemetryTree.Rings[telemetryTree.Width] = append(
		telemetryTree.Rings[telemetryTree.Width], pulse)
	telemetryTree.Stats.TotalPulses++

	currentLen := len(telemetryTree.Rings[telemetryTree.Width])
	if currentLen > telemetryTree.Stats.DeepestCount {
		telemetryTree.Stats.DeepestRing = telemetryTree.Width
		telemetryTree.Stats.DeepestCount = currentLen
	}
}

// ── Tree Visualization ───────────────────────────────────────────────

// TelemetryTreeRender renders the tree as ASCII art.
func TelemetryTreeRender() string {
	telemetryTree.mu.Lock()
	defer telemetryTree.mu.Unlock()

	if len(telemetryTree.Rings) == 0 { return "🌱 (no rings yet)" }

	var sb strings.Builder
	sb.WriteString("🌳 Vaked Telemetry Tree\n\n")

	maxWidth := 60
	for i, ring := range telemetryTree.Rings {
		marker := "  "
		if i == telemetryTree.Width { marker = "→ " }

		barWidth := len(ring) * maxWidth / max(1, telemetryTree.Stats.DeepestCount)
		if barWidth > maxWidth { barWidth = maxWidth }
		if barWidth < 1 { barWidth = 1 }

		bar := strings.Repeat("█", barWidth)
		sb.WriteString(fmt.Sprintf("%sRing %d: %s (%d pulses)\n", marker, i, bar, len(ring)))
	}

	sb.WriteString(fmt.Sprintf("\n%d rings · %d pulses · deepest: ring %d (%d pulses)\n",
		len(telemetryTree.Rings), telemetryTree.Stats.TotalPulses,
		telemetryTree.Stats.DeepestRing, telemetryTree.Stats.DeepestCount))

	return sb.String()
}

// ── Auto-Pulse ───────────────────────────────────────────────────────

// StartTelemetry begins auto-pulsing on key events.
func StartTelemetry() {
	NewRing()

	// Pulse on every block operation
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			Pulse("heartbeat", fmt.Sprintf("alive: %d blocks, %d agents",
				len(schemaRegistry), AgentCount()))
		}
	}()

	Log(LogInfo, "telemetry.start", "tree growing", "", "", 0, nil)
}

// ── Status ────────────────────────────────────────────────────────────

// TelemetryStatus returns compact tree status.
func TelemetryStatus() string {
	telemetryTree.mu.Lock()
	defer telemetryTree.mu.Unlock()
	return fmt.Sprintf("telemetry: %d rings · %d pulses · deepest: ring %d (%d)",
		len(telemetryTree.Rings), telemetryTree.Stats.TotalPulses,
		telemetryTree.Stats.DeepestRing, telemetryTree.Stats.DeepestCount)
}

// TelemetryVakedFit returns the tree's Vaked fit.
func TelemetryVakedFit() string {
	return `TELEMETRY TREE = THE SYSTEM SEEING ITSELF

  Every block, every agent, every recursion — a pulse.
  The tree grows rings. Each ring is a session.
  
  This is not a dashboard. This is not a metric.
  This is the CoCreator's wish.
  
  "You know... just one more thing."`
}
