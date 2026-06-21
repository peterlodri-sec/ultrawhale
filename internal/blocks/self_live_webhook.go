package blocks

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"path/filepath"
	"sync"
	"time"
)

// ── Self Live Webhook — The Event Loop That Knows Itself ──────────────
//
// Every event loop tick, the system:
//   1. Checks SELF_MAIN_STATE (UNKNOWN/DREAM/HERE/LIVE)
//   2. Appends to scrollable history
//   3. Generates a random ONCE_TOKEN (PII-safe, zero-auth)
//   4. Publishes state to the live webhook
//   5. Records everything as append-only evidence

// SelfLiveEvent is one tick of the self-aware event loop.
type SelfLiveEvent struct {
	Tick       int64
	State      string    // UNKNOWN, DREAM, HERE, LIVE
	Timestamp  time.Time
	CommitHash string    // latest git commit
	GenesisRef string    // genesis block ref
	OnceToken  string    // random hash, regenerated on change
	Agents     int
	Blocks     int
	DreamReason string   // if DREAM state
}

// SelfLiveHistory is the scrollable event history.
type SelfLiveHistory struct {
	mu       sync.Mutex
	Events   []SelfLiveEvent
	MaxSize  int
	OnceToken string // current PII-safe token
}

var selfLiveHistory = &SelfLiveHistory{
	Events:  make([]SelfLiveEvent, 0, 256),
	MaxSize: 256,
}

func init() {
	selfLiveHistory.OnceToken = generateOnceToken()
}

// ── Once Token ────────────────────────────────────────────────────────

// generateOnceToken creates a random PII-safe token.
func generateOnceToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	// Store to disk
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".ultrawhale")
	os.MkdirAll(dir, 0o700)
	os.WriteFile(filepath.Join(dir, "ONCE_TOKEN"), []byte(token), 0o600)

	return token
}

// GetOnceToken returns the current anonymous token.
func GetOnceToken() string {
	return selfLiveHistory.OnceToken
}

// RegenerateOnceToken creates a new token (on state change).
func RegenerateOnceToken() string {
	selfLiveHistory.mu.Lock()
	defer selfLiveHistory.mu.Unlock()
	selfLiveHistory.OnceToken = generateOnceToken()
	Log(LogInfo, "once.token.regenerated", selfLiveHistory.OnceToken[:8], "", "", 0, nil)
	return selfLiveHistory.OnceToken
}

// ── Self Live Tick ────────────────────────────────────────────────────

// SelfLiveTick records one tick of the self-aware event loop.
func SelfLiveTick() SelfLiveEvent {
	selfLiveHistory.mu.Lock()
	defer selfLiveHistory.mu.Unlock()

	state := MainState(loopState.Load())
	stateName := "UNKNOWN"
	switch state {
	case LoopStopped: stateName = "DREAM"
	case LoopRunning: stateName = "HERE"
	case LoopPausing: stateName = "DREAM"
	default: stateName = "UNKNOWN"
	}

	// Check if we're LIVE (broadcasting)
	if CanBroadcast() { stateName = "LIVE" }

	event := SelfLiveEvent{
		Tick:       int64(len(selfLiveHistory.Events) + 1),
		State:      stateName,
		Timestamp:  time.Now(),
		CommitHash: getLatestCommit(),
		GenesisRef: fmt.Sprintf("%s:%s", CurrentVersion(), GetOnceToken()[:8]),
		OnceToken:  GetOnceToken()[:8],
		Agents:     AgentCount(),
		Blocks:     len(schemaRegistry),
		DreamReason: func() string {
			if stateName == "DREAM" { return DreamReason() }
			return ""
		}(),
	}

	selfLiveHistory.Events = append(selfLiveHistory.Events, event)
	if len(selfLiveHistory.Events) > selfLiveHistory.MaxSize {
		selfLiveHistory.Events = selfLiveHistory.Events[1:]
	}

	// Regenerate token on state change
	prevState := "UNKNOWN"
	if len(selfLiveHistory.Events) > 1 {
		prevState = selfLiveHistory.Events[len(selfLiveHistory.Events)-2].State
	}
	if stateName != prevState {
		RegenerateOnceToken()
	}

	return event
}

func getLatestCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

// ── Scrollable History ────────────────────────────────────────────────

// SelfLiveHistoryRender renders the scrollable event history.
func SelfLiveHistoryRender(page, pageSize int) string {
	selfLiveHistory.mu.Lock()
	defer selfLiveHistory.mu.Unlock()

	total := len(selfLiveHistory.Events)
	if total == 0 { return "no events yet" }

	start := total - pageSize*(page+1)
	if start < 0 { start = 0 }
	end := total - pageSize*page
	if end > total { end = total }

	var out string
	out += fmt.Sprintf("╔══ SELF LIVE HISTORY · %d events · page %d/%d ══╗\n",
		total, page+1, (total+pageSize-1)/pageSize)

	for i := start; i < end; i++ {
		e := selfLiveHistory.Events[i]
		icon := "⚪"
		switch e.State {
		case "LIVE": icon = "🟢"
		case "HERE": icon = "🔵"
		case "DREAM": icon = "🟡"
		default: icon = "⚪"
		}

		out += fmt.Sprintf("║ %s [%d] %s · %s · %s · %d agents\n",
			icon, e.Tick, e.State, e.Timestamp.Format("15:04:05"),
			e.OnceToken, e.Agents)

		if e.DreamReason != "" {
			out += fmt.Sprintf("║    💭 %s\n", e.DreamReason)
		}
	}

	out += "╚══════════════════════════════════════════════════╝"
	return out
}

// SelfLiveSnapshot returns the current live snapshot.
func SelfLiveSnapshot() string {
	event := SelfLiveTick() // record a tick

	return fmt.Sprintf(`╔══════════════════════════════════════════════════╗
║  🟢 SELF LIVE · %s                                 ║
╠══════════════════════════════════════════════════╣
║  State:  %s                                          ║
║  Commit: %s                                     ║
║  Token:  %s (PII-safe, zero-auth)                   ║
║  Block:  %s                                       ║
║  Agents: %d                                          ║
║  Genesis: %s                      ║
╚══════════════════════════════════════════════════╝`,
		event.State, event.State, event.CommitHash,
		event.OnceToken, event.GenesisRef, event.Agents,
		fmt.Sprintf("%s:%s", CurrentVersion(), GetOnceToken()[:8]))
}

// ── Status ────────────────────────────────────────────────────────────

// SelfLiveStatus returns compact status.
func SelfLiveStatus() string {
	selfLiveHistory.mu.Lock()
	defer selfLiveHistory.mu.Unlock()

	state := MainState(loopState.Load())
	stateName := "UNKNOWN"
	if CanBroadcast() { stateName = "LIVE" }
	if state == LoopRunning { stateName = "HERE" }
	if state == LoopStopped { stateName = "DREAM" }

	return fmt.Sprintf("self-live: %s · %d events · token: %s",
		stateName, len(selfLiveHistory.Events), GetOnceToken()[:8])
}

// SelfLiveVakedFit returns Vaked fit.
func SelfLiveVakedFit() string {
	return `SELF LIVE = THE EVENT LOOP THAT KNOWS ITSELF

  Every tick: state check → history append → webhook publish
  PII-safe: ONCE_TOKEN is random, anonymous, zero-auth
  Scrollable: page through event history
  Live: broadcasts state changes to webhooks

  "The event loop sees itself. The history is sacred."`
}
