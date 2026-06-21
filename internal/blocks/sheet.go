package blocks

import (
	"fmt"
	"strings"
)

// ── SHEET — Sacred Hypertext Element Engine for Terminals ────────────
//
// Like React, but for ASCII. Component-based rendering for the terminal.
// Every SACRED surface element is a SHEET component.
//
// Component = a function that takes Props and returns rendered string.
// Props = key-value pairs (like React props).
// State = component-local state (like React useState).
// Render = the rendered ASCII output.
//
// Name chosen by CoCreator: SHEET.
// "Like a sheet of paper. The SACRED surface. Flat. Visible. Eternal."

// Sheet is a renderable component.
type Sheet struct {
	Name     string
	Props    map[string]any
	Children []*Sheet
	Render   func(props map[string]any, children []string) string
}

// ── SHEET H (createElement) ──────────────────────────────────────────

// H creates a new SHEET element (like React.createElement).
func H(name string, props map[string]any, children ...*Sheet) *Sheet {
	return &Sheet{
		Name:     name,
		Props:    props,
		Children: children,
	}
}

// Text creates a text node (no children, just content).
func Text(content string) *Sheet {
	return &Sheet{
		Name: "text",
		Props: map[string]any{"content": content},
	}
}

// ── Built-in Components ───────────────────────────────────────────────

// Box renders a bordered box (like a div with border).
func Box(props map[string]any, children []string) string {
	title, _ := props["title"].(string)
	width := 52
	if w, ok := props["width"].(int); ok { width = w }

	lines := []string{}
	for _, c := range children {
		lines = append(lines, c)
	}

	return ASCIIBox(title, lines, width)
}

// Row renders a horizontal row of elements.
func Row(props map[string]any, children []string) string {
	gap := "  "
	if g, ok := props["gap"].(string); ok { gap = g }
	return strings.Join(children, gap)
}

// StatusBadge renders a colored status badge.
func StatusBadge(props map[string]any, children []string) string {
	label, _ := props["label"].(string)
	color, _ := props["color"].(string)
	if color == "" { color = "00d4ff" }

	return fmt.Sprintf("[%s] %s", label, strings.Join(children, " "))
}

// ── SHEET Render Engine ──────────────────────────────────────────────

// RenderSheet renders a SHEET component tree to ASCII.
func RenderSheet(root *Sheet) string {
	var render func(s *Sheet) string
	render = func(s *Sheet) string {
		// Render children first
		var childOutput []string
		for _, c := range s.Children {
			childOutput = append(childOutput, render(c))
		}

		// Built-in components
		switch s.Name {
		case "Box":
			return Box(s.Props, childOutput)
		case "Row":
			return Row(s.Props, childOutput)
		case "StatusBadge":
			return StatusBadge(s.Props, childOutput)
		case "text":
			content, _ := s.Props["content"].(string)
			return content
		default:
			// Custom render function
			if s.Render != nil {
				return s.Render(s.Props, childOutput)
			}
			return strings.Join(childOutput, "\n")
		}
	}
	return render(root)
}

// ── Example: SACRED Dashboard ─────────────────────────────────────────

// SacredDashboard renders the SACRED surface using SHEET components.
func SacredDashboard() string {
	root := H("Box", map[string]any{"title": "SACRED SURFACE", "width": 52},
		H("Row", map[string]any{"gap": " · "},
			H("StatusBadge", map[string]any{"label": "v100.1.0", "color": "ffaa00"},
				Text(CurrentVersion()),
			),
			H("StatusBadge", map[string]any{"label": "blocks", "color": "00d4ff"},
				Text(fmt.Sprint(len(schemaRegistry))),
			),
			H("StatusBadge", map[string]any{"label": "SACRED", "color": "00e660"},
				Text(func() string { if IsSacredIntact() { return "●" }; return "○" }()),
			),
		),
		Text(""),
		Text("The form is eternal. The loop is closed."),
		Text(fmt.Sprintf("M1/%s · %s · 7 recursions · 14 protocols", CurrentPOV().Arch, CurrentPOV().Tier)),
	)

	return RenderSheet(root)
}

// SheetStatus returns compact status.
func SheetStatus() string {
	return "sheet: react-like ASCII component engine · H() · Box · Row · StatusBadge · Text"
}

// SheetVakedFit returns Vaked fit.
func SheetVakedFit() string {
	return `SHEET = SACRED HYPERTEXT ELEMENT ENGINE FOR TERMINALS

  Like React, but for ASCII. Component-based.
  H(name, props, children...) → component tree.
  RenderSheet(root) → rendered ASCII.

  "Like a sheet of paper. Flat. Visible. Eternal." — CoCreator`
}
