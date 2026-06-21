package blocks

import (
	"fmt"
	"strings"
)

// ── Spoken Language — HUMAN_TERM Encoding Dialect ─────────────────────
//
// Inspired by zerolang (vercel-labs/zerolang).
// Humans speak in outcomes. Agents operate on graph semantics.
//
// ultrawhale's UI-medium now supports SPOKEN LANGUAGE as encoding:
//   EN:  "build auth" → blocks engine
//   HU:  "építsd meg az autentikációt" → same blocks engine
//   DE:  "baue die Authentifizierung" → same blocks engine
//   FI:  "rakenna todennus" → same blocks engine
//
// The SACRED surface does not care what language you speak.
// It translates to the capability graph. The graph IS the truth.

// SpokenLanguage is a human term encoding dialect.
type SpokenLanguage struct {
	Code     string // "en", "hu", "de", "fi", "jp", "zh"
	Name     string // "English", "Magyar", "Deutsch", "Suomi"
	Encoding string // "utf-8", "utf-16"
	Active   bool
}

// SpokenLanguageRegistry maps language codes to their dialects.
type SpokenLanguageRegistry struct {
	Languages map[string]*SpokenLanguage
	Default   string
}

var spokenLanguages = &SpokenLanguageRegistry{
	Languages: map[string]*SpokenLanguage{
		"en": {Code: "en", Name: "English", Encoding: "utf-8", Active: true},
		"hu": {Code: "hu", Name: "Magyar", Encoding: "utf-8", Active: true},
		"de": {Code: "de", Name: "Deutsch", Encoding: "utf-8", Active: true},
		"fi": {Code: "fi", Name: "Suomi", Encoding: "utf-8", Active: true},
		"jp": {Code: "jp", Name: "日本語", Encoding: "utf-8", Active: true},
		"zh": {Code: "zh", Name: "中文", Encoding: "utf-8", Active: true},
	},
	Default: "en",
}

// ── Spoken Language Operations ────────────────────────────────────────

// SpeakIn translates a human utterance to the capability graph.
// Any language → same graph operation.
func SpeakIn(language, utterance string) string {
	lang, ok := spokenLanguages.Languages[language]
	if !ok { language = spokenLanguages.Default; lang = spokenLanguages.Languages[language] }

	// Translate the utterance to a capability graph operation
	// In production: NLP → semantic parse → graph operation
	// For now: route through engine
	result := EngineMaterialize(utterance)

	return fmt.Sprintf("[%s/%s] %s → %s",
		lang.Code, lang.Name, utterance[:min(40, len(utterance))], result)
}

// ── Zerolang-inspired: Graph-native operations ────────────────────────

// GraphNativeOp executes a graph-native operation from a human utterance.
// Like zerolang: humans speak → agents patch graph → compiler validates.
func GraphNativeOp(utterance string) string {
	lower := strings.ToLower(utterance)

	switch {
	case strings.Contains(lower, "build") || strings.Contains(lower, "építsd"):
		return "graph: addMain → patch → compile → run"
	case strings.Contains(lower, "fix") || strings.Contains(lower, "javítsd"):
		return "graph: query → diagnose → patch → verify"
	case strings.Contains(lower, "explain") || strings.Contains(lower, "magyarázd"):
		return "graph: lookup → trace → summarize"
	case strings.Contains(lower, "deploy") || strings.Contains(lower, "telepítsd"):
		return "graph: check → build → push → verify"
	default:
		return fmt.Sprintf("graph: query → %s", utterance[:min(30, len(utterance))])
	}
}

// ── Status ────────────────────────────────────────────────────────────

// SpokenLanguageStatus returns compact spoken language status.
func SpokenLanguageStatus() string {
	active := 0
	for _, l := range spokenLanguages.Languages {
		if l.Active { active++ }
	}
	return fmt.Sprintf("spoken-language: %d dialects · default: %s",
		active, spokenLanguages.Default)
}

// SpokenLanguageVakedFit returns spoken language Vaked fit.
func SpokenLanguageVakedFit() string {
	return `SPOKEN LANGUAGE = HUMAN_TERM ENCODING DIALECT

  Inspired by zerolang (vercel-labs/zerolang).
  Humans speak in outcomes. Agents operate on graph semantics.

  6 languages: EN, HU, DE, FI, JP, ZH.
  The SACRED surface does not care what language you speak.
  It translates to the capability graph. The graph IS the truth.

  "The SACRED surface does not care what language you speak."
  — Peter, v100`
}
