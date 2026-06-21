package blocks

import (
	"fmt"
	"strings"
)

// ── UI-Engine — The Reveals Layer as an Engine ───────────────────────
//
// If the blocks engine IS the Materializes layer,
// then the UI-engine IS the Reveals layer.
//
// UI-Engine takes engine output and renders it to surfaces:
//   TUI (Bubble Tea) · Surface (HTTP/WS) · AG-UI (themed blocks)
//
// Vaked: Declares → Engine → Materializes → UI-ENGINE → Reveals
//                                                  ↑
//                                           renders to all surfaces

// UIEngine renders block output to all surfaces.
type UIEngine struct {
	Name      string
	Version   string
	Surfaces  []string // "tui", "surface", "agui", "vfs", "blog"
	Theme     string   // "dense", "cyberpunk", "graveyard"
	Stats     UIEngineStats
}

// UIEngineStats tracks UI rendering activity.
type UIEngineStats struct {
	Renders       int64
	TUIUpdates    int64
	APIRequests   int64
	A2UIEvents    int64
	BlocksRendered int64
}

var uiEngine = &UIEngine{
	Name:     "ultrawhale-ui-engine",
	Version:  CurrentVersion(),
	Surfaces: []string{"tui", "surface", "agui", "vfs"},
	Theme:    "dense",
}

// ── UI-Engine Render Pipeline ─────────────────────────────────────────

// UIRender renders engine output to the appropriate surface.
func UIRender(target string, content string) string {
	uiEngine.Stats.Renders++

	switch target {
	case "tui":
		uiEngine.Stats.TUIUpdates++
		return renderToTUI(content)
	case "surface":
		uiEngine.Stats.APIRequests++
		return renderToSurface(content)
	case "agui":
		uiEngine.Stats.BlocksRendered++
		return renderToAGUI(content)
	case "vfs":
		return renderToVFS(content)
	default:
		return fmt.Sprintf("ui-engine: unknown surface '%s'", target)
	}
}

func renderToTUI(content string) string {
	return fmt.Sprintf("[TUI] %s", content[:min(80, len(content))])
}

func renderToSurface(content string) string {
	return fmt.Sprintf(`{"surface": "%s"}`, content[:min(80, len(content))])
}

func renderToAGUI(content string) string {
	return fmt.Sprintf("[AG-UI ▸ %s]: %s", uiEngine.Theme, content[:min(60, len(content))])
}

func renderToVFS(content string) string {
	// VFS: echo to /ultrawhale/ui-engine/output
	VFSEcho("/ultrawhale/ui-engine/output", content)
	return "vfs: written"
}

// ── UI-Engine Surface Management ──────────────────────────────────────

// UIEngineStatus returns compact UI-Engine status.
func UIEngineStatus() string {
	return fmt.Sprintf("ui-engine: %s · %d surfaces (%s) · %d renders (%d tui, %d api, %d agui)",
		uiEngine.Name, len(uiEngine.Surfaces), strings.Join(uiEngine.Surfaces, ", "),
		uiEngine.Stats.Renders, uiEngine.Stats.TUIUpdates,
		uiEngine.Stats.APIRequests, uiEngine.Stats.BlocksRendered)
}

// UIEngineVakedFit returns how the UI-Engine fits into Vaked.
func UIEngineVakedFit() string {
	return `Vaked:  Declares → Engine → Materializes → UI-ENGINE → Reveals
                                                    ↑
                                             ultrawhale-ui-engine
                                          (TUI + Surface + AG-UI + VFS)
              
The UI-Engine IS the Reveals layer.
It takes engine output and renders it to all surfaces:
  TUI (Bubble Tea terminal)
  Surface (HTTP REST + WebSocket)
  AG-UI (themed chat blocks, cards, shaders)
  VFS (virtual filesystem navigation)`
}
