// Package widgets provides composable, reusable TUI components.
// Each widget implements tea.Model and encapsulates its own state,
// rendering, and update logic. Widgets are composed in View().
package widgets

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/usewhale/whale/internal/tui/agui"
)

// Widget extends tea.Model with sizing and focus management.
type Widget interface {
	tea.Model
	SetSize(width, height int)
	Focused() bool
	Focus()
	Blur()
}

// Base provides common widget fields: size, focus, theme.
type Base struct {
	Width   int
	Height  int
	focused bool
	Theme   agui.Theme
}

func NewBase(width, height int) Base {
	return Base{
		Width:  width,
		Height: height,
		Theme:  agui.Current,
	}
}

func (b *Base) SetSize(width, height int) {
	b.Width = width
	b.Height = height
}

func (b Base) Focused() bool { return b.focused }
func (b *Base) Focus()       { b.focused = true }
func (b *Base) Blur()        { b.focused = false }

// Style returns a lipgloss.Style with the widget theme colors.
func (b Base) Style(fg, bg lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(fg).Background(bg)
}

// Accent returns the theme accent color.
func (b Base) Accent() lipgloss.Color { return b.Theme.Accent }

// Dim returns the theme dim color.
func (b Base) Dim() lipgloss.Color { return b.Theme.Dim }

// Fg returns the theme foreground color.
func (b Base) Fg() lipgloss.Color { return b.Theme.Fg }

// Good returns a success green color.
func (b Base) Good() lipgloss.Color { return lipgloss.Color("#00e660") }

// WarnColor returns a warning yellow color.
func (b Base) WarnColor() lipgloss.Color { return lipgloss.Color("#ffaa00") }
