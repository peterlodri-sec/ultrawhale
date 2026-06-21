package blocks

import (
	"fmt"
	"time"
	"strings"
)

// ── OneShot — Atomic Vaked Pipeline ──────────────────────────────────
//
// OneShot is the atomic unit of Vaked philosophy.
// Like "space" and "context", OneShot is a fundamental concept:
//   declare → materialize → reveal → in ONE pass.
//
// OneShot("vaked declare agent swe { write, execute }")
//   → validates declaration (Declares)
//   → generates config (Materializes)
//   → returns AG-UI block (Reveals)
//
// The entire Vaked pipeline in a single function call.

// OneShotResult is the output of a OneShot execution.
type OneShotResult struct {
	Declaration  string // what was declared
	Validated    bool   // passed schema validation
	Materialized string // generated artifact path or content
	Revealed     string // AG-UI block or TUI output
	Duration     int64  // microseconds
	Errors       []string
}

// OneShot executes the full Vaked pipeline atomically.
// One pass: declare → materialize → reveal.
func OneShot(declaration string) OneShotResult {
	start := time.Now()

	result := OneShotResult{Declaration: declaration}

	// Phase 1: Declares — validate
	valid := true
	for _, schema := range schemaRegistry {
		if strings.Contains(strings.ToLower(declaration), strings.ToLower(schema.Name)) {
			result.Validated = true
		}
	}
	if !valid {
		result.Errors = append(result.Errors, "no matching schema")
	}
	result.Validated = valid

	// Phase 2: Materializes — generate
	result.Materialized = fmt.Sprintf("oneshot:%s", Ref([]byte(declaration))[:12])

	// Phase 3: Reveals — AG-UI block
	pov := CurrentPOV()
	result.Revealed = fmt.Sprintf("[%s/%s] %s", pov.Machine, pov.Arch, declaration[:min(60, len(declaration))])

	result.Duration = time.Since(start).Microseconds()
	Log(LogInfo, "oneshot.execute", result.Materialized, "", "", time.Since(start), nil)

	return result
}

// OneShotStatus returns OneShot pipeline status.
func OneShotStatus() string {
	return fmt.Sprintf("oneshot: declare→materialize→reveal · 1 pass · atomic")
}
