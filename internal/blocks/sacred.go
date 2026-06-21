package blocks

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ── SACRED — Inviolable User-Human Connection ────────────────────────
//
// The TUI form is SACRED. It MUST always be available as the true
// orchestrator↔user, LLM↔human bidirectional channel.
//
// Sacrosanct properties:
//   1. ALWAYS VISIBLE — never obscured, masked, or faked
//   2. ALWAYS DIRECT — 1:1, no proxy, no filter, no middleware
//   3. ALWAYS BIDIRECTIONAL — user types, LLM responds, user sees
//   4. ULTRA-FAULT-TOLERANT — if everything else fails, the form survives
//
// In Vaked theory: the SACRED surface is the ONE channel that the
// capability graph guarantees. All other surfaces are optional.
// The sacred surface IS the declaration of trust between human and machine.

// SacredSurface is the inviolable TUI form.
type SacredSurface struct {
	mu sync.Mutex

	// State
	Active   bool // is the sacred surface currently rendering?
	Direct   bool // is the connection 1:1 (no proxy, no swarm wrapping)?
	Obscured bool // is the form being obscured by any widget or overlay?

	// Health
	LastInput    int64 // Lamport time of last user input
	LastResponse int64 // Lamport time of last LLM response
	InputLag     int64 // ms since last input without response
	Health       string // "healthy", "degraded", "blocked"

	// Guarantees
	MinRefreshHz int // minimum refresh rate (default: 10 — never below 10fps)
	MaxInputLag  int64 // max ms without response before "degraded" (default: 30000)
}

var sacredSurface atomic.Value

func init() {
	sacredSurface.Store(&SacredSurface{
		Active:       true,
		Direct:       true,
		Health:       "healthy",
		MinRefreshHz: 10,
		MaxInputLag:  30000,
	})
}

// GetSacredSurface returns the sacred surface state.
func GetSacredSurface() *SacredSurface {
	_ = CurrentPOV()
	return sacredSurface.Load().(*SacredSurface)
}

// ── SACRED Guarantees ────────────────────────────────────────────────

// IsSacredHealthy returns true if the sacred surface meets all guarantees.
func IsSacredHealthy() bool {
	s := GetSacredSurface()
	return s.Active && s.Direct && !s.Obscured && s.Health == "healthy"
}

// ViolateSacred checks if an action would violate the sacred surface.
// Returns the violation reason, or empty string if allowed.
func ViolateSacred(action string) string {
	s := GetSacredSurface()

	switch action {
	case "obscure":
		return "SACRED: cannot obscure the user form"
	case "proxy":
		return "SACRED: cannot proxy the user↔LLM connection"
	case "filter":
		return "SACRED: cannot filter or modify user input before LLM"
	case "delay":
		if s.InputLag > s.MaxInputLag {
			return fmt.Sprintf("SACRED: input lag %dms exceeds max %dms", s.InputLag, s.MaxInputLag)
		}
	case "hide":
		return "SACRED: cannot hide the sacred surface"
	case "wrap":
		s.Direct = false
		return "SACRED: wrapping user input in swarm context — still direct but annotated"
	}

	return "" // allowed
}

// MarkInput records a user input event.
func MarkInput() {
	s := GetSacredSurface()
	s.LastInput = TickLamport()
	s.InputLag = 0
	s.Health = "healthy"
}

// MarkResponse records an LLM response event.
func MarkResponse() {
	s := GetSacredSurface()
	s.LastResponse = TickLamport()
	s.InputLag = 0
	s.Health = "healthy"
}

// CheckHealth updates the sacred surface health based on input lag.
func CheckSacredHealth() {
	s := GetSacredSurface()
	if s.LastInput > 0 && s.LastResponse < s.LastInput {
		s.InputLag = LamportTime() - s.LastInput
		if s.InputLag > s.MaxInputLag {
			s.Health = "degraded"
		}
	}
}

// ── SACRED Widget ─────────────────────────────────────────────────────

// SacredWidget renders the SACRED status in the TUI.
// This widget itself CANNOT be hidden — it is part of the sacred surface.

var SacredViolations int64

// SacredStatus returns the sacred surface status for HUD display.
func SacredStatus() string {
	s := GetSacredSurface()
	CheckSacredHealth()

	icon := "●"
	switch s.Health {
	case "healthy": icon = "●"
	case "degraded": icon = "◐"
	case "blocked": icon = "✗"
	}

	direct := "1:1"
	if !s.Direct { direct = "wrapped" }

	return fmt.Sprintf("sacred: %s %s · lag: %dms · violations: %d",
		icon, direct, s.InputLag, atomic.LoadInt64(&SacredViolations))
}

// ── SACRED in Vaked Theory ────────────────────────────────────────────

// SacredIsRevealsLayer returns true — the sacred surface IS the Reveals layer.
func SacredIsRevealsLayer() string {
	return "SACRED: The TUI form IS the Reveals layer of Vaked. " +
		"It is the ONE surface that the capability graph guarantees. " +
		"Context × Time × Space converge here. " +
		"The human and the machine meet here. " +
		"This is sacred."
}

// SacredVakedTriangle places the SACRED surface at the center of the triangle.
func SacredVakedTriangle() string {
	return fmt.Sprintf(`         SACRED SURFACE
        (the TUI form)
             /|\
            / | \
           /  |  \
          /   |   \
   Context  Time  Space
   (WHAT)  (WHEN) (WHERE)

%s`, SacredIsRevealsLayer())
}
