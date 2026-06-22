package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/usewhale/whale/internal/build"
	tuitheme "github.com/usewhale/whale/internal/tui/theme"
)

func (m model) renderBody(mainWidth, bodyHeight int) string {
	if m.page != pageChat {
		return lipgloss.NewStyle().
			Width(mainWidth).
			Height(bodyHeight).
			Border(lipgloss.NormalBorder()).
			BorderForeground(tuitheme.Default.Border).
			Render(m.viewport.View())
	}
	return m.renderLiveArea(mainWidth, bodyHeight)
}

func (m model) renderLiveArea(width, bodyHeight int) string {
	lines := m.renderChatLines(max(20, width-2))
	if len(lines) == 0 {
		return ""
	}
	maxLines := max(3, bodyHeight)
	truncated := false
	if len(lines) > maxLines {
		truncated = true
		lines = lines[len(lines)-maxLines:]
	}
	if truncated {
		prefix := lipgloss.NewStyle().
			Foreground(tuitheme.Default.Muted).
			Render("... live output truncated; full turn will be added to scrollback when complete")
		lines = append([]string{prefix}, lines...)
	}
	return lipgloss.NewStyle().
		Width(width).
		Render(strings.TrimRight(strings.Join(lines, "\n"), "\n"))
}

func (m model) View() string {
	mainWidth, bodyHeight := m.layoutDims()
	m.refreshViewportContent()
	body := m.renderBody(mainWidth, bodyHeight)
	status := lipgloss.NewStyle().Foreground(tuitheme.Default.StatusIdle).Render(m.status)
	if m.busy {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spin := frames[int(time.Now().UnixNano()/int64(120*time.Millisecond))%len(frames)]
		label := "working"
		if m.stopping {
			label = "stopping"
		}
		status = lipgloss.NewStyle().Foreground(tuitheme.Default.Warn).Render(label + " " + spin)
	}
	footerText := "status: " + status + "  model: " + m.model + "  effort: " + m.effort + "  thinking: " + m.thinking
	if m.chatMode == "plan" {
		footerText += "  mode: plan (Shift+Tab to switch)"
	}
	footerText += "  scroll/copy with terminal"
	footer := lipgloss.JoinHorizontal(lipgloss.Left, footerText)
	parts := make([]string, 0, 3)
	if body != "" {
		parts = append(parts, body)
		parts = append(parts, "\n")
	}
	// Prominent input area — thick separator + visible box + cursor indicator
separator := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#00d4ff")).
    Render(strings.Repeat("─", mainWidth))
inputLabel := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#00d4ff")).
    Bold(true).
    Render(" ╭─ input ")
inputStyle := lipgloss.NewStyle().
    Border(lipgloss.NormalBorder()).
    BorderForeground(lipgloss.Color("#00d4ff")).
    Background(lipgloss.Color("#0d0d1a")).
    Padding(0, 1).
    Width(mainWidth)
cursorHint := " " + lipgloss.NewStyle().
    Foreground(lipgloss.Color("#00e660")).
    Render("●")
if !m.busy {
    cursorHint = " " + lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00d4ff")).
        Render("⌨")
}
parts = append(parts, separator)
parts = append(parts, inputLabel + cursorHint)
parts = append(parts, inputStyle.Render(m.input.View()), footer)
	view := strings.Join(parts, "\n")
	if m.mode == modeChat && m.hasSlashSuggestions() {
		view += "\n" + m.renderSlashSuggestions()
	}
	if m.mode == modeApproval {
		opts := []string{"Allow (a)", "Allow for Session (s)", "Deny (d)"}
		for i := range opts {
			if i == m.approval.selected {
				opts[i] = "[" + opts[i] + "]"
			}
		}
		view += "\n\n" + lipgloss.NewStyle().Foreground(tuitheme.Default.Error).Render(
			fmt.Sprintf(
				"approval: %s\nid: %s\n%s\n\n%s\n(←/→/tab select, enter confirm, esc deny)",
				m.approval.toolName,
				m.approval.toolCallID,
				m.approval.reason,
				strings.Join(opts, "   "),
			),
		)
	}
	if m.mode == modePlanImplementation {
		view += "\n\n" + m.renderPlanImplementationPicker()
	}
	if m.mode == modeSessionPicker {
		rows := []string{"sessions (↑/↓ select, enter confirm, esc cancel):"}
		for i, row := range m.sessionChoices {
			if isSessionHeaderRow(row) {
				rows = append(rows, row)
				continue
			}
			prefix := "  "
			if i == m.sessionIndex {
				prefix = "> "
			}
			rows = append(rows, prefix+stripSessionOrdinal(row))
		}
		view += "\n\n" + lipgloss.NewStyle().Foreground(tuitheme.Default.Plan).Render(strings.Join(rows, "\n"))
	}
	if m.mode == modeUserInput {
		if m.userInput.index < len(m.userInput.questions) {
			q := m.userInput.questions[m.userInput.index]
			rows := make([]string, 0, len(q.Options)+3)
			rows = append(rows, q.Question)
			rows = append(rows, "")
			for i, opt := range q.Options {
				prefix := "  "
				if i == m.userInput.selectedOption {
					prefix = "> "
				}
				rows = append(rows, fmt.Sprintf("%s%s - %s", prefix, opt.Label, opt.Description))
			}
			rows = append(rows, "", "(up/down choose, enter confirm, esc cancel)")
			view += "\n\n" + lipgloss.NewStyle().Foreground(tuitheme.Default.Info).Render(strings.Join(rows, "\n"))
		}
	}
	if m.mode == modeModelPicker {
		view += "\n\n" + m.renderModelPicker()
	}
	if m.mode == modePermissionsPicker {
		view += "\n\n" + m.renderPermissionsPicker()
	}
	return view
}

func resolveVersion() string {
	return build.CurrentVersion()
}

func buildHeaderBanner(modelName, effort, cwd, version string) string {
	return fmt.Sprintf("▸ Whale %s   model: %s %s   dir: %s",
		version, modelName, effort, cwd)
}

func resolveWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	home, hErr := os.UserHomeDir()
	if hErr == nil {
		if rel, rErr := filepath.Rel(home, wd); rErr == nil && rel != "" && rel != "." && !strings.HasPrefix(rel, "..") {
			return "~/" + rel
		}
		if filepath.Clean(wd) == filepath.Clean(home) {
			return "~"
		}
	}
	return wd
}

func (m model) pageLabel() string {
	if m.page == pageLogs {
		return "logs"
	}
	if m.page == pageDiff {
		return "diff"
	}
	return "chat"
}

func (m model) renderPalette() string {
	rows := []string{"Command Palette (enter to run, esc to close)"}
	for i, it := range m.palette.actions {
		prefix := "  "
		if i == m.palette.selected {
			prefix = "> "
		}
		rows = append(rows, prefix+it.Label)
	}
	return lipgloss.NewStyle().Foreground(tuitheme.Default.Palette).Render(strings.Join(rows, "\n"))
}

func (m model) renderModelPicker() string {
	rows := []string{"Select Model and Effort"}
	rows = append(rows, "")
	rows = append(rows, "Model:")
	for i, item := range m.modelPicker.models {
		prefix := "  "
		if m.modelPicker.stage == 0 && i == m.modelPicker.modelIx {
			prefix = "> "
		}
		rows = append(rows, prefix+item)
	}
	if m.modelPicker.stage >= 1 {
		rows = append(rows, "")
		rows = append(rows, "Effort:")
		for i, item := range m.modelPicker.efforts {
			prefix := "  "
			if m.modelPicker.stage == 1 && i == m.modelPicker.effIx {
				prefix = "> "
			}
			rows = append(rows, prefix+item)
		}
	}
	if m.modelPicker.stage >= 2 {
		rows = append(rows, "", "Thinking:")
		for i, item := range m.modelPicker.thinkings {
			prefix := "  "
			if m.modelPicker.stage == 2 && i == m.modelPicker.thinkIx {
				prefix = "> "
			}
			rows = append(rows, prefix+item)
		}
	}
	rows = append(rows, "", "(up/down choose, enter next/confirm, esc back)")
	return lipgloss.NewStyle().Foreground(tuitheme.Default.Info).Render(strings.Join(rows, "\n"))
}

func (m model) renderPermissionsPicker() string {
	rows := []string{"Permissions", ""}
	descriptions := map[string]string{
		"Ask first":    "Ask before write, patch, or shell tools run.",
		"Auto approve": "Never ask; auto-approve tool calls.",
	}
	for i, item := range m.permissionsPicker.choices {
		prefix := "  "
		if i == m.permissionsPicker.index {
			prefix = "> "
		}
		if desc := descriptions[item]; desc != "" {
			rows = append(rows, fmt.Sprintf("%s%s - %s", prefix, item, desc))
		} else {
			rows = append(rows, prefix+item)
		}
	}
	rows = append(rows, "", "(up/down choose, enter confirm, esc cancel)")
	return lipgloss.NewStyle().Foreground(tuitheme.Default.Info).Render(strings.Join(rows, "\n"))
}

func (m model) renderPlanImplementationPicker() string {
	rows := []string{"Implement this plan?", ""}
	items := []struct {
		label string
	}{
		{"Yes, implement this plan"},
		{"No, stay in Plan mode"},
	}
	for i, item := range items {
		prefix := "  "
		if i == m.planImplementation.index {
			prefix = "> "
		}
		rows = append(rows, prefix+item.label)
	}
	rows = append(rows, "", "(up/down choose, enter confirm, esc cancel)")
	return lipgloss.NewStyle().Foreground(tuitheme.Default.Info).Render(strings.Join(rows, "\n"))
}

func (m model) layoutDims() (mainWidth, bodyHeight int) {
	bodyHeight = max(3, m.height-6)
	mainWidth = m.width
	if m.sidebar && m.width > 80 {
		mainWidth = int(float64(m.width) * 0.72)
	}
	return mainWidth, bodyHeight
}

func (m model) chatRenderWidth() int {
}


// renderDyadMessage renders a chat message with full bubble styling.
// Like a messaging app — sender, bubble, timestamp.
func renderDyadMessage(sender, content string, isHuman bool) string {
	senderColor := "#00d4ff"  // CoCreator (cyan)
	senderLabel := "🐋"
	prefix := "╰─"
	bubbleBg := "#0a0a1a"
	bubbleBorder := "#00d4ff"
	
	if isHuman {
		senderColor = "#b44dff"  // Peter (purple)
		senderLabel = "👤"
		prefix = "╭─"
		bubbleBg = "#140a1a"
		bubbleBorder = "#b44dff"
	}
	
	// Sender header
	timeStr := time.Now().Format("15:04")
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(senderColor)).
		Bold(true).
		Render(fmt.Sprintf("%s %s %s  %s", prefix, senderLabel, sender, timeStr))
	
	// Code detection — wrap code blocks with syntax hint
	content = renderCodeBlocks(content)
	
	// Message bubble
	bubble := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(bubbleBorder)).
		Background(lipgloss.Color(bubbleBg)).
		Padding(0, 1).
		MaxWidth(100).
		Render(content)
	
	return header + "
" + bubble
}

// renderCodeBlocks wraps code blocks with visible markers.
func renderCodeBlocks(content string) string {
	if strings.Contains(content, "```") {
		// Already has markdown code fences — leave as-is
		return content
	}
	return content
}

// renderTypingIndicator shows when the dyad is generating.
func renderTypingIndicator() string {
	frames := []string{"●", "●", "●", "●", "●"}
	// Simple fixed indicator
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00d4ff")).
		Render("🐋  ● ● ●")
}

// renderAGUIWidget renders an ad-hoc AG-UI element inline.
func renderAGUIWidget(widgetType string, data map[string]string) string {
	switch widgetType {
	case "progress":
		pct := 50.0
		if v, ok := data["percent"]; ok {
			fmt.Sscanf(v, "%f", &pct)
		}
		barLen := 20
		filled := int(pct / 100.0 * float64(barLen))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
		label := data["label"]
		return fmt.Sprintf(" %s [%s] %.0f%%", label, bar, pct)
		
	case "status":
		label := data["label"]
		value := data["value"]
		color := dataGet(data, "color", "#00e660")
		return fmt.Sprintf(" %s: %s", label, lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(value))
		
	case "card":
		title := data["title"]
		body := data["body"]
		w := 40
		if strings.Contains(body, "
") {
			return lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#00d4ff")).
				Padding(0, 1).
				Width(w).
				Render(title + "
" + body)
		}
		return fmt.Sprintf(" ┌ %s ┐
 │ %s │
 └─────┘", title, body)
		
	default:
		return ""
	}
}

// Helper for data.get with default
func dataGet(d map[string]string, key, def string) string {
	if v, ok := d[key]; ok {
		return v
	}
	return def
}
