package widgets

import (
	"strings"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastWidget displays ephemeral notification messages. Auto-dismiss after 3s.
type ToastWidget struct {
	Base
	messages []toastMsg
}

type toastMsg struct {
	Text      string
	Added     time.Time
	Level     string // "info", "warn", "error"
}

func NewToast(width int) *ToastWidget {
	t := &ToastWidget{Base: NewBase(width, 1)}
	go t.pruneLoop()
	return t
}

func (t *ToastWidget) Init() tea.Cmd { return nil }

func (t *ToastWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return t, nil }

func (t *ToastWidget) Push(text string) {
	t.messages = append(t.messages, toastMsg{Text: text, Added: time.Now(), Level: "info"})
	if len(t.messages) > 5 {
		t.messages = t.messages[1:]
	}
}

func (t *ToastWidget) View() string {
	if len(t.messages) == 0 {
		return ""
	}
	// Show last message only
	msg := t.messages[len(t.messages)-1]
	age := time.Since(msg.Added)
	if age > 3*time.Second {
		return ""
	}
	opacity := 1.0 - age.Seconds()/3.0
	if opacity < 0 {
		opacity = 0
	}
	color := lipgloss.Color("#00d4ff")
	return lipgloss.NewStyle().
		Foreground(color).
		Width(t.Width).
		Align(lipgloss.Center).
		Render(strings.TrimSpace(msg.Text))
}

func (t *ToastWidget) pruneLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-4 * time.Second)
		var kept []toastMsg
		for _, m := range t.messages {
			if m.Added.After(cutoff) {
				kept = append(kept, m)
			}
		}
		t.messages = kept
	}
}

// SubagentStarted pushes a compact subagent status notification.
func (t *ToastWidget) SubagentStarted(role, task string) {
	preview := task
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	t.Push("🤖 " + role + ": " + preview)
}

// SubagentDone marks the subagent as completed.
func (t *ToastWidget) SubagentDone(role string, durationMs int64) {
	t.Push(fmt.Sprintf("✅ %s done (%dms)", role, durationMs))
}
