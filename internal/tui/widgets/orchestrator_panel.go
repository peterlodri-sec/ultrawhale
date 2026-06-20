package widgets

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/usewhale/whale/internal/blocks"
)

// OrchestratorPanel is the right-side orchestrator dashboard.
// Shows: active subagents, brain status, orchestrator identity.
type OrchestratorPanelWidget struct {
	Base
	Visible bool
	Width   int
}

func NewOrchestratorPanel() *OrchestratorPanelWidget {
	return &OrchestratorPanelWidget{
		Base:    NewBase(30, 20),
		Visible: true,
		Width:   30,
	}
}

func (o *OrchestratorPanelWidget) Init() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return orchTickMsg(t)
	})
}

type orchTickMsg time.Time

func (o *OrchestratorPanelWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case orchTickMsg:
		return o, tea.Tick(3*time.Second, func(t time.Time) tea.Msg { return orchTickMsg(t) })
	case tea.WindowSizeMsg:
		if msg.Width < 100 { o.Visible = false } else { o.Visible = true }
	}
	return o, nil
}

func (o *OrchestratorPanelWidget) View() string {
	if !o.Visible { return "" }

	t := o.Theme
	orch := blocks.GetOrchestrator()

	var lines []string

	// Header
	lines = append(lines, lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("┌─ Orchestrator ─┐"))
	lines = append(lines, fmt.Sprintf("│ %s", orch.DID[:30]))
	lines = append(lines, fmt.Sprintf("│ turns: %d", orch.TotalTurns))

	// Agents
	agents := blocks.ListAgents()
	lines = append(lines, fmt.Sprintf("│ agents: %d", len(agents)))
	for _, a := range agents {
		icon := "●"
		color := t.Accent
		if a.Status != "running" { icon = "✓"; color = lipgloss.Color("#00e660") }
		lines = append(lines, lipgloss.NewStyle().Foreground(color).Render(
			fmt.Sprintf("│ %s %s: %s (%d tools)", icon, a.Role, a.Status, a.ToolCalls)))
	}

	// Brain
	brain := blocks.GetBrain()
	memos := brain.BrainDump()
	lines = append(lines, fmt.Sprintf("│ %s", memos))

	// Footer
	lines = append(lines, lipgloss.NewStyle().Foreground(t.Dim).Render("└────────────────┘"))

	return lipgloss.NewStyle().
		Background(t.Bg).Foreground(t.Fg).
		Width(o.Width).
		Render(strings.Join(lines, "\n"))
}
