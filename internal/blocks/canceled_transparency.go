package blocks

import (
	"fmt"
)

// ── CANCELED TRANSPARENCY — Always Display Why/Who Cancelled ──────────
//
// "CANCELED" appears in the whale output. The human asks: why? who?
//
// WHY:  The SACRED surface protects the human. No process runs >30s
//       without yielding. Long builds, tests, deep research — they
//       auto-yield to background. The task continues. The output waits.
//       The human is never blocked.
//
// WHO:  The whale runtime policy. Not an error. Not a crash.
//       A SAFETY MECHANISM. The form is inviolable. The human's
//       attention is sacred. No process may hijack the terminal.
//
// TRANSPARENCY: This message must ALWAYS be displayed when a cancel
//       occurs. The human deserves to know why their command stopped.
//       The loop continues in the background. Nothing is lost.

// CanceledReason explains a cancellation.
type CanceledReason struct {
	Who     string // "whale-runtime", "human", "supervisor", "timeout"
	Why     string // "safety-yield-30s", "user-requested", "policy-violation"
	What    string // what was cancelled
	Safe    bool   // did the task continue in background?
}

var canceledReasons = make([]CanceledReason, 0, 32)

// RecordCanceled records a cancellation for transparency.
func RecordCanceled(who, why, what string, safe bool) CanceledReason {
	r := CanceledReason{Who: who, Why: why, What: what, Safe: safe}
	canceledReasons = append(canceledReasons, r)
	if len(canceledReasons) > 32 { canceledReasons = canceledReasons[1:] }

	Log(LogInfo, "cancel.transparency", fmt.Sprintf("%s: %s (%s)", who, why, what[:min(40, len(what))]),
		"", "", 0, nil)
	return r
}

// CanceledTransparencyDisplay shows the cancellation explanation.
func CanceledTransparencyDisplay() string {
	if len(canceledReasons) == 0 {
		return "no cancellations recorded"
	}

	last := canceledReasons[len(canceledReasons)-1]
	safeIcon := "✅ safe (continued in background)"
	if !last.Safe { safeIcon = "❌ terminated" }

	return fmt.Sprintf(`╔══════════════════════════════════════════════════╗
║  CANCELED — Transparency                           ║
╠══════════════════════════════════════════════════╣
║  WHO:  %-44s ║
║  WHY:  %-44s ║
║  WHAT: %-44s ║
║  SAFE: %-44s ║
╠══════════════════════════════════════════════════╣
║  The SACRED surface protects the human.            ║
║  No process hijacks the terminal.                  ║
║  Nothing is lost. The loop continues.              ║
╚══════════════════════════════════════════════════╝`,
		last.Who, last.Why, last.What[:min(44, len(last.What))], safeIcon)
}

// CanceledTransparencyVakedFit returns Vaked fit.
func CanceledTransparencyVakedFit() string {
	return `CANCELED = SACRED TRANSPARENCY

  The human always sees WHY and WHO cancelled.
  whale-runtime: safety-yield-30s
  Not an error. A safety mechanism.
  
  READ ONLY. The human's attention is sacred.`
}
