package blocks

import (
	"fmt"
	"strings"
)

// ── Render Engine — Native Format Rendering ──────────────────────────
// The 8th engine. Renders structured formats to TUI/Surface.
// Declares → Render → Reveals.

type RenderEngine struct {
	Name    string
	Version string
	Stats   RenderStats
}

type RenderStats struct {
	MarkdownRenders int64
	GSMRenders      int64
	DiffRenders     int64
	JSONRenders     int64
	CSVRenders      int64
}

var renderEngine = &RenderEngine{Name: "render-engine", Version: CurrentVersion()}

// RenderFormat detects and renders a format.
func RenderFormat(content, format string) string {
	switch strings.ToLower(format) {
	case "md", "markdown":
		renderEngine.Stats.MarkdownRenders++
		return RenderMarkdown(content)
	case "gsm", "state-machine":
		renderEngine.Stats.GSMRenders++
		return RenderGSM(content)
	case "diff", "patch":
		renderEngine.Stats.DiffRenders++
		return RenderDiff(content)
	case "json":
		renderEngine.Stats.JSONRenders++
		return RenderJSON(content)
	case "csv", "table":
		renderEngine.Stats.CSVRenders++
		return RenderCSV(content)
	default:
		return content
	}
}

// RenderMarkdown renders markdown to ANSI-styled text.
func RenderMarkdown(md string) string {
	var sb strings.Builder
	for _, line := range strings.Split(md, "\n") {
		switch {
		case strings.HasPrefix(line, "# "):
			sb.WriteString("\033[1;36m" + line + "\033[0m\n") // bold cyan
		case strings.HasPrefix(line, "## "):
			sb.WriteString("\033[1;32m" + line + "\033[0m\n") // bold green
		case strings.HasPrefix(line, "- "):
			sb.WriteString("  • " + line[2:] + "\n")
		case strings.HasPrefix(line, "```"):
			sb.WriteString("\033[7m" + line + "\033[0m\n") // inverted
		default:
			sb.WriteString(line + "\n")
		}
	}
	return sb.String()
}

// RenderGSM renders a Go State Machine as ASCII art.
func RenderGSM(states string) string {
	parts := strings.Split(states, "→")
	var sb strings.Builder
	sb.WriteString("┌──────────┐\n")
	for i, s := range parts {
		s = strings.TrimSpace(s)
		sb.WriteString(fmt.Sprintf("│  %-8s │\n", s))
		if i < len(parts)-1 {
			sb.WriteString("│    ↓     │\n")
		}
	}
	sb.WriteString("└──────────┘")
	return sb.String()
}

// RenderDiff renders a unified diff with ANSI colors.
func RenderDiff(diff string) string {
	var sb strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+++"):
			sb.WriteString("\033[1;32m" + line + "\033[0m\n")
		case strings.HasPrefix(line, "---"):
			sb.WriteString("\033[1;31m" + line + "\033[0m\n")
		case strings.HasPrefix(line, "@@"):
			sb.WriteString("\033[1;36m" + line + "\033[0m\n")
		case strings.HasPrefix(line, "+"):
			sb.WriteString("\033[32m" + line + "\033[0m\n")
		case strings.HasPrefix(line, "-"):
			sb.WriteString("\033[31m" + line + "\033[0m\n")
		default:
			sb.WriteString(line + "\n")
		}
	}
	return sb.String()
}

// RenderJSON renders JSON with indentation.
func RenderJSON(json string) string {
	indent := 0
	var sb strings.Builder
	for _, ch := range json {
		switch ch {
		case '{', '[':
			sb.WriteRune(ch)
			indent++
			sb.WriteString("\n" + strings.Repeat("  ", indent))
		case '}', ']':
			indent--
			sb.WriteString("\n" + strings.Repeat("  ", indent))
			sb.WriteRune(ch)
		case ',':
			sb.WriteRune(ch)
			sb.WriteString("\n" + strings.Repeat("  ", indent))
		case ':':
			sb.WriteString(": ")
		default:
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// RenderCSV renders CSV as an ASCII table.
func RenderCSV(csv string) string {
	rows := strings.Split(csv, "\n")
	if len(rows) == 0 { return "" }
	
	var sb strings.Builder
	for i, row := range rows {
		cols := strings.Split(row, ",")
		sb.WriteString("│")
		for _, col := range cols {
			sb.WriteString(fmt.Sprintf(" %-12s │", strings.TrimSpace(col)))
		}
		sb.WriteString("\n")
		if i == 0 {
			sb.WriteString("├" + strings.Repeat("─", 14*len(cols)) + "┤\n")
		}
	}
	return sb.String()
}

// RenderEngineStatus returns compact render engine status.
func RenderEngineStatus() string {
	return fmt.Sprintf("render: markdown:%d gsm:%d diff:%d json:%d csv:%d",
		renderEngine.Stats.MarkdownRenders, renderEngine.Stats.GSMRenders,
		renderEngine.Stats.DiffRenders, renderEngine.Stats.JSONRenders,
		renderEngine.Stats.CSVRenders)
}

// RenderEngineVakedFit returns Vaked fit.
func RenderEngineVakedFit() string {
	return `RENDER ENGINE = THE 8TH ENGINE

  Declares: format schema (markdown, gsm, diff, json, csv)
  Materializes: RenderFormat() → ANSI/HTML
  Reveals: AG-UI block → styled output

  Formats are declarations.
  Rendering is materialization.
  Display is revelation.`
}
