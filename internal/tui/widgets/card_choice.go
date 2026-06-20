package widgets

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CardChoice renders a card-based selection UI for user choices.
// Used for approval prompts, tool selection, mode switching.
type CardChoiceWidget struct {
	Base
	Visible  bool
	Title    string
	Options  []CardOption
	Selected int
}

type CardOption struct {
	Label       string
	Description string
	Key         string // keyboard shortcut
	Action      string // what happens on select
}

func NewCardChoice(title string, options []CardOption) *CardChoiceWidget {
	return &CardChoiceWidget{
		Base:     NewBase(60, len(options)+3),
		Visible:  true,
		Title:    title,
		Options:  options,
		Selected: 0,
	}
}

func (c *CardChoiceWidget) Init() tea.Cmd { return nil }

func (c *CardChoiceWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.Selected > 0 { c.Selected-- }
		case "down", "j":
			if c.Selected < len(c.Options)-1 { c.Selected++ }
		case "enter":
			if c.Selected < len(c.Options) {
				return c, func() tea.Msg { return cardSelectedMsg{c.Options[c.Selected].Action} }
			}
		}
		// Check key shortcuts
		for i, opt := range c.Options {
			if strings.EqualFold(msg.String(), opt.Key) {
				c.Selected = i
				return c, func() tea.Msg { return cardSelectedMsg{opt.Action} }
			}
		}
	}
	return c, nil
}

type cardSelectedMsg struct{ Action string }

func (c *CardChoiceWidget) View() string {
	if !c.Visible { return "" }
	t := c.Theme
	var cards []string

	// Title
	cards = append(cards, lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("? "+c.Title))

	for i, opt := range c.Options {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(t.Dim)
		if i == c.Selected {
			prefix = "▸ "
			style = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
		}

		keyHint := ""
		if opt.Key != "" {
			keyHint = fmt.Sprintf(" [%s]", strings.ToUpper(opt.Key))
		}

		card := style.Render(fmt.Sprintf("%s%s%s", prefix, opt.Label, keyHint))
		if opt.Description != "" {
			card += "\n   " + lipgloss.NewStyle().Foreground(t.Dim).Render(opt.Description)
		}
		cards = append(cards, card)
	}

	// Footer
	cards = append(cards, lipgloss.NewStyle().Foreground(t.Dim).Render("  ↑↓ select · enter confirm · esc cancel"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Background(t.Bg).Foreground(t.Fg).
		Width(c.Width).
		Padding(1, 2).
		Render(strings.Join(cards, "\n"))
}
