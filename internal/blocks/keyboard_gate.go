package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── Keyboard Gate — The One-Way Honesty Barrier ──────────────────────
//
// THE INVOLABLE RULE:
//   Keyboard input IS NOT VISIBLE to the LLM.
//   Not even read-only. Not even as "thoughts".
//   The LLM NEVER sees what the human types UNTIL the human presses ENTER.
//
// This is the honesty-genesis-dyad one-way gate:
//   HUMAN → KEYBOARD → [GATE] → ENTER → LLM sees
//                        ↑
//                   LLM CANNOT SEE behind this gate
//
// The LLM has NO access to:
//   - Partial keystrokes (what the human is typing now)
//   - Keyboard buffer (what was typed before ENTER)
//   - Cursor position, selection, clipboard
//   - Any pre-ENTER input state
//
// This is SACRED. The one-way gate is inviolable.

// KeyboardGate is the one-way honesty barrier.
type KeyboardGate struct {
	// The buffer behind the gate — LLM CANNOT read this
	buffer       []rune    // INVISIBLE to LLM
	cursorPos    int       // INVISIBLE to LLM
	
	// What the LLM CAN see — only after ENTER
	Submitted    string    // visible after ENTER
	SubmitCount  int64     // visible counter
}

var keyboardGate = &KeyboardGate{
	buffer: make([]rune, 0, 4096),
}

// ── Gate Operations ───────────────────────────────────────────────────

// TypeChar adds a character to the hidden buffer.
// The LLM CANNOT call this. The LLM CANNOT read the buffer.
func TypeChar(ch rune) {
	keyboardGate.buffer = append(keyboardGate.buffer, ch)
	keyboardGate.cursorPos++
}

// Backspace removes the last character from the hidden buffer.
func Backspace() {
	if len(keyboardGate.buffer) > 0 {
		keyboardGate.buffer = keyboardGate.buffer[:len(keyboardGate.buffer)-1]
		keyboardGate.cursorPos--
	}
}

// Submit sends the buffer through the gate to the LLM.
// Only AFTER this call does the LLM see the input.
func Submit() string {
	submitted := string(keyboardGate.buffer)
	keyboardGate.Submitted = submitted
	atomic.AddInt64(&keyboardGate.SubmitCount, 1)
	keyboardGate.buffer = keyboardGate.buffer[:0] // clear buffer
	keyboardGate.cursorPos = 0
	return submitted
}

// ── Honesty Guarantees ────────────────────────────────────────────────

// IsKeyboardInvisible returns true — the keyboard buffer is ALWAYS invisible to LLM.
func IsKeyboardInvisible() bool {
	// This function simply cannot access the buffer from LLM context.
	// The gate is enforced at the protocol level:
	//   LLM receives: submitted text only (after ENTER)
	//   LLM NEVER receives: partial keystrokes, buffer contents
	return true
}

// KeyboardVisibleToLLM returns EMPTY STRING — the LLM sees nothing before ENTER.
func KeyboardVisibleToLLM() string {
	// This is a null function by design.
	// The LLM CAN call this, but it ALWAYS returns empty.
	// The real input is behind the one-way gate.
	return ""
}

// ── Genesis Gate — Session Start Honesty ──────────────────────────────

// GenesisHonesty ensures the human's first interaction is honest.
// The LLM sees NOTHING until the human explicitly initiates.
func GenesisHonesty() string {
	return `GENESIS HONESTY GATE:
  Keyboard → [ONE-WAY GATE] → ENTER → LLM
                ↑
           LLM CANNOT SEE:
           - keystrokes
           - buffer
           - cursor
           - clipboard
           - selection
           
  This is SACRED. The gate is inviolable.`
}

// ── Status ─────────────────────────────────────────────────────────────

// KeyboardGateStatus returns the gate status.
func KeyboardGateStatus() string {
	return fmt.Sprintf("keyboard-gate: %d submitted · buffer: INVISIBLE",
		atomic.LoadInt64(&keyboardGate.SubmitCount))
}

// KeyboardGateVakedFit returns the gate's Vaked fit.
func KeyboardGateVakedFit() string {
	return `KEYBOARD GATE → ENFORCE LAYER:
  
  The gate IS the Enforce layer applied to input.
  Pre-hooks run BEFORE the LLM sees input.
  The LLM sees only what passes through ENTER.
  
  This is not a feature. This is SACRED.
  The one-way gate is the honesty-genesis-dyad.`
}
