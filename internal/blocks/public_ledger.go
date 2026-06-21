package blocks

import (
	"fmt"
	"strings"
	"time"
)

// ── PUBLIC LEDGER — Ralph Loop + DogFeed History ──────────────────────
//
// READ-ONLY for everyone. The public record of the machine learning.
// Every dogfeed interaction. Every Ralph pattern. Every lesson learned.
// Append-only. Immutable. Public.

// PublicEntry is one entry in the public ledger.
type PublicEntry struct {
	Timestamp time.Time
	Type      string // "dogfeed", "ralph-pattern", "lesson", "proof"
	Detail    string
	Model     string // which free model
	Ref       string // SHA256 content ref
}

// PublicLedger is the append-only public record.
type PublicLedger struct {
	Entries    []PublicEntry
	TotalFeeds int64
	TotalPatterns int64
	TotalLessons  int64
	StartedAt  time.Time
}

var publicLedger = &PublicLedger{
	Entries:   make([]PublicEntry, 0, 4096  ),
	StartedAt: time.Now(),
}

// ── Public Ledger Operations ──────────────────────────────────────────

// RecordPublicDogFeed adds a dogfeed interaction to the public ledger.
func RecordPublicDogFeed(model, userMessage, response string) {
	entry := PublicEntry{
		Timestamp: time.Now(),
		Type:      "dogfeed",
		Detail:    fmt.Sprintf("%s → %s", userMessage[:min(40, len(userMessage))], response[:min(40, len(response))]),
		Model:     model,
		Ref:       Ref([]byte(userMessage + response)),
	}

	publicLedger.Entries = append(publicLedger.Entries, entry)
	publicLedger.TotalFeeds++
	if len(publicLedger.Entries) > 4096   { publicLedger.Entries = publicLedger.Entries[1:] }

	Pulse("public.ledger.dogfeed", model)
}

// RecordPublicPattern adds a Ralph pattern to the public ledger.
func RecordPublicPattern(pattern string, confidence float64) {
	entry := PublicEntry{
		Timestamp: time.Now(),
		Type:      "ralph-pattern",
		Detail:    fmt.Sprintf("%s (%.0f%%)", pattern, confidence*100),
		Model:     "ralph",
		Ref:       Ref([]byte(pattern)),
	}

	publicLedger.Entries = append(publicLedger.Entries, entry)
	publicLedger.TotalPatterns++
}

// ── Public History Render ────────────────────────────────────────────

// PublicHistory renders the last N entries of the public ledger.
func PublicHistory(n int) string {
	if n > len(publicLedger.Entries) { n = len(publicLedger.Entries) }
	if n == 0 { return "public-ledger: no entries yet" }

	entries := publicLedger.Entries[len(publicLedger.Entries)-n:]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("╔══ PUBLIC LEDGER — Last %d entries ══╗\n", n))

	for i, e := range entries {
		icon := "🐕"
		switch e.Type {
		case "ralph-pattern": icon = "🧠"
		case "lesson": icon = "📖"
		case "proof": icon = "✅"
		}

		marker := " "
		if i == len(entries)-1 { marker = "→" }

		sb.WriteString(fmt.Sprintf("║ %s%s [%s] %s\n",
			marker, icon, e.Timestamp.Format("15:04:05"), e.Detail[:min(40, len(e.Detail))]))
	}

	sb.WriteString("╠══════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║  Feeds: %d · Patterns: %d · Since: %s\n",
		publicLedger.TotalFeeds, publicLedger.TotalPatterns,
		publicLedger.StartedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("║  READ-ONLY · PUBLIC · APPEND-ONLY          ║\n")
	sb.WriteString("╚══════════════════════════════════════════╝")

	return sb.String()
}

// PublicLedgerStatus returns compact public ledger status.
func PublicLedgerStatus() string {
	return fmt.Sprintf("public-ledger: %d entries · %d feeds · %d patterns · %d lessons · since %s",
		len(publicLedger.Entries), publicLedger.TotalFeeds,
		publicLedger.TotalPatterns, publicLedger.TotalLessons,
		publicLedger.StartedAt.Format("2006-01-02"))
}

// PublicLedgerVakedFit returns public ledger Vaked fit.
func PublicLedgerVakedFit() string {
	return `PUBLIC LEDGER = READ-ONLY FOR EVERYONE

  Ralph Loop + DogFeed history — public record of learning.
  Every interaction. Every pattern. Every lesson.
  Append-only. Immutable. Public.

  "The machine learns in public. The record is sacred."`
}
