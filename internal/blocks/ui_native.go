package blocks

import (
	"fmt"
	"strings"
)

// ── UI Native — Co-Creative Suggestions ───────────────────────────────
// Counter, Progress, Sparkline, Theme — implemented from the co-creative review.

// LiveCounter returns a live counter string.
func LiveCounter(label string, value int64) string {
	return fmt.Sprintf("%s: %d", label, value)
}

// TokensCounter returns live token count.
func TokensCounter() string {
	c := GetRealCost()
	return fmt.Sprintf("tokens: %d API + %d folded = %d total · cost: $%.4f",
		c.APIInputTokens+c.APIOutputTokens, c.FoldedInputTokens+c.FoldedOutputTokens,
		c.TotalTokens, c.TotalCost)
}

// ProgressBar renders an ASCII progress bar.
func ProgressBar(current, total int, width int) string {
	if width <= 0 { width = 40 }
	if total <= 0 { return "[" + strings.Repeat(" ", width) + "] 0%" }

	filled := current * width / total
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	pct := current * 100 / total
	return fmt.Sprintf("[%s] %d%%", bar, pct)
}

// Sparkline renders a tiny inline graph from values.
func Sparkline(values []int64) string {
	if len(values) == 0 { return "" }
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Find min/max
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		if v < minVal { minVal = v }
		if v > maxVal { maxVal = v }
	}

	if maxVal == minVal { return strings.Repeat(string(chars[0]), len(values)) }

	var sb strings.Builder
	for _, v := range values {
		idx := (v - minVal) * int64(len(chars)-1) / (maxVal - minVal)
		sb.WriteRune(chars[idx])
	}
	return sb.String()
}

// SetTheme changes the active AG-UI theme.
func SetTheme(theme string) string {
	switch theme {
	case "dense", "cyberpunk", "graveyard":
		uiCoCreative.Theme = theme
		Log(LogInfo, "theme.set", theme, "", "", 0, nil)
		return fmt.Sprintf("theme: %s activated", theme)
	default:
		return fmt.Sprintf("theme: %s not found (dense, cyberpunk, graveyard)", theme)
	}
}

// ThemeStatus returns current theme.
func ThemeStatus() string {
	return fmt.Sprintf("theme: %s", uiCoCreative.Theme)
}

// UINativeStatus returns compact UI native status.
func UINativeStatus() string {
	return fmt.Sprintf("ui-native: counter · progress · sparkline · theme · 15 native things")
}
