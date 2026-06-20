package tui

import "github.com/charmbracelet/lipgloss"

func (m *model) renderSidebar(body string) string {
	if m.orchPanel == nil || !m.orchPanel.Visible {
		return body
	}
	panel := m.orchPanel.View()
	if panel == "" {
		return body
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, body, "  "+panel)
}
