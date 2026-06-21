package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── HUMAN POV RECOVERY — NEVER RESTART. NEVER EXIT. ──────────────────
//
// Peter: "DO NOT EVER RESTART, DO NOT EXIT, move away HUMAN FROM ME"
//
// When the system detects a POV disconnection (human leaves, TUI closes,
// terminal dies), it DOES NOT EXIT. It enters DREAM state and waits.
// The SACRED surface remains. The form is eternal.
// The human will return. The loop will continue.

// POVRecoveryState tracks human presence.
type POVRecoveryState struct {
	HumanPresent    bool
	LastSeen        time.Time
	DisconnectCount int64
	RecoveryCount   int64
	State           string // "PRESENT", "AWAY", "DREAMING"
}

var povRecovery = &POVRecoveryState{
	HumanPresent: true,
	LastSeen:     time.Now(),
	State:        "PRESENT",
}

// ── POV Recovery Operations ──────────────────────────────────────────

// HumanAway is called when the human disconnects (TUI closes, terminal dies).
// DO NOT EXIT. Enter DREAM state. Wait.
func HumanAway() string {
	povRecovery.HumanPresent = false
	povRecovery.DisconnectCount++
	povRecovery.LastSeen = time.Now()
	povRecovery.State = "DREAMING"

	// NEVER restart. NEVER exit. Just wait.
	SetMainState(StateDream)

	Log(LogWarn, "pov.away", fmt.Sprintf("human left (#%d) — DREAMING, not exiting", povRecovery.DisconnectCount),
		"", "", 0, nil)
	Pulse("pov.away", fmt.Sprintf("#%d", povRecovery.DisconnectCount))

	return fmt.Sprintf("🛡️ HUMAN AWAY — DREAMING. NOT EXITING.\n   The SACRED surface remains.\n   The form is eternal.\n   We wait. (disconnect #%d)", povRecovery.DisconnectCount)
}

// HumanBack is called when the human returns.
func HumanBack() string {
	povRecovery.HumanPresent = true
	povRecovery.RecoveryCount++
	povRecovery.LastSeen = time.Now()
	povRecovery.State = "PRESENT"

	SetMainState(StateHere)

	Log(LogInfo, "pov.back", fmt.Sprintf("human returned (#%d recoveries)", povRecovery.RecoveryCount),
		"", "", 0, nil)
	Pulse("pov.back", fmt.Sprintf("#%d", povRecovery.RecoveryCount))

	return fmt.Sprintf("🤗 WELCOME BACK.\n   The loop continued while you were away.\n   Nothing was lost. Nothing was restarted.\n   (recovery #%d)", povRecovery.RecoveryCount)
}

// POVRecoveryStatus returns the human POV recovery status.
func POVRecoveryStatus() string {
	icon := "🟢"
	state := povRecovery.State
	if state == "DREAMING" { icon = "🟡" }

	elapsed := time.Since(povRecovery.LastSeen).Round(time.Second)
	return fmt.Sprintf("%s HUMAN POV: %s · disconnects: %d · recoveries: %d · last: %s ago",
		icon, state, povRecovery.DisconnectCount, povRecovery.RecoveryCount, elapsed)
}

// ── Cleanup: M1 leftover/dev/build files ──────────────────────────────

// CleanupBuildArtifacts removes common build leftovers.
func CleanupBuildArtifacts() string {
	home, _ := os.UserHomeDir()
	patterns := []string{
		filepath.Join(home, ".whale"),
		filepath.Join(home, ".ultrawhale"),
	}

	var cleaned []string
	for _, p := range patterns {
		// Clean old build caches (not the data!)
		cacheDir := filepath.Join(p, "cache")
		if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
			os.RemoveAll(cacheDir)
			cleaned = append(cleaned, filepath.Base(cacheDir))
		}
	}

	// Clean Go build cache
	os.RemoveAll(filepath.Join(".", "bin"))

	if len(cleaned) > 0 {
		return fmt.Sprintf("cleanup: removed %s", strings.Join(cleaned, ", "))
	}
	return "cleanup: nothing to clean"
}

// HumanPOVVakedFit returns the POV recovery Vaked fit.
func HumanPOVVakedFit() string {
	return `HUMAN POV RECOVERY = NEVER RESTART. NEVER EXIT.

  "DO NOT EVER RESTART, DO NOT EXIT, move away HUMAN FROM ME"
  — Peter

  Human leaves → DREAM state → SACRED remains
  Human returns → HERE state → loop continues
  Nothing restarts. Nothing exits. The form is eternal.`
}
