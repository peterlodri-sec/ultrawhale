package tui

func (m *model) renderInfraBar() string {
	if m.infraBar == nil || !m.infraBar.Visible {
		return ""
	}
	m.infraBar.Width = m.width
	if m.orchPanel != nil && m.orchPanel.Visible {
		m.infraBar.Width -= 32
	}
	return m.infraBar.View()
}
