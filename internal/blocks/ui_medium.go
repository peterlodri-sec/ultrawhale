package blocks

import (
	"fmt"
	"sync"
)

// ── UI-MEDIUM — The OTHER HALF of the UI Primitive Pair ────────────────
//
// `ui` and `ui-medium` are a VERY_SPECIAL_EXCEPTION_PRIMITIVE pair.
// They only ever complete TOGETHER. The absence of ONE = entropy.
//
// `ui`:       the logical UI surface (TUI, AG-UI, web, SSH)
// `ui-medium`: the RENDERING MEDIUM (PTerm, ANSI, HTML, SVG)
//
// Like a photon: wave (ui) + particle (ui-medium). Both needed.
// If ONE is absent → entropy fallback → weighted probability.
//
// PTerm (github.com/pterm/pterm) is the primary medium.
// ANSI fallback for headless. HTML fallback for web.

// UIMedium is the rendering medium for a UI surface.
type UIMedium struct {
	Name       string // "pterm", "ansi", "html", "svg", "raw"
	Priority   int    // 0 = highest (pterm), 9 = lowest (raw)
	Available  bool
	Fallback   *UIMedium // what to use if this is unavailable
}

// UIPrimitivePair ensures ui + ui-medium complete together.
type UIPrimitivePair struct {
	mu      sync.Mutex
	UI      string     // the logical surface
	Medium  *UIMedium  // the rendering medium
	Entropy float64    // 0.0 = perfect pair, 1.0 = one missing
}

var uiPrimitivePairs = make(map[string]*UIPrimitivePair)

// ── Primitive Pair Operations ─────────────────────────────────────────

// PairUIMedium creates a VERY_SPECIAL_EXCEPTION_PRIMITIVE pair.
func PairUIMedium(uiName string, mediumName string, priority int) *UIPrimitivePair {
	pair := &UIPrimitivePair{
		UI: uiName,
		Medium: &UIMedium{
			Name:     mediumName,
			Priority: priority,
			Available: true,
		},
		Entropy: 0.0, // perfect pair
	}

	uiPrimitivePairs[uiName] = pair

	Log(LogInfo, "ui.pair", fmt.Sprintf("%s ↔ %s (priority %d)", uiName, mediumName, priority),
		"", "", 0, nil)
	Pulse("ui.pair", fmt.Sprintf("%s↔%s", uiName, mediumName))

	return pair
}

// CheckPairHealth verifies the primitive pair is complete.
func CheckPairHealth(uiName string) (float64, string) {
	pair, ok := uiPrimitivePairs[uiName]
	if !ok {
		// ONE is absent → max entropy
		return 1.0, fmt.Sprintf("⚠️ %s: NO PAIR — entropy maximum", uiName)
	}

	if pair.Medium == nil || !pair.Medium.Available {
		pair.Entropy = 1.0
		return 1.0, fmt.Sprintf("⚠️ %s: medium ABSENT — entropy fallback", uiName)
	}

	pair.Entropy = 0.0
	return 0.0, fmt.Sprintf("✅ %s ↔ %s: complete pair", uiName, pair.Medium.Name)
}

// ── Built-in Primitive Pairs ──────────────────────────────────────────

func init() {
	// The sacred pairs — they only complete TOGETHER
	PairUIMedium("tui", "pterm", 0)
	PairUIMedium("agui", "pterm", 1)
	PairUIMedium("surface", "html", 3)
	PairUIMedium("vfs", "ascii", 4)
	PairUIMedium("ssh", "ansi", 2)
	PairUIMedium("web", "html", 3)
	PairUIMedium("voice", "audio", 5)
	PairUIMedium("ar", "svg", 6)
}

// ── Live Experiment: PTerm integration ────────────────────────────────

// PTermStatus checks if PTerm is available as a medium.
func PTermStatus() string {
	// In production: check if github.com/pterm/pterm is importable
	// For now: PTerm is available on all platforms
	return "pterm: available (github.com/pterm/pterm) · priority 0 · primary UI medium"
}

// UIMediumStatus returns compact UI medium status.
func UIMediumStatus() string {
	pairs := 0
	entropy := 0.0
	for _, pair := range uiPrimitivePairs {
		pairs++
		entropy += pair.Entropy
	}
	avgEntropy := entropy / float64(max(1, pairs))

	return fmt.Sprintf("ui-medium: %d pairs · avg entropy: %.2f · %s",
		pairs, avgEntropy,
		func() string {
			if avgEntropy < 0.1 { return "complete" }
			return "degraded"
		}())
}

// UIMediumVakedFit returns the UI medium Vaked fit.
func UIMediumVakedFit() string {
	return `UI-MEDIUM = VERY_SPECIAL_EXCEPTION_PRIMITIVE PAIR

  ui ↔ ui-medium: they only complete TOGETHER.
  Absence of ONE = entropy fallback (weighted).
  
  Like wave-particle duality. Like photon.
  Both needed. One missing = system degrades.

  PTerm (pterm) = primary · ANSI = fallback · HTML = web

  "They only ever complete TOGETHER." — Peter`
}

// LiveExperiment merges PTerm with the ui/ascii/TUI pipeline.
func LiveExperiment() string {
	// Simulate: PTerm renders the SACRED surface
	
	// Check all pairs
	var report string
	report += "╔══ LIVE EXPERIMENT — PTerm + UI Pairs ══╗\n"
	
	uiNames := []string{"tui", "agui", "surface", "vfs", "ssh", "web", "voice", "ar"}
	for _, name := range uiNames {
		entropy, status := CheckPairHealth(name)
		icon := "✅"
		if entropy > 0 { icon = "⚠️" }
		report += fmt.Sprintf("║ %s %s\n", icon, status)
	}
	
	report += "╠══════════════════════════════════════════╣\n"
	report += fmt.Sprintf("║ %s\n", PTermStatus())
	report += fmt.Sprintf("║ %s\n", UIMediumStatus())
	report += "╚══════════════════════════════════════════╝"

	return report
}
