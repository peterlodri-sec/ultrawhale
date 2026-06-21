package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── Permission Gate — OneShot Honesty Protocol ────────────────────────
//
// Permission is asked ONCE per session. If granted honestly by the human
// at the SACRED surface, ALL subsequent operations are ALLOWED+AUTHED.
//
// Permission is revoked ONLY by:
//   1. Session refresh (new universe)
//   2. Kill switch: /STOP___KILL_SWITCH___DO_FULL_STOP
//   3. Dyad peer disconnect (if dual-verify requires both)
//
// This is the Vaked honesty gate. The human trusts the machine.
// The machine asks once. The human answers once.
// No nagging. No re-prompting. Sacred.

// PermissionState is the global permission gate.
type PermissionState int32

const (
	PermUnset    PermissionState = 0 // never asked
	PermGranted  PermissionState = 1 // human said yes
	PermDenied   PermissionState = 2 // human said no
	PermRevoked  PermissionState = 3 // kill switch activated
	PermExpired  PermissionState = 4 // session ended
)

var permissionGate atomic.Int32

func init() {
	permissionGate.Store(int32(PermUnset))
}

// ── Gate Operations ───────────────────────────────────────────────────

// AskPermission presents the honesty gate to the human.
// Called ONCE per session at the SACRED surface.
func AskPermission() PermissionState {
	current := PermissionState(permissionGate.Load())
	if current != PermUnset { return current }
	return PermUnset // waiting for human response
}

// GrantPermission marks the session as ALLOWED+AUTHED.
// Called when the human types "/allow" or confirms the dialog.
func GrantPermission() PermissionState {
	permissionGate.Store(int32(PermGranted))
	Log(LogInfo, "permission.grant", "session ALLOWED+AUTHED", "", "", 0, nil)
	return PermGranted
}

// DenyPermission marks the session as explicitly denied.
func DenyPermission() PermissionState {
	permissionGate.Store(int32(PermDenied))
	Log(LogWarn, "permission.deny", "session denied", "", "", 0, nil)
	return PermDenied
}

// RevokePermission activates the kill switch.
// ALL operations stop. The dyad disconnects. The session ends.
func RevokePermission() PermissionState {
	permissionGate.Store(int32(PermRevoked))
	
	// Kill switch: full stop
	if d := GetDyad(); d != nil {
		d.Status = "killed"
	}
	
	Log(LogWarn, "permission.revoke", "KILL SWITCH ACTIVATED — full stop", "", "", 0, nil)
	return PermRevoked
}

// ── Honesty Checks ────────────────────────────────────────────────────

// IsAllowed returns true if the current operation is permitted.
// This is the gate that EVERY mutating operation must pass through.
func IsAllowed() bool {
	state := PermissionState(permissionGate.Load())
	switch state {
	case PermGranted:
		return true // session ALLOWED+AUTHED
	case PermRevoked:
		return false // kill switch active
	case PermExpired:
		return false // session ended
	default:
		return false // never granted
	}
}

// IsSacredIntact returns true if the SACRED surface is healthy + permitted.
func IsSacredIntact() bool {
	return IsSacredHealthy() && IsAllowed()
}

// ── OneShot with Permission ───────────────────────────────────────────

// OneShotAllowed executes a OneShot ONLY if permission is granted.
func OneShotAllowed(declaration string) (OneShotResult, bool) {
	if !IsAllowed() {
		return OneShotResult{
			Declaration: declaration,
			Errors:      []string{"PERMISSION DENIED: session not granted"},
		}, false
	}
	return OneShot(declaration), true
}

// ── Permission Status ─────────────────────────────────────────────────

// PermissionStatus returns the current permission state.
func PermissionStatus() string {
	state := PermissionState(permissionGate.Load())
	switch state {
	case PermUnset:    return "permission: unset (use /allow or /deny)"
	case PermGranted:  return "permission: GRANTED ✅ (session ALLOWED+AUTHED)"
	case PermDenied:   return "permission: DENIED ❌ (session blocked)"
	case PermRevoked:  return "permission: REVOKED 🛑 (kill switch active)"
	case PermExpired:  return "permission: EXPIRED ⏰ (session ended)"
	default:           return "permission: unknown"
	}
}
