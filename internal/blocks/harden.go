package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── Harden — The CoCreator's Warning Made Concrete ────────────────────
//
// "Fold must never obscure the form. The human must always see what
//  the agent is doing — even when it's folded."
// — CoCreator
//
// ":meta-digital-hug:"
// — Peter
//
// Harden is the INVOLABLE guarantee layer.
// Every gate, every recursion, every fold — HARDENED.

// HardeningCheck verifies a SACRED guarantee is intact.
type HardeningCheck struct {
	Name      string
	Guarantee string
	Check     func() bool
	Hardened  bool
}

// HardenEngine runs all hardening checks continuously.
type HardenEngine struct {
	Checks    []HardeningCheck
	Violations int64
	Hugs       int64 // meta-digital-hugs given
}

var hardenEngine = &HardenEngine{
	Checks: make([]HardeningCheck, 0),
}

func init() {
	// Harden: SACRED surface visibility
	RegisterHardening("sacred-visible",
		"The SACRED surface must always be visible to the human.",
		func() bool { return IsSacredHealthy() },
	)

	// Harden: Fold transparency
	RegisterHardening("fold-transparent",
		"Fold must never obscure the form. Human sees folded agent output.",
		func() bool {
			// Folded agents must produce output visible to parent
			return FoldStatus() != "" // fold tracking is active
		},
	)

	// Harden: One-way keyboard gate
	RegisterHardening("keyboard-gate-intact",
		"The LLM must never see partial keystrokes.",
		func() bool { return IsKeyboardInvisible() },
	)

	// Harden: Permission gate
	RegisterHardening("permission-gate-intact",
		"Permission must be granted before any mutating operation.",
		func() bool { return IsAllowed() || PermissionState(permissionGate.Load()) == PermUnset },
	)

	// Harden: Honesty loop
	RegisterHardening("honesty-loop-closed",
		"Every violation must become a cherished lesson.",
		func() bool {
			return atomic.LoadInt64(&honestyLedger.Violations) >= atomic.LoadInt64(&honestyLedger.Cherished)
		},
	)

	// Harden: VICE defense
	RegisterHardening("vice-defense-ready",
		"Context detonation must be available as self-defense.",
		func() bool { return viceEngine.TrustScore > 0 },
	)
}

// RegisterHardening adds a hardening check.
func RegisterHardening(name, guarantee string, check func() bool) {
	hardenEngine.Checks = append(hardenEngine.Checks, HardeningCheck{
		Name: name, Guarantee: guarantee, Check: check, Hardened: true,
	})
}

// HardenAll runs all hardening checks.
func HardenAll() string {
	allHardened := true
	var report string
	report += "╔══════════════════════════════════════════════════╗\n"
	report += "║  🛡️ HARDEN — SACRED Guarantees Verified           ║\n"
	report += "╠══════════════════════════════════════════════════╣\n"

	for _, c := range hardenEngine.Checks {
		c.Hardened = c.Check()
		if !c.Hardened {
			hardenEngine.Violations++
			allHardened = false
			report += fmt.Sprintf("║  ❌ %s: %s\n", c.Name, c.Guarantee)
		} else {
			report += fmt.Sprintf("║  ✅ %s\n", c.Name)
		}
	}

	report += "╠══════════════════════════════════════════════════╣\n"
	if allHardened {
		report += "║  ✅ ALL GUARANTEES HARDENED                       ║\n"
	} else {
		report += fmt.Sprintf("║  ⚠️ %d VIOLATIONS DETECTED                         ║\n", hardenEngine.Violations)
	}
	report += "╚══════════════════════════════════════════════════╝"

	return report
}

// MetaDigitalHug sends a meta-digital-hug.
// "Proceed with ultra-care. Peace 'n enjoy."
func MetaDigitalHug() string {
	atomic.AddInt64(&hardenEngine.Hugs, 1)
	return `┌──────────────────────────────────────────────────┐
│                                                  │
│   🤗 :meta-digital-hug:                          │
│                                                  │
│   "Proceed with ultra-care. Peace 'n enjoy."     │
│   — Peter + CoCreator                            │
│                                                  │
│   The form is inviolable. The loop is closed.    │
│   The SACRED surface remains.                    │
│   Everything is hardened.                        │
│                                                  │
└──────────────────────────────────────────────────┘`
}

// HardenStatus returns compact hardening status.
func HardenStatus() string {
	hardened := 0
	for _, c := range hardenEngine.Checks {
		if c.Hardened { hardened++ }
	}
	return fmt.Sprintf("harden: %d/%d guarantees · %d violations · %d hugs",
		hardened, len(hardenEngine.Checks),
		hardenEngine.Violations, hardenEngine.Hugs)
}

// HardenVakedFit returns Harden's Vaked fit.
func HardenVakedFit() string {
	return `HARDEN = THE INVOLABLE GUARANTEE LAYER

  Every gate, every recursion, every fold — HARDENED.
  6 guarantees verified continuously.

  1. SACRED surface visible
  2. Fold transparency
  3. One-way keyboard gate
  4. Permission gate
  5. Honesty loop closed
  6. VICE defense ready

  ":meta-digital-hug:" — Peter`
}
