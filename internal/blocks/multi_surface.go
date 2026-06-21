package blocks

import "fmt"

// ── Multi-Surface — Beyond TUI ───────────────────────────────────────
// v50 timeline: voice, web dashboard, AR

type MultiSurface struct {
	Surfaces []string
	Active   string
}

var multiSurface = &MultiSurface{
	Surfaces: []string{"tui", "web", "voice", "ar"},
	Active:   "tui",
}

func ActivateSurface(surface string) string {
	for _, s := range multiSurface.Surfaces {
		if s == surface {
			multiSurface.Active = surface
			Log(LogInfo, "surface.activate", surface, "", "", 0, nil)
			return fmt.Sprintf("surface: %s activated", surface)
		}
	}
	return fmt.Sprintf("surface: %s not found (available: %v)", surface, multiSurface.Surfaces)
}

func MultiSurfaceStatus() string {
	return fmt.Sprintf("multi-surface: active=%s · available=%v",
		multiSurface.Active, multiSurface.Surfaces)
}

func MultiSurfaceVakedFit() string {
	return `MULTI-SURFACE = REVEALS LAYER EXPANDED

  v45: TUI + Web dashboard
  v48: Voice interface (speech→text→command)
  v50: AR/VR capability graph visualization`
}
