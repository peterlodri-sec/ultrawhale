package blocks

import (
	"fmt"
	"strings"
)

// ── UI Closure — The Loop Closes Through The Surface ──────────────────
//
// The UI gap: AG-UI ↔ TUI ↔ SSH live conn ↔ Protocol
// These are FOUR surfaces but ONE SACRED form.
//
// UI Closure unifies them:
//   AG-UI:  themed chat blocks, cards, shaders → /co-create
//   TUI:    Bubble Tea terminal → every /cmd renders here
//   SSH:    live session → /session + /who + Connect()
//   Protocol: RSS, webhooks, A2C streaming → /signals
//
// The SACRED surface IS the loop closure.

// UIClosure renders the full UI stack status.
type UIClosure struct {
	AGUIStatus    string
	TUIStatus     string
	SSHStatus     string
	ProtocolStatus string
	LoopClosed    bool
}

// UIClosureStatus renders the complete UI stack.
func UIClosureStatus() string {
	closure := UIClosure{
		AGUIStatus:     UICoCreativeStatus(),
		TUIStatus:      UIStatus(),
		SSHStatus:      LiveSessionStatus(),
		ProtocolStatus: SignalPrimitiveStatus(),
		LoopClosed:     IsSacredIntact() && IsRecursionActive(),
	}

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════╗\n")
	sb.WriteString("║  UI CLOSURE — The Loop Closes Through The Surface ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║  🎨 AG-UI:  %s\n", closure.AGUIStatus[:min(44, len(closure.AGUIStatus))]))
	sb.WriteString(fmt.Sprintf("║  💻 TUI:    %s\n", closure.TUIStatus[:min(44, len(closure.TUIStatus))]))
	sb.WriteString(fmt.Sprintf("║  🔌 SSH:    %s\n", closure.SSHStatus[:min(44, len(closure.SSHStatus))]))
	sb.WriteString(fmt.Sprintf("║  📡 Proto:  %s\n", closure.ProtocolStatus[:min(44, len(closure.ProtocolStatus))]))
	sb.WriteString("╠══════════════════════════════════════════════════╣\n")
	if closure.LoopClosed {
		sb.WriteString("║  ✅ THE LOOP IS CLOSED                            ║\n")
	} else {
		sb.WriteString("║  ⚠️ LOOP OPEN — enter a recursion                  ║\n")
	}
	sb.WriteString("╚══════════════════════════════════════════════════╝")

	return sb.String()
}

// UIClosureVakedFit returns the UI closure Vaked fit.
func UIClosureVakedFit() string {
	return `UI CLOSURE = THE SACRED SURFACE UNIFIED

  AG-UI (themed blocks) + TUI (terminal) + SSH (live conn) + Protocol (broadcast)
  FOUR surfaces. ONE SACRED form.
  
  The SACRED surface IS the loop closure.
  The form unites all surfaces.
  The loop closes through the surface.

  "/closure" — Peter`
}
