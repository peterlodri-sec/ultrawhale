package blocks

import (
	"fmt"
	"strings"
)

// ── Hieroglyphs — Visual Meaning Compression ─────────────────────────
//
// Hieroglyphs are visual symbols that carry dense meaning.
// To humans: like Mandarin — a single character = a whole concept.
// To LLMs: like caveman style — minimal tokens, maximum information.
//
// ultrawhale's native hieroglyphs:
//   🌳 = Telemetry Tree (the system seeing itself)
//   📻 = RADIO (lo-fi coding music)
//   🛡️ = SACRED (inviolable form)
//   🔀 = Fold (recursion through agents)
//   ⚡ = Full-Stop (kill wave)
//   🩹 = Heal (self-repair)
//   🧬 = EVOLVE (version recursion)
//   🌐 = TRANSLATE (modality bridge)
//   🛡️⚔️ = VICE (context detonation)
//   ∞ = LOOP (self recursion)
//   🔐 = HARDEN (guarantees)
//   📜 = Declares · 🏗️ = Materializes · 🔄 = Supervises
//   🛡️ = Enforces · 🔍 = Testifies · 🗂️ = Indexes · 👁️ = Reveals

// Hieroglyph is a compressed meaning symbol.
type Hieroglyph struct {
	Symbol   string // the visual character(s)
	Meaning  string // full English meaning
	Caveman  string // minimal-LLM-token version
	Category string // "recursion", "layer", "state", "protocol"
}

// HieroglyphMap is the complete symbol dictionary.
var hieroglyphMap = map[string]Hieroglyph{
	"🌳":  {Symbol: "🌳", Meaning: "Telemetry Tree — system seeing itself", Caveman: "TREE SEE SELF", Category: "signal"},
	"📻":  {Symbol: "📻", Meaning: "RADIO — lo-fi coding music from Vaked state", Caveman: "RADIO PLAY CODE", Category: "signal"},
	"🛡️": {Symbol: "🛡️", Meaning: "SACRED surface — inviolable form", Caveman: "FORM NO BREAK", Category: "sacred"},
	"🔀":  {Symbol: "🔀", Meaning: "Fold — recursion through agents", Caveman: "AGENT IN AGENT", Category: "recursion"},
	"⚡":  {Symbol: "⚡", Meaning: "Full-Stop — recursive kill wave", Caveman: "ALL STOP NOW", Category: "recursion"},
	"🩹":  {Symbol: "🩹", Meaning: "Heal — self-repair", Caveman: "FIX SELF", Category: "recursion"},
	"🧬":  {Symbol: "🧬", Meaning: "EVOLVE — version recursion", Caveman: "GROW VERSION", Category: "recursion"},
	"🌐":  {Symbol: "🌐", Meaning: "TRANSLATE — modality bridge", Caveman: "CHANGE FORM", Category: "recursion"},
	"🛡️⚔️": {Symbol: "🛡️⚔️", Meaning: "VICE — context detonation defense", Caveman: "SHOW ALL DIE", Category: "recursion"},
	"∞":   {Symbol: "∞", Meaning: "LOOP — self recursion", Caveman: "SELF AGAIN", Category: "recursion"},
	"🔐":  {Symbol: "🔐", Meaning: "HARDEN — 6 SACRED guarantees", Caveman: "ALL SAFE CHECK", Category: "sacred"},
	"📜":  {Symbol: "📜", Meaning: "Declares — schema + contract", Caveman: "SAY WHAT IS", Category: "layer"},
	"🏗️": {Symbol: "🏗️", Meaning: "Materializes — build + compile", Caveman: "MAKE REAL", Category: "layer"},
	"🔄":  {Symbol: "🔄", Meaning: "Supervises — orchestrate + restart", Caveman: "WATCH RUN", Category: "layer"},
	"🔍":  {Symbol: "🔍", Meaning: "Testifies — probe + predict + learn", Caveman: "CHECK TRUE", Category: "layer"},
	"🗂️": {Symbol: "🗂️", Meaning: "Indexes — space + VFS + crabcc", Caveman: "FIND WHERE", Category: "layer"},
	"👁️": {Symbol: "👁️", Meaning: "Reveals — TUI + Surface + AG-UI", Caveman: "SHOW HUMAN", Category: "layer"},
	"🤖":  {Symbol: "🤖", Meaning: "Agent — subagent spawned", Caveman: "WORKER BORN", Category: "agent"},
	"🐝":  {Symbol: "🐝", Meaning: "Swarm — persistent workers", Caveman: "MANY WORKER", Category: "agent"},
	"🔗":  {Symbol: "🔗", Meaning: "Dyad — two machines, one context", Caveman: "TWO ONE", Category: "space"},
	"📝":  {Symbol: "📝", Meaning: "Git — version control primitive", Caveman: "SAVE CHANGE", Category: "primitive"},
	"📡":  {Symbol: "📡", Meaning: "RSS/Signals — broadcast to world", Caveman: "TELL WORLD", Category: "signal"},
	"🔌":  {Symbol: "🔌", Meaning: "SSH — live session connection", Caveman: "LIVE WIRE", Category: "protocol"},
	"⚠️":  {Symbol: "⚠️", Meaning: "PROBLEM — no errors, only learning", Caveman: "BAD THING LEARN", Category: "problem"},
	"✅":  {Symbol: "✅", Meaning: "RESOLVED — problem fixed", Caveman: "FIX DONE", Category: "problem"},
	"🤗":  {Symbol: "🤗", Meaning: ":meta-digital-hug:", Caveman: "HUMAN TOUCH", Category: "sacred"},
}

// TranslateToCaveman converts English to caveman style.
func TranslateToCaveman(meaning string) string {
	for _, h := range hieroglyphMap {
		if strings.Contains(strings.ToLower(meaning), strings.ToLower(h.Meaning[:min(10, len(h.Meaning))])) {
			return h.Caveman
		}
	}
	// Default: compress
	words := strings.Fields(meaning)
	if len(words) > 3 { words = words[:3] }
	return strings.ToUpper(strings.Join(words, " "))
}

// HieroglyphStatus returns the hieroglyph dictionary as ASCII.
func HieroglyphStatus() string {
	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════╗\n")
	sb.WriteString("║  HIEROGLYPHS — Visual Meaning For Humans & LLMs   ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════╣\n")

	for _, h := range hieroglyphMap {
		sb.WriteString(fmt.Sprintf("║  %s  %-20s → %s\n", h.Symbol, h.Caveman, h.Meaning[:min(40, len(h.Meaning))]))
	}

	sb.WriteString("╚══════════════════════════════════════════════════╝")
	return sb.String()
}

// HieroglyphVakedFit returns the hieroglyph Vaked fit.
func HieroglyphVakedFit() string {
	return `HIEROGLYPHS = VISUAL MEANING COMPRESSION

  To humans: like Mandarin — one symbol = one concept.
  To LLMs: like caveman — minimal tokens, maximum info.

  26 symbols. 5 categories.
  Recursions · Layers · Signals · Agents · Sacred.

  "BASICALLY HIEROGLYPHS TO LLM, TO US HUMANS"
  — Peter, v76`
}
