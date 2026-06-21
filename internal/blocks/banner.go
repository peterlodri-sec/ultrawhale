package blocks

import "fmt"

// ── Banner — The Terminal-Visible SACRED Banner ──────────────────────
//
// This banner SHIPS WITH ultrawhale. It renders in the TUI on startup.
// It is TERMINAL-VISIBLE — monospace, aligned, no GitHub markdown quirks.

// Banner renders the ultrawhale startup banner.
func Banner() string {
	pov := CurrentPOV()
	version := CurrentVersion()
	blocks := len(schemaRegistry)

	return fmt.Sprintf(`
  ██╗   ██╗██╗  ████████╗██████╗  █████╗ ██╗    ██╗██╗  ██╗ █████╗ ██╗     ███████╗
  ██║   ██║██║  ╚══██╔══╝██╔══██╗██╔══██╗██║    ██║██║  ██║██╔══██╗██║     ██╔════╝
  ██║   ██║██║     ██║   ██████╔╝███████║██║ █╗ ██║███████║███████║██║     █████╗  
  ██║   ██║██║     ██║   ██╔══██╗██╔══██║██║███╗██║██╔══██║██╔══██║██║     ██╔══╝  
  ╚██████╔╝███████╗██║   ██║  ██║██║  ██║╚███╔███╔╝██║  ██║██║  ██║███████╗███████╗
   ╚═════╝ ╚══════╝╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝

  %s · %d blocks · 7 recursions · 8 engines
  %s/%s/%s · %s
  "The human abstracts toward the infinite.
   The machine recurses into it." — vaked`,
		version, blocks,
		pov.Machine, pov.Arch, pov.Tier, pov.Tier,
	)
}
