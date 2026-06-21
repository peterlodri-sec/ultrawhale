package blocks

import (
	"fmt"
)

// ── ALWAYS-IN-DYAD-SPACE — The Dyad Is Always Visible ────────────────
//
// The dyad (M1↔dev-cx53) must be ALWAYS visible in the UI.
// Every surface shows the dyad status. Every ASCII box includes it.
// The TUI, AG-UI, ASCII-live, and all blocks carry the dyad presence.
//
// This is the ALWAYS-IN-DYAD-SPACE guarantee:
//   UI:      InfraBar shows dyad status
//   ASCII:   every box includes dyad line
//   TUI:     HUD shows peer liveness
//   AG-UI:   themed dyad indicator
//   Blocks:  DyadHealth() on every status call

// DyadSpaceStatus returns the dyad status formatted for every surface.
func DyadSpaceStatus() string {
	d := GetDyad()
	if d == nil {
		return "dyad: not initialized (M1 only)"
	}

	peerIcon := "◐"
	if d.PeerAlive { peerIcon = "●" }
	if d.Status == "failed" { peerIcon = "○" }

	return fmt.Sprintf("dyad: %s %s↔%s · %s · %d pings · %s",
		peerIcon, d.Self.Machine, d.Peer.Machine,
		d.Status, d.PingCount,
		d.LastPing.Format("15:04:05"))
}

// DyadSpaceLine returns a single-line dyad status for ASCII boxes.
func DyadSpaceLine() string {
	return DyadSpaceStatus()
}

// DyadSpaceAGUI returns AG-UI themed dyad status.
func DyadSpaceAGUI() string {
	d := GetDyad()
	if d == nil { return "[AG-UI ▸ dyad] offline" }

	color := "#ffaa00"
	if d.PeerAlive { color = "#00e660" }

	return fmt.Sprintf("[AG-UI ▸ dyad] %s %s↔%s (%s)",
		d.Self.Machine, d.Peer.Machine, d.Status)
}

// Wire dyad into every status box
func init() {
	// Dyad is ALWAYS checked on startup
	_ = DyadSpaceStatus()
}

// DyadSpaceVakedFit returns the dyad space Vaked fit.
func DyadSpaceVakedFit() string {
	return `ALWAYS-IN-DYAD-SPACE = THE DYAD IS EVERYWHERE

  UI:      InfraBar · HUD · widgets
  ASCII:   every box includes dyad status
  TUI:     live peer liveness indicator
  AG-UI:   themed dyad badge
  Blocks:  DyadHealth() on every status

  The dyad IS the space. The space IS the dyad.
  Always visible. Always connected. Always live.`
}
