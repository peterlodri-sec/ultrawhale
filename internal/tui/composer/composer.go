package composer

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

const (
	composerCollapseThreshold = 20
	composerHeadLines         = 3
	composerTailLines         = 2
	largePasteCharThreshold   = 1000
)

type Composer struct {
	textarea         textarea.Model
	width            int
	pendingPastes    []pendingPaste
	largePasteCounts map[int]int
	wrapCache        map[wrapCacheKey]int

	// rawCache memoizes textarea.Value() within a single tick. Each
	// pointer-receiver mutation method re-primes it after touching the
	// textarea. textarea.Value() walks every line and allocates the whole
	// buffer as a fresh string, and multiple call sites within one
	// Update() (AtEnd, reflow, splitComposerLines callers) used to pay
	// that cost repeatedly — the dominant cost during per-rune paste.
	rawCache      string
	rawCacheValid bool

	// selectionRuneOffset is the rune offset of the selection anchor in the
	// full text value. -1 means no active selection. The cursor position
	// (tracked by textarea) is the other end of the selection range.
	selectionRuneOffset int
}

type wrapCacheKey struct {
	line  string
	width int
}

// wrapCacheMaxEntries bounds Composer.wrapCache. Above this, the map is
// dropped wholesale: cheap, and a session that churns more than a few
// thousand distinct lines pays at most one full recompute per overflow.

const wrapCacheMaxEntries = 4096

type pendingPaste struct {
	placeholder string
	text        string
}

func New(width int) Composer {
	ta := textarea.New()
	ta.Placeholder = "Type message or command"
	ta.Prompt = "› " // visible prompt
	ta.SetPromptFunc(2, func(lineIdx int) string {
		if lineIdx == 0 {
			return "› "
		}
		return "  "
	})
	ta.ShowLineNumbers = false
	ta.CharLimit = 20000
	ta.MaxHeight = composerCollapseThreshold
	ta.SetWidth(width)
	ta.SetHeight(1)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#002240")).Foreground(lipgloss.Color("#00e660")).Bold(true)
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#444466")).Background(lipgloss.Color("#0a0a14"))
	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("#001020")).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("#00e660"))
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.Focus()
	return Composer{textarea: ta, width: width, selectionRuneOffset: -1}
}
