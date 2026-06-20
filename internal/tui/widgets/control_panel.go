package widgets

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ControlPanel is a floating, fixed-position ASCII mini-dashboard.
// Shows AgentField status, recent activity, and phase summary.
// Auto-dismisses when terminal is too small (<80 cols or <20 rows).
type ControlPanelWidget struct {
	Base
	Visible  bool
	X, Y     int // fixed position (0-indexed from top-left)
	MaxW     int
	MaxH     int
	Status   ControlPanelStatus
}

type ControlPanelStatus struct {
	Agent     string
	Version   string
	DID       string
	Supabase  string
	Phase     string // "ultracode: ✓→●→·→·→·→·→·"
	Uptime    time.Duration
	ToolCalls int
}

func NewControlPanel(x, y int) *ControlPanelWidget {
	return &ControlPanelWidget{
		Base:    NewBase(30, 8),
		X:       x,
		Y:       y,
		Visible: true,
		Status: ControlPanelStatus{
			Agent:   "ultrawhale",
			Version: "v1.9.0",
		},
	}
}

func (c *ControlPanelWidget) Init() tea.Cmd { return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
	return controlPanelTickMsg(t)
})}

type controlPanelTickMsg time.Time

func (c *ControlPanelWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case controlPanelTickMsg:
		c.Status.Uptime += 2 * time.Second
		return c, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return controlPanelTickMsg(t)
		})
	case tea.WindowSizeMsg:
		c.MaxW = msg.Width
		c.MaxH = msg.Height
		c.Visible = msg.Width >= 80 && msg.Height >= 20
		// Reposition to bottom-right
		c.X = msg.Width - c.Width - 2
		c.Y = msg.Height - c.Height - 3
	}
	return c, nil
}

func (c *ControlPanelWidget) View() string {
	if !c.Visible {
		return ""
	}
	t := c.Theme

	// Build panel content
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render(" ⚡ AgentField"))
	lines = append(lines, fmt.Sprintf(" %s %s", c.Status.Agent, c.Status.Version))
	if c.Status.DID != "" {
		lines = append(lines, fmt.Sprintf(" DID: %s...", c.Status.DID[:40]))
	}
	if c.Status.Supabase != "" {
		lines = append(lines, fmt.Sprintf(" DB: %s", c.Status.Supabase))
	}
	if c.Status.Phase != "" {
		lines = append(lines, fmt.Sprintf(" %s", c.Status.Phase))
	}
	lines = append(lines, fmt.Sprintf(" uptime: %s | tools: %d", c.Status.Uptime.Round(time.Second), c.Status.ToolCalls))

	content := strings.Join(lines, "\n")

	// Render with AG-UI border
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Background(t.Bg).
		Foreground(t.Fg).
		Width(c.Width).
		Padding(0, 1).
		Render(content)

	// Position: absolute via lipgloss.Place
	return lipgloss.Place(c.X, c.Y, lipgloss.Right, lipgloss.Bottom, panel)
}
