package blocks

import (
	"fmt"
	"strings"
)

// ── Display Primitive — Keyboard→Screen→TUI Pipeline ─────────────────
//
// The SACRED surface is a bidirectional pipeline:
//   KEYBOARD → INPUT → PROCESS → RENDER → SCREEN → HUMAN
//
// This is the Vaked "Reveals" layer made physical:
//   Human presses key → ultrawhale receives → engine processes → TUI renders → Human sees
//
// Display = the complete text rendering engine for the terminal.

// DisplayCell is one character cell on the screen.
type DisplayCell struct {
	Char     rune
	FgColor  string // AG-UI foreground
	BgColor  string // AG-UI background
	Bold     bool
	Italic   bool
	Underline bool
}

// DisplayBuffer is the full screen buffer.
type DisplayBuffer struct {
	Width  int
	Height int
	Cells  [][]DisplayCell
	CursorX int
	CursorY int
}

// DisplayEngine manages text rendering.
type DisplayEngine struct {
	Buffer   *DisplayBuffer
	Theme    string
	Stats    DisplayStats
}

// DisplayStats tracks rendering activity.
type DisplayStats struct {
	CharsRendered int64
	LinesDrawn    int64
	Refreshes     int64
	ScreenResizes int64
}

var displayEngine = &DisplayEngine{
	Buffer: &DisplayBuffer{Width: 80, Height: 24, CursorX: 0, CursorY: 0},
	Theme:  "dense",
}

// ── Display Pipeline ─────────────────────────────────────────────────

// KeyPress represents a keyboard input event.
type KeyPress struct {
	Key     string // "a", "enter", "tab", "up", "down", "escape"
	Ctrl    bool
	Alt     bool
	Shift   bool
	Meta    bool
}

// DisplayRender renders text to the screen buffer.
func DisplayRender(text string, style string) string {
	displayEngine.Stats.CharsRendered += int64(len(text))
	displayEngine.Stats.LinesDrawn++
	displayEngine.Stats.Refreshes++

	switch style {
	case "ascii_box":
		return renderASCIIBox(text)
	case "pipeline":
		return renderPipeline(text)
	case "hud":
		return renderHUDLine(text)
	default:
		return text
	}
}

func renderASCIIBox(text string) string {
	width := 60
	lines := strings.Split(text, "\n")
	var sb strings.Builder
	sb.WriteString("┌" + strings.Repeat("─", width-2) + "┐\n")
	for _, line := range lines {
		padded := line
		if len(padded) < width-4 { padded += strings.Repeat(" ", width-4-len(padded)) }
		sb.WriteString("│ " + padded + " │\n")
	}
	sb.WriteString("└" + strings.Repeat("─", width-2) + "┘")
	return sb.String()
}

func renderPipeline(text string) string {
	stages := strings.Split(text, "→")
	var sb strings.Builder
	for i, stage := range stages {
		stage = strings.TrimSpace(stage)
		sb.WriteString("[" + stage + "]")
		if i < len(stages)-1 { sb.WriteString("──→") }
	}
	return sb.String()
}

func renderHUDLine(text string) string {
	return fmt.Sprintf("│ %s │", text)
}

// ── Keyboard Navigation ───────────────────────────────────────────────

// ArrowNavigate handles arrow key presses in the TUI.
func ArrowNavigate(key KeyPress, currentLine, maxLines int) int {
	switch key.Key {
	case "up", "k":
		if currentLine > 0 { return currentLine - 1 }
	case "down", "j":
		if currentLine < maxLines-1 { return currentLine + 1 }
	case "home", "gg":
		return 0
	case "end", "G":
		return maxLines - 1
	}
	return currentLine
}

// ── Display Status ────────────────────────────────────────────────────

// DisplayStatus returns compact display status.
func DisplayStatus() string {
	return fmt.Sprintf("display: %dx%d · %d chars · %d lines · %d refreshes · theme: %s",
		displayEngine.Buffer.Width, displayEngine.Buffer.Height,
		displayEngine.Stats.CharsRendered, displayEngine.Stats.LinesDrawn,
		displayEngine.Stats.Refreshes, displayEngine.Theme)
}

// DisplayVakedFit returns the display engine's Vaked fit.
func DisplayVakedFit() string {
	return `Keyboard → Screen → Display → Text → ASCII → TUI
    ↑                                              ↑
  SACRED input                              SACRED output

The display IS the Reveals layer made physical.
Every keypress is sacred. Every pixel is a declaration.
Context × Time × Space converge on the screen.`
}
