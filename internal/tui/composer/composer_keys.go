package composer

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (c *Composer) Update(msg tea.Msg) tea.Cmd {
	c.ensureInitialized()
	wasAtEnd := c.AtEnd()
	prevHeight := c.textarea.Height()
	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)
	c.markRawCacheStale()
	c.prunePendingPastes()
	c.reflow()
	if wasAtEnd && c.textarea.Height() > prevHeight {
		c.realignViewportAtEnd()
	}
	return cmd
}

func (c *Composer) HandleKey(msg tea.KeyMsg) bool {
	c.ensureInitialized()
	switch msg.String() {
	case "ctrl+p", "ctrl+n":
		return false
	case "ctrl+j", "shift+enter":
		c.InsertNewline()
		return true
	case "up":
		return c.moveFoldedVisibleLine(-1)
	case "down":
		return c.moveFoldedVisibleLine(1)
	case "pgup":
		for c.textarea.Line() > 0 {
			c.textarea.CursorUp()
		}
		c.textarea.CursorStart()
		return true
	case "pgdown":
		c.moveToEnd()
		return true
	}
	return false
}

func (c *Composer) moveToEnd() {
	for c.textarea.Line() < c.textarea.LineCount()-1 {
		c.textarea.CursorDown()
	}
	c.textarea.CursorEnd()
}

func (c *Composer) moveCursorToRuneOffset(offset int) {
	if offset < 0 {
		offset = 0
	}
	lines := splitComposerLines(c.rawValue())
	line := 0
	col := offset
	for line < len(lines) {
		lineLen := len([]rune(lines[line]))
		if col <= lineLen {
			break
		}
		col -= lineLen + 1
		line++
	}
	if line >= len(lines) {
		c.moveToEnd()
		return
	}
	c.moveToLine(line)
	c.textarea.SetCursor(col)
}

func (c *Composer) moveToLine(line int) {
	line = max(0, min(line, c.textarea.LineCount()-1))
	for c.textarea.Line() > line {
		c.textarea.CursorUp()
	}
	for c.textarea.Line() < line {
		c.textarea.CursorDown()
	}
}

func (c *Composer) moveFoldedVisibleLine(direction int) bool {
	lines := splitComposerLines(c.rawValue())
	if len(lines) <= composerCollapseThreshold {
		return false
	}
	visible := foldedVisibleLineIndexes(len(lines), c.textarea.Line())
	current := c.textarea.Line()
	switch {
	case direction < 0:
		for i := len(visible) - 1; i >= 0; i-- {
			if visible[i] < current {
				c.moveToLine(visible[i])
				c.reflow()
				return true
			}
		}
	case direction > 0:
		for _, line := range visible {
			if line > current {
				c.moveToLine(line)
				c.reflow()
				return true
			}
		}
	}
	return true
}

func (c *Composer) realignViewportAtEnd() {
	value := c.textarea.Value()
	c.textarea.SetValue(value)
}
