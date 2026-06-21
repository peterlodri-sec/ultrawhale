package blocks

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ── Best Practices — Go + Systems Engineering ─────────────────────────
//
// Self-reflection on our 133-block codebase. What patterns emerged?
// What should we formalize as best practices?
//
// 1. Atomic state → sync/atomic for counters, mutex for structs
// 2. Single source of truth → one function per concept (ASCIIBox, Ref, POV)
// 3. Append-only ledgers → promises, proofs, claims — never delete
// 4. Graceful degradation → DREAM state, entropy fallback, SEALING reserve
// 5. Content addressing → every operation has a Ref (SHA256)
// 6. Pulse-driven → every state change emits a Pulse for telemetry
// 7. SACRED surface → every UI element flows through ASCIIBox
// 8. VakedFit → every primitive declares its Vaked layer

// BestPractice is one codified pattern.
type BestPractice struct {
	Name        string
	Category    string // "go", "systems", "vaked"
	Description string
	Example     string
	Adopted     bool
}

// BestPracticeRegistry tracks which practices we follow.
type BestPracticeRegistry struct {
	mu        sync.Mutex
	Practices []BestPractice
}

var bestPractices = &BestPracticeRegistry{
	Practices: []BestPractice{
		{Name: "atomic-state", Category: "go",
			Description: "Use sync/atomic for counters, sync.Mutex for structs",
			Example: "var loopState atomic.Int32", Adopted: true},
		{Name: "single-truth", Category: "systems",
			Description: "One function per concept — ASCIIBox(), Ref(), POV()",
			Example: "ASCIIBox(title, lines, width) → every box", Adopted: true},
		{Name: "append-only", Category: "systems",
			Description: "Ledgers never delete — promises, proofs, claims",
			Example: "promiseLedger.Promises = append(...)", Adopted: true},
		{Name: "graceful-degrade", Category: "systems",
			Description: "DREAM state, entropy fallback, SEALING 10% reserve",
			Example: "SetMainState(StateDream) — never exit", Adopted: true},
		{Name: "content-address", Category: "go",
			Description: "Every operation has a SHA256 Ref",
			Example: "Ref([]byte(content))[:12]", Adopted: true},
		{Name: "pulse-driven", Category: "systems",
			Description: "Every state change emits a Pulse for telemetry",
			Example: "Pulse(\"block.write\", path)", Adopted: true},
		{Name: "sacred-surface", Category: "vaked",
			Description: "Every UI element flows through ASCIIBox",
			Example: "ASCIIBox(title, lines, 52)", Adopted: true},
		{Name: "vaked-fit", Category: "vaked",
			Description: "Every primitive declares its Vaked layer",
			Example: "func XxxVakedFit() string", Adopted: true},
	},
}

// ── Session Upgrade Graph ─────────────────────────────────────────────

// SessionUpgrade tracks TUI → SSH → Graph upgrade path.
type SessionUpgrade struct {
	From    string // "tui-local", "ssh", "websocket"
	To      string // "ssh", "graph", "minecraft"
	Status  string // "available", "active", "planned"
	Latency time.Duration
}

var sessionUpgrades = []SessionUpgrade{
	{From: "tui-local", To: "ssh", Status: "available", Latency: 0},
	{From: "ssh", To: "websocket", Status: "available", Latency: 0},
	{From: "websocket", To: "graph", Status: "planned", Latency: 0},
	{From: "graph", To: "minecraft", Status: "planned", Latency: 0},
}

// SessionUpgradeGraph renders the upgrade graph as ASCII.
func SessionUpgradeGraph() string {
	var sb strings.Builder
	sb.WriteString("╔══ SESSION UPGRADE GRAPH ══╗\n")
	
	for _, u := range sessionUpgrades {
		icon := "✅"
		if u.Status == "planned" { icon = "🔜" }
		sb.WriteString(fmt.Sprintf("║ %s %s → %s\n", icon, u.From, u.To))
	}
	
	sb.WriteString("╠══════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║ Active: %s\n", GetUIMode()))
	sb.WriteString(fmt.Sprintf("║ Session: %s\n", LiveSessionStatus()))
	sb.WriteString("╚══════════════════════════╝")
	
	return sb.String()
}

// BestPracticesReport returns the best practices report.
func BestPracticesReport() string {
	var sb strings.Builder
	adopted := 0
	for _, p := range bestPractices.Practices {
		if p.Adopted { adopted++ }
	}
	
	sb.WriteString(fmt.Sprintf("╔══ BEST PRACTICES — %d/%d adopted ══╗\n", adopted, len(bestPractices.Practices)))
	
	for _, p := range bestPractices.Practices {
		icon := "✅"
		if !p.Adopted { icon = "🟡" }
		sb.WriteString(fmt.Sprintf("║ %s %s: %s\n", icon, p.Name, p.Description))
	}
	
	sb.WriteString(fmt.Sprintf("╠══════════════════════════════════════╣\n"))
	sb.WriteString(fmt.Sprintf("║ Go:      %s\n", runtime.Version()))
	sb.WriteString(fmt.Sprintf("║ Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("║ Blocks:  %d\n", len(schemaRegistry)))
	sb.WriteString("╚══════════════════════════════════════╝")
	
	return sb.String()
}
