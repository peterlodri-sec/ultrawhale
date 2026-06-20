package agui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBlock renders a single content block with AG-UI chrome: icon, title, colored left border.
func RenderBlock(blockType BlockType, title, content string, width int) string {
	t := Current
	icon := blockIcon(blockType)
	color := blockColor(blockType, t)

	header := lipgloss.NewStyle().Foreground(color).Render(icon + " " + title)
	body := lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "▎"}, false, false, false, true).
		BorderForeground(color).
		Padding(0, 1).
		Width(width - 2).
		Foreground(t.Fg).
		Render(content)

	return header + "\n" + body
}

// BlockType identifies the kind of content being rendered.
type BlockType int

const (
	BlockText      BlockType = iota
	BlockThinking
	BlockToolCall
	BlockToolResult
	BlockCodeDiff
	BlockPlanCard
	BlockFileTree
)

func blockIcon(bt BlockType) string {
	switch bt {
	case BlockThinking:   return "⏳"
	case BlockToolCall:   return "🔧"
	case BlockToolResult: return "📋"
	case BlockCodeDiff:   return "Δ"
	case BlockPlanCard:   return "📐"
	case BlockFileTree:   return "📁"
	default:              return "·"
	}
}

func blockColor(bt BlockType, t Theme) lipgloss.Color {
	switch bt {
	case BlockThinking:   return t.Dim
	case BlockToolCall:   return t.Accent
	case BlockToolResult: return t.Fg
	case BlockCodeDiff:   return lipgloss.Color("#00e660")
	case BlockPlanCard:   return lipgloss.Color("#00d4ff")
	case BlockFileTree:   return t.Accent
	default:              return t.Fg
	}
}

func PadRight(s string, width int) string {
	if len(s) >= width { return s }
	return s + strings.Repeat(" ", width-len(s))
}
