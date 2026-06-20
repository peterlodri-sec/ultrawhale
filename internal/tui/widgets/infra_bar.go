// Package widgets — InfraBar is the always-visible top-of-screen service status bar.
// 5 core services + 1 perf stat. Auto-collapses on small terminals.
// Double-tap Ctrl+Shift+I toggles expanded dashboard.
package widgets

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/usewhale/whale/internal/blocks"
)

// InfraBarWidget is the always-on top bar.
type InfraBarWidget struct {
	Base
	Visible    bool
	Expanded   bool
	Services   []ServiceStatus
	PerfStat   string // most descriptive perf stat from Current
	lastProbe  time.Time
	probeMu    sync.Mutex
}

// ServiceStatus is a single internal service with live health.
type ServiceStatus struct {
	Name   string        // "agentfield", "supabase", "nats", "langfuse", "bao"
	Port   string        // "8585" or URL "langfuse.crabcc.app"
	URL    string        // full probe URL
	Alive  bool
	Detail string        // extra info
	Color  lipgloss.Color
}

func NewInfraBar() *InfraBarWidget {
	ib := &InfraBarWidget{
		Base:    NewBase(120, 1),
		Visible: true,
		Services: []ServiceStatus{
			{Name: "af", Port: "8585", URL: "http://localhost:8585/health"},
			{Name: "db", Port: "8586", URL: "http://localhost:8586/"},
			{Name: "nats", Port: "4222", URL: "nats://crabcc-nats:4222"},
			{Name: "lf", Port: "443", URL: "https://langfuse.crabcc.app/api/public/health"},
			{Name: "bao", Port: "443", URL: "https://bao.crabcc.app/v1/sys/health"},
		},
	}
	return ib
}

func (i *InfraBarWidget) Init() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return infraTickMsg(t)
	})
}

type infraTickMsg time.Time

func (i *InfraBarWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case infraTickMsg:
		go i.probeAll()
		i.updatePerfStat()
		return i, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return infraTickMsg(t)
		})
	case tea.WindowSizeMsg:
		i.Width = msg.Width
		i.Height = msg.Height
		i.Visible = msg.Width >= 60
	}
	return i, nil
}

// ── Health probes ──────────────────────────────────────────────────────

func (i *InfraBarWidget) probeAll() {
	i.probeMu.Lock()
	defer i.probeMu.Unlock()
	i.lastProbe = time.Now()

	for idx := range i.Services {
		svc := &i.Services[idx]
		svc.Alive = i.probe(svc)
		switch svc.Name {
		case "af": svc.Detail = "agentfield"
		case "db": svc.Detail = "supabase"
		case "nats": svc.Detail = "eventbus"
		case "lf": svc.Detail = "langfuse"
		case "bao": svc.Detail = "vault"
		}
	}
}

func (i *InfraBarWidget) probe(svc *ServiceStatus) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}

	switch {
	case strings.HasPrefix(svc.URL, "nats://"):
		addr := strings.TrimPrefix(strings.TrimPrefix(svc.URL, "nats://"), "tls://")
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err != nil { return false }
		conn.Close()
		return true

	case strings.HasPrefix(svc.URL, "https://"):
		resp, err := client.Get(svc.URL)
		if err != nil { return false }
		resp.Body.Close()
		return resp.StatusCode < 400

	case strings.HasPrefix(svc.URL, "http://"):
		resp, err := client.Get(svc.URL)
		if err != nil { return false }
		resp.Body.Close()
		return resp.StatusCode < 400

	default:
		return false
	}
}

func (i *InfraBarWidget) updatePerfStat() {
	c := blocks.GetCurrent()
	if c.Busy {
		i.PerfStat = fmt.Sprintf("%.0f/s", c.TokensPerSec)
	} else {
		i.PerfStat = fmt.Sprintf("%dMB", c.MemoryMB)
	}
}

// ── View ───────────────────────────────────────────────────────────────

func (i *InfraBarWidget) View() string {
	if !i.Visible { return "" }
	if i.Expanded { return i.renderDashboard() }
	return i.renderBar()
}

func (i *InfraBarWidget) renderBar() string {
	t := i.Theme
	bg := lipgloss.NewStyle().Background(t.Bg).Foreground(t.Fg)

	// Left: ultrawhale branding
	metalInfo := blocks.MetalStatus()
	if strings.Contains(metalInfo, "available") {
		metalInfo = "gpu"
	} else {
		metalInfo = ""
	}
	left := lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("▸ultrawhale")
	// Shell nesting indicator
	if ShellActive {
		left += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00")).Render("[sh]")
	}

	// Center: service indicators
	var services []string
	i.probeMu.Lock()
	for _, svc := range i.Services {
		color := i.goodColor()
		if !svc.Alive { color = i.dimColor() }
		services = append(services, lipgloss.NewStyle().Foreground(color).Render(
			fmt.Sprintf("%s:%s", svc.Name, svc.Port)))
	}
	i.probeMu.Unlock()

	// Right: perf stat + uptime
	right := lipgloss.NewStyle().Foreground(i.dimColor()).Render(
		i.PerfStat + " · " + fmt.Sprintf("%d tok", blocks.GetCurrent().TotalTokens))

	// Assemble: left · services · right
	center := strings.Join(services, " │ ")
	full := fmt.Sprintf(" %s │ %s │ %s ", left, center, right)

	if lipgloss.Width(full) > i.Width {
		// Collapse: just ultrawhale + service count + perf
		alive := 0
		for _, svc := range i.Services {
			if svc.Alive { alive++ }
		}
		full = fmt.Sprintf(" ▸ultrawhale │ %d/%d services │ %s ", alive, len(i.Services), i.PerfStat)
	}

	return bg.Width(i.Width).MaxWidth(i.Width).Render(full)
}

func (i *InfraBarWidget) renderDashboard() string {
	t := i.Theme
	var lines []string

	// Header
	lines = append(lines, lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Render("╔══ Infra Dashboard ══╗"))
	lines = append(lines, fmt.Sprintf("║ ultrawhale %s · uptime: %s", blocks.GetSessionSelf().Version, "now"))

	// Service details
	i.probeMu.Lock()
	for _, svc := range i.Services {
		icon := "✓"
		color := i.goodColor()
		if !svc.Alive { icon = "✗"; color = i.dimColor() }
		lines = append(lines, lipgloss.NewStyle().Foreground(color).Render(
			fmt.Sprintf("║ %s %s: %s (%s)", icon, svc.Name, svc.Detail, svc.URL)))
	}
	i.probeMu.Unlock()

	// Perf
	c := blocks.GetCurrent()
	lines = append(lines, fmt.Sprintf("║ tier: %s · %dt · %.0f/s · cache:%.0f%% · %dMB · $%.4f",
		c.Tier, c.TotalTokens, c.TokensPerSec, c.CacheHitPct, c.MemoryMB, c.CostUSD))

	lines = append(lines, lipgloss.NewStyle().Foreground(t.Dim).Render("╚══════════════════╝"))

	return lipgloss.NewStyle().Background(t.Bg).Foreground(t.Fg).Width(i.Width).Render(
		strings.Join(lines, "\n"))
}

func (i *InfraBarWidget) goodColor() lipgloss.Color { return lipgloss.Color("#00e660") }
func (i *InfraBarWidget) dimColor() lipgloss.Color  { return lipgloss.Color("#6878a0") }

// ToggleExpanded switches between bar and dashboard.
func (i *InfraBarWidget) ToggleExpanded() {
	i.Expanded = !i.Expanded
	if i.Expanded {
		i.Height = 10
	} else {
		i.Height = 1
	}
}
