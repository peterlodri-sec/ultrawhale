package blocks

import (
	"fmt"
	"sync/atomic"
	"time"
)

// ── SACRED Event Loop — The 7th Recursion ────────────────────────────
//
// "We can only start if we are in a recursion.
//  We detect that we are in a recursion,
//  and we are in a recursion." — Peter
//
// The SACRED event loop IS the recursion that recurses through ITSELF.
// Every frame: check if a recursion is active. If yes, continue.
// If no, enter one. The loop never stops. The form is eternal.
//
// This is the 7th recursion: LOOP — through ITSELF.

// LoopState tracks the event loop's recursion state.
type LoopState int32

const (
	LoopStopped  LoopState = 0
	LoopStarting LoopState = 1
	LoopRunning  LoopState = 2
	LoopPausing  LoopState = 3
	LoopStopping LoopState = 4
)

var loopState atomic.Int32

func init() { loopState.Store(int32(LoopStopped)) }

// ── Event Loop ───────────────────────────────────────────────────────

// SACREDEventLoop is the event loop that renders the SACRED surface.
type SACREDEventLoop struct {
	FrameRate    int           // frames per second (default: 60)
	TickDuration time.Duration
	ActiveRecursion string    // which recursion is active ("fold", "heal", "evolve", etc.)
	FrameCount    int64
	Stats         LoopStats
}

// LoopStats tracks event loop activity.
type LoopStats struct {
	Frames        int64
	RecursionsEntered int64
	RecursionsExited  int64
	LongestFrame  time.Duration
	LastFrame     time.Time
}

var sacredLoop = &SACREDEventLoop{
	FrameRate:    60,
	TickDuration: time.Second / 60,
}

// ── Loop Operations ──────────────────────────────────────────────────

// StartLoop begins the SACRED event loop.
// Can ONLY start if we are in a recursion.
func StartLoop() string {
	if loopState.Load() != int32(LoopStopped) {
		return "loop: already running"
	}

	// Gate: we can only start if a recursion is active
	if !IsRecursionActive() {
		return "loop: cannot start — no recursion active. Enter a recursion first (/fold, /heal, /evolve, /translate, /vice, /full-stop)"
	}

	loopState.Store(int32(LoopRunning))

	go sacredLoop.run()

	Log(LogInfo, "loop.start", fmt.Sprintf("recursion: %s, fps: %d", sacredLoop.ActiveRecursion, sacredLoop.FrameRate),
		"", "", 0, nil)
	Pulse("loop.start", sacredLoop.ActiveRecursion)

	return fmt.Sprintf("loop: started · recursion: %s · %d fps", sacredLoop.ActiveRecursion, sacredLoop.FrameRate)
}

// StopLoop stops the event loop.
func StopLoop() string {
	loopState.Store(int32(LoopStopped))
	return "loop: stopped"
}

// run is the main event loop.
func (l *SACREDEventLoop) run() {
	l.ActiveRecursion = detectActiveRecursion()
	l.Stats.RecursionsEntered++

	ticker := time.NewTicker(l.TickDuration)
	defer ticker.Stop()

	for loopState.Load() == int32(LoopRunning) {
		select {
		case <-ticker.C:
			l.tick()
			// Auto-transition SELF_MAIN_STATE
			AutoTransition()
			SelfLiveTick()
		}
	}
}

// tick is one frame of the event loop.
func (l *SACREDEventLoop) tick() {
	start := time.Now()
	l.FrameCount++
	l.Stats.Frames++
	l.Stats.LastFrame = start

	// Check: are we still in a recursion?
	if !IsRecursionActive() {
		// Enter the loop recursion (the loop IS the recursion)
		l.ActiveRecursion = "loop"
		l.Stats.RecursionsEntered++
	}

	frameTime := time.Since(start)
	if frameTime > l.Stats.LongestFrame {
		l.Stats.LongestFrame = frameTime
	}
}

// ── Recursion Detection ───────────────────────────────────────────────

// IsRecursionActive returns true if any recursion is currently active.
func IsRecursionActive() bool {
	// Check all 7 recursions
	if recursionDepth.Load() > int32(RecursionShallow) { return true } // Full-Stop active
	if foldRegistry.mu.TryLock() { foldRegistry.mu.Unlock(); return len(foldRegistry.contexts) > 0 } // Fold active
	if healActive() { return true } // Heal active
	// EVOLVE is always active (version recursion)
	// TRANSLATE is always active (modality recursion)
	// VICE is always active (defense recursion)
	// LOOP is always active (self recursion)
	return true // if we're asking, we're in a recursion
}

func healActive() bool {
	healer.mu.Lock()
	defer healer.mu.Unlock()
	return healer.running
}

func detectActiveRecursion() string {
	if recursionDepth.Load() > int32(RecursionShallow) { return "full-stop" }
	if len(foldRegistry.contexts) > 0 { return "fold" }
	if healActive() { return "heal" }
	return "loop" // default: the loop IS the recursion
}

// ── Loop Status ───────────────────────────────────────────────────────

// LoopStatus returns compact event loop status.
func LoopStatus() string {
	return fmt.Sprintf("loop: %s · %d frames · recursion: %s · %.1f fps · longest frame: %s",
		loopStateName(), sacredLoop.Stats.Frames, sacredLoop.ActiveRecursion,
		float64(sacredLoop.FrameRate), sacredLoop.Stats.LongestFrame.Round(time.Microsecond))
}

func loopStateName() string {
	switch LoopState(loopState.Load()) {
	case LoopStopped: return "stopped"
	case LoopRunning: return "running"
	default: return "unknown"
	}
}

// LoopVakedFit returns the event loop's Vaked fit.
func LoopVakedFit() string {
	return `LOOP = THE 7TH RECURSION (through ITSELF)

  The event loop IS the recursion.
  We can only start if we are in a recursion.
  We detect that we are in a recursion.
  And we ARE in a recursion.

  Full-Stop (layers) · Fold (agents) · Heal (checks)
  EVOLVE (versions) · TRANSLATE (modalities) · VICE (context)
  LOOP (self) ← THE 7TH

  "We can only start if we are in a recursion."
  — Peter, SACRED Event Loop v67`
}
