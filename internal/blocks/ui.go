package blocks

import (
	"fmt"
	"strings"
)

// ── UI Primitive — AGUI × VAKED × TUI Blocks ─────────────────────────
// The UI layer is not just rendering — it's a capability-graph surface.
// UI blocks declare what they reveal, how they compose, and their state.

// UIBlock is a composable TUI/AGUI rendering primitive.
type UIBlock struct {
	ID       string    // unique block identifier
	Kind     string    // "infra_bar", "sidebar", "chat", "card_choice", "hud", "surface"
	Reveals  []string  // what Vaked layer this reveals (Declares, Supervises, etc.)
	Composes []string  // other UI blocks this block composes with
	State    string    // "visible", "collapsed", "hidden", "detached"
	Width    int       // current width
	Height   int       // current height
}

// ── UI Registry ───────────────────────────────────────────────────────

var uiRegistry = make(map[string]*UIBlock)

// RegisterUIBlock adds a UI block to the registry.
func RegisterUIBlock(b *UIBlock) {
	uiRegistry[b.ID] = b
}

// ── Built-in UI Blocks ────────────────────────────────────────────────

func init() {
	RegisterUIBlock(&UIBlock{ID: "infra_bar", Kind: "infra_bar",
		Reveals: []string{"Supervises", "Indexes"},
		Composes: []string{"sidebar", "hud"}, State: "visible"})

	RegisterUIBlock(&UIBlock{ID: "sidebar", Kind: "sidebar",
		Reveals: []string{"Supervises", "Declares"},
		Composes: []string{"infra_bar", "chat"}, State: "visible"})

	RegisterUIBlock(&UIBlock{ID: "chat", Kind: "chat",
		Reveals: []string{"Reveals"},
		Composes: []string{"infra_bar", "sidebar", "hud"}, State: "visible"})

	RegisterUIBlock(&UIBlock{ID: "card_choice", Kind: "card_choice",
		Reveals: []string{"Reveals"},
		Composes: []string{"chat"}, State: "hidden"})

	RegisterUIBlock(&UIBlock{ID: "hud", Kind: "hud",
		Reveals: []string{"Testifies", "Materializes"},
		Composes: []string{"infra_bar", "chat"}, State: "visible"})

	RegisterUIBlock(&UIBlock{ID: "surface", Kind: "surface",
		Reveals: []string{"Reveals"},
		Composes: []string{}, State: "visible"})
}

// ── UI Modes ──────────────────────────────────────────────────────────

// UIMode defines how the UI renders in different contexts.
type UIMode string

const (
	UIModeTUI      UIMode = "tui"      // full Bubble Tea terminal UI
	UIModeHeadless UIMode = "headless" // no rendering, API only
	UIModeDetached UIMode = "detached" // swarm/edge mode — minimal
	UIModeAGUI     UIMode = "agui"     // AG-UI themed rendering
)

// GetUIMode returns the current UI mode.
func GetUIMode() UIMode {
	_ = CurrentPOV()
	if IsHeadless() { return UIModeHeadless }
	// Detached mode for swarms (no TUI, but surface active)
	if GetDyad() != nil && !IsHeadless() { return UIModeTUI }
	return UIModeTUI
}

var headlessActive bool

func SetHeadless(v bool) { headlessActive = v }
func IsHeadless() bool   { return headlessActive }

// UIStatus returns compact UI registry status.
func UIStatus() string {
	var visible, hidden int
	for _, b := range uiRegistry {
		if b.State == "visible" { visible++ } else { hidden++ }
	}
	mode := GetUIMode()
	layers := make(map[string]bool)
	for _, b := range uiRegistry {
		for _, r := range b.Reveals { layers[r] = true }
	}
	return fmt.Sprintf("ui: %d blocks (%d visible, %d hidden) · mode: %s · reveals: %d vaked layers",
		len(uiRegistry), visible, hidden, mode, len(layers))
}

// ComposableLayout returns the current UI layout based on mode.
func ComposableLayout() string {
	mode := GetUIMode()
	switch mode {
	case UIModeHeadless:
		return "surface"
	case UIModeDetached:
		return "surface+hud"
	default:
		return "infra_bar+chat+sidebar+hud"
	}
}

// VakedLayersRevealed returns which Vaked layers are currently visible.
func VakedLayersRevealed() string {
	revealed := make(map[string]bool)
	for _, b := range uiRegistry {
		if b.State == "visible" {
			for _, r := range b.Reveals { revealed[r] = true }
		}
	}
	var layers []string
	for _, l := range []string{"Declares", "Materializes", "Supervises", "Enforces", "Testifies", "Indexes", "Reveals"} {
		if revealed[l] { layers = append(layers, l) }
	}
	return strings.Join(layers, " → ")
}
