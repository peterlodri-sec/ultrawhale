// Package agui implements AG-UI themed rendering for Whale's Bubble Tea TUI.
// Ported from AG-UI/Core/Navigation/ViewportLayoutSchema.swift.
// GENESIS_SEAL: 7c242080
package agui

import "github.com/charmbracelet/lipgloss"

// ThemeName identifies an AG-UI theme.
type ThemeName string

const (
	DenseMatrixGreen    ThemeName = "dense"
	CleanGraphCyberpunk ThemeName = "cyberpunk"
	TacticalGraveyard   ThemeName = "graveyard"
)

// Theme holds the AG-UI color palette for one theme.
// Flat struct — data-oriented design, no methods.
type Theme struct {
	Name    ThemeName
	Bg      lipgloss.Color // background
	Surface lipgloss.Color // card/panel surface
	Fg      lipgloss.Color // primary text
	Dim     lipgloss.Color // secondary/subdued text
	Accent  lipgloss.Color // accent/highlight
	Border  lipgloss.Color // borders and dividers
}

// Themes is the array of all AG-UI themes, indexed by name.
var Themes = map[ThemeName]Theme{
	DenseMatrixGreen: {
		Name:    DenseMatrixGreen,
		Bg:      lipgloss.Color("#040804"),
		Surface: lipgloss.Color("#0a140a"),
		Fg:      lipgloss.Color("#c8f5c8"),
		Dim:     lipgloss.Color("#5a8c5a"),
		Accent:  lipgloss.Color("#00e660"),
		Border:  lipgloss.Color("#143a14"),
	},
	CleanGraphCyberpunk: {
		Name:    CleanGraphCyberpunk,
		Bg:      lipgloss.Color("#0a0a14"),
		Surface: lipgloss.Color("#14141f"),
		Fg:      lipgloss.Color("#e0e8f5"),
		Dim:     lipgloss.Color("#6878a0"),
		Accent:  lipgloss.Color("#00d4ff"),
		Border:  lipgloss.Color("#26304a"),
	},
	TacticalGraveyard: {
		Name:    TacticalGraveyard,
		Bg:      lipgloss.Color("#141414"),
		Surface: lipgloss.Color("#1e1e1e"),
		Fg:      lipgloss.Color("#d4d4d4"),
		Dim:     lipgloss.Color("#7a7a7a"),
		Accent:  lipgloss.Color("#b0b0b0"),
		Border:  lipgloss.Color("#333333"),
	},
}

// Current is the active theme. Changed via Ctrl+Shift+T in the TUI.
var Current = Themes[DenseMatrixGreen]

// CycleTheme advances to the next theme and returns its name.
func CycleTheme() ThemeName {
	order := []ThemeName{DenseMatrixGreen, CleanGraphCyberpunk, TacticalGraveyard}
	for i, t := range order {
		if t == Current.Name {
			next := order[(i+1)%len(order)]
			Current = Themes[next]
			return next
		}
	}
	Current = Themes[DenseMatrixGreen]
	return DenseMatrixGreen
}

// SetTheme activates a theme by name.
func SetTheme(name ThemeName) {
	if t, ok := Themes[name]; ok {
		Current = t
	}
}

// Style returns a lipgloss.Style with the given foreground and background colors from the current theme.
func (t Theme) Style(fg, bg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(fg).Background(bg)
}
