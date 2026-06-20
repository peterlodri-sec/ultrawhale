package statusline

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type HUD struct {
	Width int
	Model      string
	Version    string
	Mode       string
	Branch     string
	CWD        string
	Busy        bool
	Elapsed     time.Duration
	TokenCount  int
	TokensPerSec float64
	CacheHitPct float64
	MemoryMB    int64
	CostUSD     float64
	Plugins     int
	Theme       string
	InfraBao    bool
	InfraLangfuse bool
	InfraNATS   bool
	InfraGPU    int
	BgColor     lipgloss.Color
	FgColor     lipgloss.Color
	AccentColor lipgloss.Color
	DimColor    lipgloss.Color
	WarnColor   lipgloss.Color
	GoodColor   lipgloss.Color
}

func DefaultHUD(width int) *HUD {
	return &HUD{
		Width:       width,
		BgColor:     lipgloss.Color("#0a0a14"),
		FgColor:     lipgloss.Color("#e0e8f5"),
		AccentColor: lipgloss.Color("#00d4ff"),
		DimColor:    lipgloss.Color("#6878a0"),
		WarnColor:   lipgloss.Color("#ffaa00"),
		GoodColor:   lipgloss.Color("#00e660"),
	}
}

func (h HUD) Render() string {
	if h.Width <= 0 {
		return ""
	}
	left := h.renderLeft()
	center := h.renderCenter()
	right := h.renderRight()
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	centerW := h.Width - leftW - rightW - 4
	if centerW < 10 {
		centerW = 10
	}
	center = lipgloss.NewStyle().Width(centerW).Align(lipgloss.Center).Render(center)
	bar := lipgloss.NewStyle().Background(h.BgColor).Foreground(h.FgColor).Width(h.Width).Padding(0, 1).Render(lipgloss.JoinHorizontal(lipgloss.Left, left, center, right))
	return bar
}

func (h HUD) renderLeft() string {
	var parts []string
	parts = append(parts, lipgloss.NewStyle().Foreground(h.AccentColor).Bold(true).Render(h.Model))
	if h.Version != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.DimColor).Render(h.Version))
	}
	if h.Mode != "" && h.Mode != "agent" {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.BgColor).Background(h.WarnColor).Padding(0, 1).Render(h.Mode))
	}
	if h.Branch != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.DimColor).Render("⎇ "+h.Branch))
	}
	return lipgloss.NewStyle().Foreground(h.FgColor).Background(h.BgColor).Render(strings.Join(parts, " "))
}

func (h HUD) renderCenter() string {
	if !h.Busy {
		return lipgloss.NewStyle().Foreground(h.DimColor).Render("ready")
	}
	var parts []string
	parts = append(parts, lipgloss.NewStyle().Foreground(h.WarnColor).Render("● "+formatDuration(h.Elapsed)))
	if h.TokenCount > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.FgColor).Render(fmt.Sprintf("%dt", h.TokenCount)))
	}
	if h.TokensPerSec > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.GoodColor).Render(fmt.Sprintf("%.0f/s", h.TokensPerSec)))
	}
	if h.CacheHitPct > 0 {
		color := h.GoodColor
		if h.CacheHitPct < 50 { color = h.WarnColor }
		parts = append(parts, lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("⎆%.0f%%", h.CacheHitPct)))
	}
	return lipgloss.NewStyle().Foreground(h.FgColor).Render(strings.Join(parts, " · "))
}

func (h HUD) renderRight() string {
	var parts []string
	// Infra status — compact indicators
	if h.InfraBao || h.InfraLangfuse || h.InfraNATS || h.InfraGPU > 0 {
		var infra []string
		if h.InfraBao {
			infra = append(infra, lipgloss.NewStyle().Foreground(h.GoodColor).Render("bao"))
		}
		if h.InfraLangfuse {
			infra = append(infra, lipgloss.NewStyle().Foreground(h.AccentColor).Render("lf"))
		}
		if h.InfraNATS {
			infra = append(infra, lipgloss.NewStyle().Foreground(h.GoodColor).Render("nats"))
		}
		if h.InfraGPU > 0 {
			infra = append(infra, lipgloss.NewStyle().Foreground(h.AccentColor).Render(fmt.Sprintf("%dGPU", h.InfraGPU)))
		}
		parts = append(parts, strings.Join(infra, "·"))
	}
	if h.MemoryMB > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.DimColor).Render(fmt.Sprintf("%dMB", h.MemoryMB)))
	}
	if h.CostUSD > 0 {
		color := h.DimColor
		if h.CostUSD > 0.50 { color = h.WarnColor }
		parts = append(parts, lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("$%.4f", h.CostUSD)))
	}
	if h.Plugins > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.AccentColor).Render(fmt.Sprintf("⚙%d", h.Plugins)))
	}
	if h.Theme != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(h.DimColor).Render(h.Theme[:2]))
	}
	return lipgloss.NewStyle().Foreground(h.FgColor).Background(h.BgColor).Render(strings.Join(parts, " "))
}

func formatDuration(d time.Duration) string {
	if d < time.Second { return fmt.Sprintf("%dms", d.Milliseconds()) }
	if d < time.Minute { return fmt.Sprintf("%.1fs", d.Seconds()) }
	return fmt.Sprintf("%d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}
