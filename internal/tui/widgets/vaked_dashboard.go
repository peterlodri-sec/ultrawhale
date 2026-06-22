package widgets

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/usewhale/whale/internal/blocks"
	"github.com/usewhale/whale/internal/tui/agui"
)

// VakedDashboardWidget shows all 7 Vaked layers with live status.
type VakedDashboardWidget struct {
	Base
	Visible bool
	Width   int
}

func NewVakedDashboard() *VakedDashboardWidget {
	return &VakedDashboardWidget{
		Base:    NewBase(40, 12),
		Visible: true,
		Width:   40,
	}
}

func (v *VakedDashboardWidget) Init() agui.Cmd { return nil }
func (v *VakedDashboardWidget) Update(msg agui.Msg) (agui.Model, agui.Cmd) {
	if msg, ok := msg.(agui.WindowSizeMsg); ok {
		v.Visible = msg.Width >= 80
		v.Width = msg.Width / 4
		if v.Width > 50 { v.Width = 50 }
	}
	return v, nil
}

func (v *VakedDashboardWidget) View() string {
	if !v.Visible { return "" }
	t := v.Theme

	layers := []struct {
		Name   string
		Status string
		Icon   string
	}{
		{"Declares", blocks.SchemaStatus(), "📜"},
		{"Materializes", blocks.NixStatus(), "🏗️"},
		{"Supervises", blocks.GetOrchestrator().OrchestratorStatus(), "🔄"},
		{"Enforces", blocks.SacredStatus(), "🛡️"},
		{"Testifies", blocks.ProbeStatus(), "🔍"},
		{"Indexes", blocks.SpaceStatus(), "🗂️"},
		{"Reveals", blocks.UIStatus(), "👁️"},
	}

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("╔══ Vaked Layers ══╗"))

	for _, l := range layers {
		_ = "✅"
		color := t.Accent
		if strings.Contains(l.Status, "0") && strings.Contains(l.Status, "nodes") {
			icon = "🟡"; color = lipgloss.Color("#ffaa00")
		}
		line := fmt.Sprintf("║ %s %s %s", l.Icon, l.Name, l.Status[:minLen(30, len(l.Status))])
		lines = append(lines, lipgloss.NewStyle().Foreground(color).Render(line))
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(t.Dim).Render("╚══════════════════╝"))

	return lipgloss.NewStyle().
		Background(t.Bg).Foreground(t.Fg).
		Width(v.Width).
		Render(strings.Join(lines, "\n"))
}

func minLen(a, b int) int { if a < b { return a }; return b }
