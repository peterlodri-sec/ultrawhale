package composer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tuitheme "github.com/usewhale/whale/internal/tui/theme"
)

func (c Composer) View() string {
	c = c.initialized()
	var view string
	if c.rawValue() == "" {
		copy := c.textarea
		copy.SetHeight(1)
		view = copy.View()
	} else {
		value := c.rawValue()
		lines := splitComposerLines(value)
		if len(lines) <= composerCollapseThreshold {
			view = c.plainView(lines)
		} else {
			view = c.foldedView(lines)
		}
	}
	return c.normalizeView(view)
}

func (c Composer) foldedView(lines []string) string {
	cursorLine := c.textarea.Line()
	keep := map[int]bool{}
	for _, line := range foldedVisibleLineIndexes(len(lines), cursorLine) {
		keep[line] = true
	}

	out := make([]string, 0, composerHeadLines+composerTailLines+4)
	prev := -1
	for i := 0; i < len(lines); i++ {
		if !keep[i] {
			continue
		}
		if i-prev > 1 {
			out = append(out, c.hiddenLine(i-prev-1))
		}
		out = append(out, c.promptLine(lines[i], i == 0, i == cursorLine))
		prev = i
	}
	out = append(out, c.hintLine(len(lines)))
	return strings.Join(out, "\n")
}

func foldedVisibleLineIndexes(lineCount int, cursorLine int) []int {
	if lineCount <= 0 {
		return nil
	}
	keep := map[int]bool{}
	for i := 0; i < composerHeadLines && i < lineCount; i++ {
		keep[i] = true
	}
	for i := max(0, lineCount-composerTailLines); i < lineCount; i++ {
		keep[i] = true
	}
	if cursorLine >= 0 && cursorLine < lineCount {
		keep[cursorLine] = true
	}

	out := make([]int, 0, len(keep))
	for i := 0; i < lineCount; i++ {
		if keep[i] {
			out = append(out, i)
		}
	}
	return out
}

func (c Composer) plainView(lines []string) string {
	cursorLine := c.textarea.Line()
	cursorCol := -1
	if cursorLine >= 0 && cursorLine < len(lines) {
		info := c.textarea.LineInfo()
		cursorCol = info.StartColumn + info.ColumnOffset
	}

	out := make([]string, 0, c.visualLineCount())
	displayLine := 0
	wrapWidth := c.textarea.Width()
	for i, line := range lines {
		lineRunes := []rune(line)
		for _, segment := range wrapComposerLine(line, wrapWidth) {
			hasCursor := false
			relativeCursor := 0
			if i == cursorLine {
				switch {
				case cursorCol >= segment.start && cursorCol < segment.end:
					hasCursor = true
					relativeCursor = cursorCol - segment.start
				case cursorCol == len(lineRunes) && segment.end == len(lineRunes):
					hasCursor = true
					relativeCursor = len([]rune(segment.text))
				case len(lineRunes) == 0 && segment.start == 0 && segment.end == 0:
					hasCursor = true
				}
			}
			out = append(out, c.promptLineAt(segment.text, displayLine == 0, hasCursor, relativeCursor))
			displayLine++
		}
	}
	return strings.Join(out, "\n")
}

func (c Composer) promptLine(line string, first bool, cursor bool) string {
	info := c.textarea.LineInfo()
	return c.promptLineAt(line, first, cursor, info.StartColumn+info.ColumnOffset)
}

func (c Composer) promptLineAt(line string, first bool, cursor bool, col int) string {
	prefix := "  "
	if first {
		prefix = lipgloss.NewStyle().Foreground(tuitheme.Default.Accent).Bold(true).Render("›") + " "
	}
	return prefix + renderComposerLineText(line, cursor, col)
}

func renderComposerLineText(line string, cursor bool, col int) string {
	runes := []rune(line)
	if col < 0 {
		col = 0
	}
	if col > len(runes) {
		col = len(runes)
	}
	var out strings.Builder
	if cursor && col == 0 {
		out.WriteString("█")
	}
	for i, r := range runes {
		out.WriteRune(r)
		if cursor && col == i+1 {
			out.WriteString("█")
		}
	}
	if len(runes) == 0 && cursor && col == 0 {
		return "█"
	}
	return out.String()
}

func (c Composer) hiddenLine(n int) string {
	return lipgloss.NewStyle().
		Foreground(tuitheme.Default.Muted).
		Render(fmt.Sprintf("  [… %d lines hidden - full content kept …]", n))
}

func (c Composer) hintLine(n int) string {
	return lipgloss.NewStyle().
		Foreground(tuitheme.Default.Muted).
		Render(fmt.Sprintf("  [%d lines · Ctrl+A/E/K/U line · Ctrl+W/Alt+B,F word · Ctrl+C clear · PgUp/PgDn]", n))
}

func splitComposerLines(value string) []string {
	if value == "" {
		return []string{""}
	}
	return strings.Split(value, "\n")
}

func (c Composer) normalizeView(view string) string {
	view = strings.TrimRight(view, "\n")
	if view == "" {
		return ""
	}
	lines := strings.Split(view, "\n")
	padded := make([]string, 0, len(lines))
	style := lipgloss.NewStyle().Width(c.width).MaxWidth(c.width)
	for _, line := range lines {
		padded = append(padded, style.Render(line))
	}
	return strings.Join(padded, "\n")
}
