package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── SELF_MAIN_STATE — The 4 Fundamental States ────────────────────────
//
// Every system has a main state. The SACRED event loop transitions
// between 4 states:
//
//   UNKNOWN — booting, initializing, "what am I?"
//   DREAM   — disconnected, offline, "connection error, but I'm still here"
//   HERE    — connected, alive, rendering, "the human is present"
//   LIVE    — broadcasting, streaming, "the radio is on, HF webhook active"
//
// These are the SELF_MAIN_STATE. The event loop checks this every frame.

// MainState is the fundamental system state.
type MainState int32

const (
	StateUnknown MainState = iota // booting, initializing
	StateDream                     // disconnected, offline
	StateHere                      // connected, alive
	StateLive                      // broadcasting, streaming
)

var mainState atomic.Int32

func init() { mainState.Store(int32(StateUnknown)) }

// ── State Transitions ────────────────────────────────────────────────

// SetMainState transitions the system to a new main state.
func SetMainState(state MainState) string {
	old := MainState(mainState.Swap(int32(state)))
	Log(LogInfo, "state.transition", fmt.Sprintf("%s → %s", stateName(old), stateName(state)),
		"", "", 0, nil)
	Pulse("state.change", fmt.Sprintf("%s→%s", stateName(old), stateName(state)))
	return fmt.Sprintf("state: %s → %s", stateName(old), stateName(state))
}

// GetMainState returns the current main state.
func GetMainState() MainState { return MainState(mainState.Load()) }

// ── State-Aware Operations ───────────────────────────────────────────

// CanRender returns true if the SACRED surface can render.
func CanRender() bool {
	s := GetMainState()
	return s == StateHere || s == StateLive
}

// CanBroadcast returns true if we can stream/broadcast.
func CanBroadcast() bool {
	return GetMainState() == StateLive
}

// DreamReason returns why we're dreaming.
func DreamReason() string {
	if GetMainState() != StateDream { return "not dreaming" }
	
	// Check what's missing
	reasons := []string{}
	if !IsSacredHealthy() { reasons = append(reasons, "sacred degraded") }
	if GetDyad() == nil || !GetDyad().PeerAlive { reasons = append(reasons, "no dyad peer") }
	if AgentCount() == 0 { reasons = append(reasons, "no agents") }
	
	if len(reasons) == 0 { return "dreaming: unknown reason" }
	return "dreaming: " + fmt.Sprintf("%v", reasons)
}

// ── Auto-Transition Logic ─────────────────────────────────────────────

// AutoTransition detects the current system state and sets it.
func AutoTransition() string {
	state := GetMainState()

	// UNKNOWN → HERE (first connection)
	if state == StateUnknown && IsSacredHealthy() {
		return SetMainState(StateHere)
	}

	// HERE → LIVE (broadcasting)
	if state == StateHere && CanBroadcast() {
		return SetMainState(StateLive)
	}

	// HERE → DREAM (disconnected)
	if state == StateHere && !IsSacredHealthy() {
		return SetMainState(StateDream)
	}

	// DREAM → HERE (reconnected)
	if state == StateDream && IsSacredHealthy() {
		return SetMainState(StateHere)
	}

	return fmt.Sprintf("state: %s (stable)", stateName(state))
}

// ── State Status ──────────────────────────────────────────────────────

func stateName(s MainState) string {
	switch s {
	case StateUnknown: return "UNKNOWN"
	case StateDream: return "DREAM"
	case StateHere: return "HERE/LIVE"
	case StateLive: return "LIVE"
	default: return "???"
	}
}

// SelfMainStateStatus returns compact state status.
func SelfMainStateStatus() string {
	state := GetMainState()
	result := fmt.Sprintf("state: %s", stateName(state))
	if state == StateDream { result += " · " + DreamReason() }
	if state == StateLive { result += " · broadcasting" }
	return result
}

// SelfMainStateVakedFit returns the state's Vaked fit.
func SelfMainStateVakedFit() string {
	return `SELF_MAIN_STATE = THE 4 SACRED STATES

  UNKNOWN → DREAM → HERE/LIVE → LIVE → DREAM → HERE → ...
  
  UNKNOWN: booting, "what am I?"
  DREAM:   connection error, "but I'm still here"
  HERE:    human present, rendering the SACRED surface
  LIVE:    broadcasting to HuggingFace, webhook active, radio on

  The event loop checks this every frame.
  Auto-transition based on SACRED health, dyad state, agent count.

  "DREAM is connection error." — Peter`
}
