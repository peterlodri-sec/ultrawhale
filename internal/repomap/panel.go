package repomap

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PanelStyle defines colors for the repo map TUI panel.
type PanelStyle struct {
	Title    lipgloss.Style
	File     lipgloss.Style
	Func     lipgloss.Style
	Type     lipgloss.Style
	Import   lipgloss.Style
	Exported lipgloss.Style
	Edge     lipgloss.Style
	Dimmed   lipgloss.Style
}

// DefaultPanelStyle returns the default panel styling.
func DefaultPanelStyle() PanelStyle {
	return PanelStyle{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		File:     lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		Func:     lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		Type:     lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		Import:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Exported: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		Edge:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Dimmed:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

// PanelModel is a Bubble Tea model for the repo map sidebar.
type PanelModel struct {
	graph    *Graph
	viewport viewport.Model
	style    PanelStyle
	width    int
	height   int
	focused  bool
}

// NewPanel creates a repo map panel model.
func NewPanel(g *Graph, width, height int) PanelModel {
	vp := viewport.New(width, height)
	vp.SetContent(renderPanel(g, width, DefaultPanelStyle()))

	return PanelModel{
		graph:    g,
		viewport: vp,
		style:    DefaultPanelStyle(),
		width:    width,
		height:   height,
	}
}

// Init implements tea.Model.
func (m PanelModel) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m PanelModel) Update(msg tea.Msg) (PanelModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the panel.
func (m PanelModel) View() string {
	content := renderPanel(m.graph, m.width, m.style)
	m.viewport.SetContent(content)
	return m.viewport.View()
}

// SetSize updates the panel dimensions.
func (m *PanelModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
}

// UpdateGraph replaces the graph data and refreshes.
func (m *PanelModel) UpdateGraph(g *Graph) {
	m.graph = g
}

// Focused returns whether the panel has keyboard focus.
func (m PanelModel) Focused() bool { return m.focused }

// Focus sets keyboard focus.
func (m *PanelModel) Focus()  { m.focused = true }
func (m *PanelModel) Blur()   { m.focused = false }

// ── Render ────────────────────────────────────────────────────────────────

func renderPanel(g *Graph, width int, style PanelStyle) string {
	if g == nil || len(g.Files) == 0 {
		return style.Dimmed.Render("(no symbols yet — building...)")
	}

	var sb strings.Builder

	// Header
	sb.WriteString(style.Title.Render("Repo Map"))
	sb.WriteString(style.Dimmed.Render(fmt.Sprintf("  %d files · %d symbols · %d edges\n",
		len(g.Files), len(g.Symbols), len(g.Edges))))
	sb.WriteString(strings.Repeat("─", width) + "\n")

	// Files and their symbols
	files := sortedFiles(g)
	for _, file := range files {
		syms := g.SymbolsInFile(file)
		if len(syms) == 0 {
			continue
		}

		// File header
		sb.WriteString(style.File.Render(file))
		deps := g.DependentsOf(file)
		if len(deps) > 0 {
			sb.WriteString(style.Edge.Render(fmt.Sprintf("  ← %d dependents", len(deps))))
		}
		sb.WriteString("\n")

		// Symbols
		for _, sym := range syms {
			prefix := "  "
			switch sym.Kind {
			case "func", "method":
				sb.WriteString(prefix + style.Func.Render(sym.Name+"()"))
			case "type", "class", "struct", "enum":
				sb.WriteString(prefix + style.Type.Render(sym.Name))
			case "import":
				sb.WriteString(prefix + style.Import.Render("← "+sym.Name))
			default:
				sb.WriteString(prefix + style.Dimmed.Render(sym.Name))
			}

			// Export marker
			if sym.Exported {
				sb.WriteString(style.Exported.Render(" ●"))
			}

			// Line number
			sb.WriteString(style.Dimmed.Render(fmt.Sprintf(" :%d", sym.Line)))

			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func sortedFiles(g *Graph) []string {
	files := make([]string, len(g.Files))
	copy(files, g.Files)
	sort.Slice(files, func(i, j int) bool {
		// Sort by dependency count (most depended-on files first)
		di := len(g.DependentsOf(files[i]))
		dj := len(g.DependentsOf(files[j]))
		if di != dj {
			return di > dj
		}
		return files[i] < files[j]
	})
	return files
}
