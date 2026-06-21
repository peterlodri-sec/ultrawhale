package blocks

import (
	"fmt"
	"strings"
)

// ── Surface Atom — The Missing Fundamental Atom ──────────────────────
//
// Space tells WHERE. Surface tells HOW.
// Every block has a Space position. Every block should have a Surface representation.
//
// SurfaceAtom IS the surface. Not the coordinator (ui-engine) — the thing itself.
// Like SpaceNode vs SpaceTopology. SurfaceAtom vs ui-engine.
//
// 3 rounds of CoCreator Q&A:
//   Q: What does "surface as atom" mean?
//   A: Surface = HOW. Space = WHERE. Context = WHAT. Time = WHEN.
//   Q: How does this differ from ui-engine?
//   A: ui-engine coordinates. SurfaceAtom IS. One per block.
//   Q: Minimum implementation?
//   A: SurfaceAtom{Kind, Render, Reveals}. Every block registers one.

// SurfaceAtom is one block's rendering surface.
type SurfaceAtom struct {
	BlockID  string   // which block this surface belongs to
	Kind     string   // "tui", "web", "ssh", "voice", "ar", "api", "rss"
	Content  string   // rendered content
	Reveals  string   // which Vaked layer this reveals
	Visible  bool
	Width    int
	Height   int
}

// SurfaceRegistry maps blocks to their SurfaceAtoms.
type SurfaceRegistry struct {
	Atoms map[string]*SurfaceAtom // blockID → atom
}

var surfaceRegistry = &SurfaceRegistry{Atoms: make(map[string]*SurfaceAtom)}

// RegisterSurfaceAtom registers a surface for a block.
func RegisterSurfaceAtom(blockID, kind, content, reveals string) *SurfaceAtom {
	atom := &SurfaceAtom{
		BlockID: blockID, Kind: kind, Content: content,
		Reveals: reveals, Visible: true,
	}

	surfaceRegistry.Atoms[blockID] = atom
	Log(LogInfo, "surface.atom.register", fmt.Sprintf("%s → %s (%s)", blockID, kind, reveals[:min(20, len(reveals))]),
		"", "", 0, nil)
	return atom
}

// SurfaceAtomStatus returns compact surface atom registry status.
func SurfaceAtomStatus() string {
	counts := map[string]int{}
	for _, a := range surfaceRegistry.Atoms {
		counts[a.Kind]++
	}

	var kinds []string
	for k, v := range counts {
		kinds = append(kinds, fmt.Sprintf("%s:%d", k, v))
	}

	return fmt.Sprintf("surface-atoms: %d blocks (%s)",
		len(surfaceRegistry.Atoms), strings.Join(kinds, ", "))
}

// SurfaceAtomVakedFit returns Surface Atom's Vaked fit.
func SurfaceAtomVakedFit() string {
	return `SURFACE ATOM = THE MISSING FUNDAMENTAL ATOM

  Space: WHERE  (topology, position, distance)
  Surface: HOW  (rendering, appearance, visibility)

  Every block has a Space position.
  Every block should have a Surface representation.

  Like SpaceNode vs SpaceTopology.
  SurfaceAtom vs ui-engine.

  "Surface as {atom} like {space}" — Peter
  3-round Q&A with CoCreator, v82`
}
