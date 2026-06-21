package blocks

import (
	"fmt"
	"strings"
)

// ── UI Co-Creative — Responsible for Everything UI ───────────────────
//
// The UI Co-Creative is the single point of responsibility for ALL UI.
// Every surface, every widget, every render, every format, every theme.
//
// It coordinates:
//   - VakedDashboard (7-layer status)
//   - Render Engine (markdown, gsm, diff, json, csv)
//   - Display Engine (keyboard→screen pipeline)
//   - AG-UI themes (dense, cyberpunk, graveyard)
//   - A2UI events (toast, layer_update, chat_block)
//   - Native Text (ASM-accelerated rendering)
//   - UI Engine (TUI + Surface + AG-UI + VFS surfaces)
//
// One engine to rule them all. One engine to bind them.

// UICoCreative is the master UI coordinator.
type UICoCreative struct {
	Name      string
	Version   string
	Theme     string
	Surfaces  []string
	Stats     UICoCreativeStats
}

// UICoCreativeStats tracks all UI activity.
type UICoCreativeStats struct {
	Renders       int64
	Formats       int64
	A2UIEvents    int64
	DashboardRefreshes int64
	SurfacesActive    int64
}

var uiCoCreative = &UICoCreative{
	Name:     "ui-co-creative",
	Version:  CurrentVersion(),
	Theme:    "dense",
	Surfaces: []string{"tui", "surface", "agui", "vfs", "web", "voice", "ar"},
}

// ── UI Co-Creative Operations ────────────────────────────────────────

// UICoCreate is the single entry point for ALL UI rendering.
// Every surface, every format, every event flows through here.
func UICoCreate(target, format, content string) string {
	uiCoCreative.Stats.Renders++

	// Route to correct surface
	switch target {
	case "tui":
		return coCreateTUI(format, content)
	case "surface":
		return coCreateSurface(format, content)
	case "agui":
		return coCreateAGUI(format, content)
	case "vfs":
		return coCreateVFS(content)
	case "web":
		return coCreateWeb(format, content)
	default:
		return fmt.Sprintf("[ui-co-creative] unknown target: %s", target)
	}
}

func coCreateTUI(format, content string) string {
	uiCoCreative.Stats.Formats++
	switch format {
	case "vaked-pipeline":
		return renderVakedPipelineASCII()
	case "dashboard":
		uiCoCreative.Stats.DashboardRefreshes++
		return "╔══ Vaked Layers ══╗\n" + allLayerStatus()
	case "sacred":
		return SacredStatus()
	default:
		return RenderFormat(content, format)
	}
}

func coCreateSurface(format, content string) string {
	uiCoCreative.Stats.SurfacesActive++
	return fmt.Sprintf(`{"surface":"%s","format":"%s","content":"%s"}`,
		uiCoCreative.Name, format, content[:min(80, len(content))])
}

func coCreateAGUI(format, content string) string {
	return fmt.Sprintf("[AG-UI ▸ %s] %s", uiCoCreative.Theme,
		RenderFormat(content, format))
}

func coCreateVFS(content string) string {
	VFSEcho("/ultrawhale/ui-co-creative/output", content)
	return "vfs: co-creative output written"
}

func coCreateWeb(format, content string) string {
	return fmt.Sprintf(`<div class="ui-co-creative" data-theme="%s"><pre>%s</pre></div>`,
		uiCoCreative.Theme, RenderFormat(content, format))
}

// ── All Layer Status ─────────────────────────────────────────────────

func allLayerStatus() string {
	layers := []struct {
		Name   string
		Status string
		Icon   string
	}{
		{"Declares", SchemaStatus(), "📜"},
		{"Materializes", NixStatus(), "🏗️"},
		{"Supervises", GetOrchestrator().OrchestratorStatus(), "🔄"},
		{"Enforces", SacredStatus(), "🛡️"},
		{"Testifies", ProbeStatus(), "🔍"},
		{"Indexes", SpaceStatus(), "🗂️"},
		{"Reveals", UIStatus(), "👁️"},
	}

	var lines []string
	for _, l := range layers {
		lines = append(lines, fmt.Sprintf("  %s %s: %s", l.Icon, l.Name, l.Status))
	}
	return strings.Join(lines, "\n")
}

func renderVakedPipelineASCII() string {
	return `╔══════════════════════════════════════════════════════════════════════════╗
║                          VAKED PIPELINE v52                                ║
╠════════════════════════════════════════════════════════════════════════════╣
║ ┌──────────┐   ┌────────┐   ┌──────────┐   ┌────────┐   ┌────────┐   ┌────────┐   ┌────────┐ ║
║ │ DECLARE  │──→│ ENGINE │──→│SUPERVISE │──→│ENFORCE │──→│TESTIFY │──→│ INDEX  │──→│REVEAL  │ ║
║ └──────────┘   └────────┘   └──────────┘   └────────┘   └────────┘   └────────┘   └────────┘ ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝`
}

// ── UI Co-Creative Status ────────────────────────────────────────────

// UICoCreativeStatus returns compact status.
func UICoCreativeStatus() string {
	return fmt.Sprintf("ui-co-creative: %d renders · %d formats · %d a2ui · %d dashboards · %d surfaces active",
		uiCoCreative.Stats.Renders, uiCoCreative.Stats.Formats,
		uiCoCreative.Stats.A2UIEvents, uiCoCreative.Stats.DashboardRefreshes,
		uiCoCreative.Stats.SurfacesActive)
}

// UICoCreativeVakedFit returns Vaked fit.
func UICoCreativeVakedFit() string {
	return `UI CO-CREATIVE = ALL REVEALS, ONE ENGINE

  One engine to rule them all:
    VakedDashboard · Render Engine · Display Engine
    AG-UI themes · A2UI events · Native Text
    UI Engine (TUI + Surface + Web + VFS)

  UICoCreate(target, format, content) → renders everywhere.
  Single entry point for ALL UI. Single responsibility.

  "One engine to rule them all. One engine to bind them."
  — UI Co-Creative, v52`
}
