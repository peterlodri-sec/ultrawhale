package agui

import (
	"github.com/charmbracelet/lipgloss"
)

// ChatBlock is a rendered message block in the TUI chat view.
// Wraps AG-UI BlockType with content and an optional tool name.
type ChatBlock struct {
	Type    BlockType
	Title   string
	Content string
	Width   int
	Streaming bool // true if content is still arriving
}

// Render returns the chat block rendered with AG-UI theme chrome.
func (cb ChatBlock) Render() string {
	t := Current
	icon := blockIcon(cb.Type)
	color := blockColor(cb.Type, t)

	header := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(icon + " " + cb.Title)

	body := lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "▎"}, false, false, false, true).
		BorderForeground(color).
		Padding(0, 2).
		Width(cb.Width - 4).
		Foreground(t.Fg).
		Render(cb.Content)

	suffix := ""
	if cb.Streaming { suffix = " ⏳" }
	return header + suffix + "\n" + body + "\n"
}

// NewChatBlock creates a typed chat block from common patterns.
func NewChatBlock(kind string, title, content string, width int) ChatBlock {
	bt := classifyBlock(kind)
	return ChatBlock{Type: bt, Title: title, Content: content, Width: width}
}

func classifyBlock(kind string) BlockType {
	switch kind {
	case "thinking":
		return BlockThinking
	case "tool_call", "tool-call", "tool_use":
		return BlockToolCall
	case "tool_result", "tool-result", "observation":
		return BlockToolResult
	case "diff", "code_diff", "code-diff":
		return BlockCodeDiff
	case "plan", "plan_card", "plan-card":
		return BlockPlanCard
	case "file_tree", "file-tree", "directory":
		return BlockFileTree
	default:
		return BlockText
	}
}


// OneShot renders a complete Vaked pipeline: declare → materialize → reveal.
// Single-pass from .vaked declaration to AG-UI rendered block.
func OneShot(declaration string) string {
	// Declares: the input is a Vaked declaration
	// Materializes: render as AG-UI block
	// Reveals: return the rendered output
	
	block := NewChatBlock("plan_card", "Vaked OneShot", declaration, 80)
	block.Streaming = false
	return block.Render()
}

// StreamingBlock renders a partial ChatBlock with ⏳ suffix.
func StreamingBlock(kind, title, content string, width int) string {
	block := NewChatBlock(kind, title, content, width)
	block.Streaming = true
	return block.Render()
}
