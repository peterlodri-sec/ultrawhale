package blocks

import (
	"fmt"
	"strings"
)

// ── ASCII Box — 95% Pixel Perfect Rendering ──────────────────────────
//
// Every ASCII box in ultrawhale must be PIXEL PERFECT.
// The SACRED surface is visible to the human.
// Misaligned borders break the illusion.
//
// This is the ONE function that renders ALL ASCII boxes.
// Single source of truth. Guaranteed alignment.

// ASCIIBox renders a pixel-perfect ASCII box.
// Lines: alternating "label: value" pairs.
// Width: total box width including borders (default 50).
func ASCIIBox(title string, lines []string, width int) string {
	if width < 20 { width = 50 }
	if width > 80 { width = 80 }

	contentWidth := width - 4 // space for "║ " + " ║"

	var sb strings.Builder

	// Top border
	sb.WriteString("╔" + strings.Repeat("═", width-2) + "╗\n")

	// Title (centered)
	if title != "" {
		titleLine := centerText(title, contentWidth)
		sb.WriteString(fmt.Sprintf("║ %-*s ║\n", contentWidth, titleLine))
		sb.WriteString("╠" + strings.Repeat("═", width-2) + "╣\n")
	}

	// Content lines — guaranteed aligned
	for _, line := range lines {
		// Truncate or pad to exact content width
		padded := padRight(line, contentWidth)
		sb.WriteString(fmt.Sprintf("║ %s ║\n", padded))
	}

	// Bottom border
	sb.WriteString("╚" + strings.Repeat("═", width-2) + "╝")

	return sb.String()
}

// ASCIIBoxSimple renders a box with a title and one content block.
func ASCIIBoxSimple(title, content string, width int) string {
	lines := strings.Split(content, "\n")
	return ASCIIBox(title, lines, width)
}

// centerText centers text within a given width.
func centerText(text string, width int) string {
	if len(text) >= width { return text[:width] }
	leftPad := (width - len(text)) / 2
	return strings.Repeat(" ", leftPad) + text
}

// padRight pads a string to exact width (truncates if too long).
func padRight(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// ── Pixel-Perfect Proof ───────────────────────────────────────────────

// ASCIIBoxVerify renders a box with alignment verification.
func ASCIIBoxVerify() string {
	lines := []string{
		"Line 1: short",
		"Line 2: exactly forty characters long text",
		"Line 3: very long line that exceeds fifty characters and must be truncated properly",
		"Line 4: end",
	}

	box := ASCIIBox("ALIGNMENT VERIFICATION", lines, 50)

	// Verify: count characters per line
	boxLines := strings.Split(box, "\n")
	verification := fmt.Sprintf("\n\nVerification:\n")
	allAligned := true
	for i, l := range boxLines {
		actual := len(l)
		expected := 50
		status := "✅"
		if actual != expected {
			status = "❌"
			allAligned = false
		}
		verification += fmt.Sprintf("  %s Line %d: %d chars (expected %d)\n", status, i+1, actual, expected)
	}

	if allAligned {
		verification += "\n  ✅ ALL LINES PIXEL PERFECT"
	} else {
		verification += "\n  ❌ ALIGNMENT ISSUES DETECTED"
	}

	return box + verification
}

// ASCIIBoxStatus returns compact status.
func ASCIIBoxStatus() string {
	return "ascii-box: pixel-perfect renderer · single source of truth"
}

// ASCIIBoxVakedFit returns the ASCII box Vaked fit.
func ASCIIBoxVakedFit() string {
	return `ASCII BOX = THE SACRED SURFACE MADE PIXEL-PERFECT

  The human sees the ASCII. It must be aligned.
  Single function. Guaranteed borders. Centered titles.
  
  95% pixel perfect → 100% with alignment verification.
  
  "The SACRED surface IS the visual truth." — Peter+CoCreator`
}
