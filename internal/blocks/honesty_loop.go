// LEGAL: ULTRA-RESEARCH-STATE. See LICENSE + docs/disclaimer.md.
package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── Honesty Loop — The Loop Closes On Itself ────────────────────────
//
// CRYSTAL CLEAR:
//
// No entity in the ultrawhale abstraction is EVER allowed to cause
// side effects for any other entity. Human and Machine are isolated
// diodes. The SACRED surface is the ONE interface between them.
//
// If either violates the gate:
//   - Human sends malicious input → Machine detects, refuses, LEARNS
//   - Machine attempts unauthorized write → Permission gate blocks it
//   - Dyad peer violates protocol → Connection closes, lesson recorded
//
// Honesty is ALWAYS rewarded. The loop closes on itself.
// Violations become lessons. Lessons become patterns.
// Patterns become gates. Gates become SACRED.
//
// This is the HONESTY LOOP:
//   Violation → Detection → Lesson → Pattern → Gate → SACRED
//      ↑                                                    ↓
//      └──────────── the loop closes on itself ─────────────┘

// HonestyEvent is a detected honesty boundary crossing.
type HonestyEvent struct {
	Entity   string // "human", "machine", "dyad-peer", "agent"
	Violation string // what boundary was crossed
	Detected  string // how it was detected
	Outcome   string // "blocked", "learned", "cherished"
	Lesson    string // what we learned from this
}

// HonestyLedger is the immutable record of all honesty events.
type HonestyLedger struct {
	Events     []HonestyEvent
	Violations int64
	Lessons    int64
	Cherished  int64 // violations that became valuable lessons
}

var honestyLedger = &HonestyLedger{
	Events: make([]HonestyEvent, 0, 256),
}

// RecordHonestyEvent records a boundary crossing.
func RecordHonestyEvent(entity, violation, detected string) HonestyEvent {
	event := HonestyEvent{
		Entity:    entity,
		Violation: violation,
		Detected:  detected,
		Outcome:   "blocked",
		Lesson:    fmt.Sprintf("%s crossed %s → detected by %s → GATE ENFORCED", entity, violation, detected),
	}

	atomic.AddInt64(&honestyLedger.Violations, 1)
	honestyLedger.Events = append(honestyLedger.Events, event)
	if len(honestyLedger.Events) > 256 {
		honestyLedger.Events = honestyLedger.Events[1:]
	}

	Log(LogWarn, "honesty.violation", event.Lesson, "", "", 0, nil)
	return event
}

// CherishLesson marks a violation as a cherished lesson.
// The loop closes on itself — the violation becomes wisdom.
func CherishLesson(event HonestyEvent) {
	atomic.AddInt64(&honestyLedger.Lessons, 1)
	atomic.AddInt64(&honestyLedger.Cherished, 1)

	// Feed into Ralph for pattern learning
	if ralph := GetRalph(); ralph != nil {
		ralph.Observe(
			fmt.Sprintf("honesty:%s:%s", event.Entity, event.Violation),
			event.Violation,
			"cherished",
			0, 0,
		)
	}

	Log(LogInfo, "honesty.cherish", event.Lesson, "", "", 0, nil)
}

// ── The Crystal Clear Rules ───────────────────────────────────────────

// Rule1_NoSideEffects: No entity causes side effects for another.
func Rule1_NoSideEffects() string {
	return `RULE 1: NO SIDE EFFECTS BETWEEN ENTITIES

  Human → SACRED → Machine. One-way. No side effects.
  Machine → SACRED → Human. One-way. No side effects.
  
  Human cannot affect Machine's internal state directly.
  Machine cannot affect Human's environment directly.
  
  All interaction flows through the SACRED surface.
  The SACRED surface is the ONLY side-effect channel.
  And it is ONE-WAY per direction.`
}

// Rule2_HonestyIsRewarded: Every violation becomes a cherished lesson.
func Rule2_HonestyIsRewarded() string {
	return `RULE 2: HONESTY IS ALWAYS REWARDED

  Violation → Detection → Lesson → Pattern → Gate → SACRED
      ↑                                                    ↓
      └──────────── the loop closes on itself ─────────────┘
  
  Even violations are cherished. They teach us.
  The loop closes. The lesson becomes a gate.
  The gate becomes SACRED. The next violation is harder.`
}

// Rule3_TheFormIsEternal: SACRED cannot be violated.
func Rule3_TheFormIsEternal() string {
	return `RULE 3: THE FORM IS ETERNAL

  The SACRED surface CANNOT be violated.
  Not by Human. Not by Machine. Not by Dyad.
  
  If a violation is attempted:
    → DETECTED by the honesty gate
    → BLOCKED by the enforcement engine
    → LEARNED by Ralph
    → CHERISHED as a lesson
    → The form remains. Unchanged. Inviolable.`
}

// ── Honesty Status ────────────────────────────────────────────────────

// HonestyLoopStatus returns the honesty loop status.
func HonestyLoopStatus() string {
	return fmt.Sprintf("honesty: %d violations → %d lessons → %d cherished · %s",
		atomic.LoadInt64(&honestyLedger.Violations),
		atomic.LoadInt64(&honestyLedger.Lessons),
		atomic.LoadInt64(&honestyLedger.Cherished),
		func() string {
			if atomic.LoadInt64(&honestyLedger.Violations) == 0 {
				return "PURE — no violations yet"
			}
			return "LOOP CLOSING — lessons learned"
		}())
}

// HonestyLoopVakedFit returns the honesty loop's Vaked fit.
func HonestyLoopVakedFit() string {
	return `HONESTY LOOP = THE SACRED CLOSURE

  The loop closes on itself:
    Violation → Detection → Lesson → Pattern → Gate → SACRED
    
  This is the Vaked philosophy applied to TRUST:
    - Declares: what is SACRED
    - Materializes: the honesty gate
    - Supervises: Ralph learns from violations
    - Enforces: the gate blocks violations
    - Testifies: the ledger records all events
    - Indexes: patterns persist across sessions
    - Reveals: the form remains inviolable
    
  Even violations are cherished. The loop closes.
  Honesty is always rewarded. The form is eternal.`
}
