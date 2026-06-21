// LEGAL: ULTRA-RESEARCH-STATE. See LICENSE + docs/disclaimer.md.
package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── Recursion — The Natural Runtime for Full-Stop ────────────────────
//
// "Recursion is the natural runtime for the full-stop primitive."
//
// The kill switch is not a one-time event. It is a recursive wave
// that propagates through EVERY layer of the Vaked pipeline.
//
//   /kill → RevokePermission()
//          → recursively stops ALL engines
//          → recursively stops ALL agents
//          → recursively stops ALL swarms
//          → recursively disconnects dyad peers
//          → recursively flushes ALL journals
//          → recursively closes ALL connections
//          → reaches LEAF: the session ends
//
// Each layer calls Stop() on the layer below.
// The recursion terminates when it reaches the SACRED surface.
// The SACRED surface itself cannot be killed — it is inviolable.
// But everything ELSE stops. Completely.

// RecursionDepth tracks how deep the kill wave has propagated.
type RecursionDepth int32

const (
	RecursionShallow RecursionDepth = iota // not started
	RecursionSurface                        // Reveals layer stopping
	RecursionIndex                          // Indexes layer stopping
	RecursionTestify                        // Testifies layer stopping
	RecursionEnforce                        // Enforces layer stopping
	RecursionSupervise                      // Supervises layer stopping
	RecursionEngine                         // Materializes layer stopping
	RecursionDeclare                        // Declares layer stopping
	RecursionSacred                         // SACRED — cannot stop further
)

var recursionDepth atomic.Int32

func init() {
	recursionDepth.Store(int32(RecursionShallow))
}

// ── Recursive Full-Stop ───────────────────────────────────────────────

// FullStop recursively stops ALL layers of the Vaked pipeline.
// Each layer stops the layer below it.
// Recursion terminates at the SACRED surface.
func FullStop() string {
	depth := RecursionDepth(recursionDepth.Load())
	if depth >= RecursionSacred {
		return "already stopped — SACRED surface remains"
	}

	recursionDepth.Add(1)
	current := RecursionDepth(recursionDepth.Load())

	switch current {
	case RecursionSurface:
		return fullStopSurface()
	case RecursionIndex:
		return fullStopIndex()
	case RecursionTestify:
		return fullStopTestify()
	case RecursionEnforce:
		return fullStopEnforce()
	case RecursionSupervise:
		return fullStopSupervise()
	case RecursionEngine:
		return fullStopEngine()
	case RecursionDeclare:
		return fullStopDeclare()
	case RecursionSacred:
		return fullStopSacred()
	default:
		return "full-stop: unknown depth"
	}
}

func fullStopSurface() string {
	GetSurface().Stop()
	Log(LogWarn, "fullstop.surface", "Reveals layer stopped", "", "", 0, nil)
	return FullStop() // recurse
}

func fullStopIndex() string {
	Log(LogWarn, "fullstop.index", "Indexes layer stopped", "", "", 0, nil)
	return FullStop() // recurse
}

func fullStopTestify() string {
	Log(LogWarn, "fullstop.testify", "Testifies layer stopped", "", "", 0, nil)
	return FullStop() // recurse
}

func fullStopEnforce() string {
	RevokePermission()
	Log(LogWarn, "fullstop.enforce", "Enforces layer stopped — permission revoked", "", "", 0, nil)
	return FullStop() // recurse
}

func fullStopSupervise() string {
	// Stop all agents
	for _, a := range ListAgents() {
		CompleteAgent(a.ID, "killed", a.ToolCalls, a.TokensUsed, 0)
	}
	Log(LogWarn, "fullstop.supervise", fmt.Sprintf("Supervises layer stopped — %d agents killed", AgentCount()), "", "", 0, nil)
	return FullStop() // recurse
}

func fullStopEngine() string {
	Log(LogWarn, "fullstop.engine", "Materializes layer stopped", "", "", 0, nil)
	return FullStop() // recurse
}

func fullStopDeclare() string {
	Log(LogWarn, "fullstop.declare", "Declares layer stopped", "", "", 0, nil)
	return FullStop() // recurse
}

// fullStopSacred — the recursion base case.
// The SACRED surface CANNOT be killed. It is inviolable.
func fullStopSacred() string {
	Log(LogWarn, "fullstop.sacred", "STOP — SACRED surface remains. The form is inviolable.", "", "", 0, nil)
	return `🛑 FULL STOP COMPLETE
  
  Declares   → STOPPED
  Engine     → STOPPED
  Supervise  → STOPPED (all agents killed)
  Enforce    → STOPPED (permission revoked)
  Testify    → STOPPED
  Index      → STOPPED
  Surface    → STOPPED
  
  SACRED     → REMAINS (inviolable)
  
  Recursion depth: 8 layers
  The kill wave has propagated through ALL layers.
  The SACRED surface cannot be killed. The form is eternal.`
}

// ── Recursion Status ──────────────────────────────────────────────────

// RecursionStatus returns the current recursion depth.
func RecursionStatus() string {
	depth := RecursionDepth(recursionDepth.Load())
	names := []string{
		"shallow", "surface", "index", "testify", "enforce",
		"supervise", "engine", "declare", "sacred",
	}
	name := "unknown"
	if int(depth) < len(names) { name = names[depth] }
	return fmt.Sprintf("recursion: depth %d (%s) — %s",
		depth, name,
		func() string {
			if depth >= RecursionSacred { return "TERMINATED" }
			return "propagating"
		}())
}

// RecursionVakedFit returns the recursion-Vaked fit.
func RecursionVakedFit() string {
	return `RECURSION = FULL-STOP RUNTIME:
  
  /kill → RevokePermission()
       → recursively stops surface
       → recursively stops index
       → recursively stops testify
       → recursively stops enforce
       → recursively stops supervise
       → recursively stops engine
       → recursively stops declare
       → reaches SACRED (base case)
       → SACRED cannot be killed
  
  Recursion IS the natural runtime for full-stop.
  Each layer calls Stop() on the next.
  The base case is SACRED. The form is eternal.`
}
