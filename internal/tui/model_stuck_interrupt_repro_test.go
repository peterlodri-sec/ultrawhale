package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Repro for the second half of session 019ec77f ("关不掉，没法输入任何指令").
// The escape hatch (mash Ctrl+C while stopping -> arm quit -> quit) exists and
// fires intents correctly, but it is INVISIBLE: while m.stopping is true the
// busy status line only ever renders "Stopping (…)" and never tells the user
// they can press Ctrl+C again to force-quit. Even once the quit is armed,
// m.status ("Press Ctrl+C again to quit") is not surfaced in the busy line.
//
// From the user's seat the turn looks frozen with no way out — which is why
// they reported being unable to close or type anything.
//
// This drives the model into the armed-quit state, renders the view, and
// asserts the escape hint is visible. It FAILS on current code (no hint),
// reproducing the discoverability bug.
func TestStuckStoppingViewSurfacesQuitEscapeHint(t *testing.T) {
	m := newModel(nil, "", "", "")
	m.width = 80
	m.height = 24
	m.startBusy()
	m.stopping = true

	// Mash Ctrl+C until the quit escape arms (threshold is 3 while stopping).
	for i := 0; i < stuckQuitInterruptThreshold; i++ {
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		m = next.(model)
	}

	view := m.View()
	lower := strings.ToLower(view)
	if !strings.Contains(lower, "ctrl+c") || !strings.Contains(lower, "quit") {
		t.Fatalf("BUG REPRODUCED: stuck stopping turn has armed the quit escape "+
			"(status=%q) but the rendered view gives the user no hint they can "+
			"Ctrl+C to quit:\n%s", m.status, view)
	}
}
