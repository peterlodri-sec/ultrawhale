// Package blocks — Current is the runtime state snapshot.
// Distinct from Self (identity): Current measures what is happening NOW.
// Tokens, memory, cost, uptime, active hooks, busy/idle, tier.
package blocks

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Current is the runtime state snapshot.
type Current struct {
	mu sync.RWMutex

	// Session metrics
	Uptime     time.Duration
	TurnCount  int64
	Busy       bool

	// Token metrics
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
	TokensPerSec     float64
	CacheHitPct      float64

	// System metrics
	MemoryMB  int64
	CostUSD   float64
	GoRoutines int

	// Infra
	Tier          string // go/asm/gpu
	ActiveHooks   int
	QueuedPrompts int
}

var currentState atomic.Value

func init() {
	currentState.Store(&Current{Tier: CurrentTier().String()})
}

// GetCurrent returns the current runtime state snapshot.
func GetCurrent() *Current {
	_ = CurrentPOV() // POV context for current state
	c, _ := currentState.Load().(*Current)
	if c == nil {
		c = &Current{Tier: "go"}
	}
	c.GoRoutines = runtime.NumGoroutine()
	c.Tier = CurrentTier().String()

	// Memory: RSS approximation
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.MemoryMB = int64(m.Alloc / 1024 / 1024)

	return c
}

// UpdateCurrent atomically updates the runtime state.
func UpdateCurrent(fn func(*Current)) {
	c := GetCurrent()
	clone := &Current{
		PromptTokens:     c.PromptTokens,
		CompletionTokens: c.CompletionTokens,
		TotalTokens:      c.TotalTokens,
		TokensPerSec:     c.TokensPerSec,
		CacheHitPct:      c.CacheHitPct,
		MemoryMB:         c.MemoryMB,
		CostUSD:          c.CostUSD,
		Tier:             c.Tier,
		ActiveHooks:      c.ActiveHooks,
		QueuedPrompts:    c.QueuedPrompts,
	}
	fn(clone)
	currentState.Store(clone)
}

// ── Current API ────────────────────────────────────────────────────────

// RecordTokens updates token counters and tokens/sec.
func RecordTokens(prompt, completion int64, elapsed time.Duration) {
	UpdateCurrent(func(c *Current) {
		c.PromptTokens += prompt
		c.CompletionTokens += completion
		c.TotalTokens = c.PromptTokens + c.CompletionTokens
		if elapsed > 0 {
			c.TokensPerSec = float64(completion) / elapsed.Seconds()
		}
	})
}

// RecordCost adds to the session cost.
func RecordCost(cost float64) {
	UpdateCurrent(func(c *Current) {
		c.CostUSD += cost
	})
}

// SetBusy marks the agent as busy/idle.
func SetBusy(busy bool) {
	UpdateCurrent(func(c *Current) {
		c.Busy = busy
	})
}

// IncrementTurns bumps the turn counter.
func IncrementTurns() {
	UpdateCurrent(func(c *Current) {
		c.TurnCount++
	})
}

// SetCacheHitPct updates the cache hit percentage.
func SetCacheHitPct(pct float64) {
	UpdateCurrent(func(c *Current) {
		c.CacheHitPct = pct
	})
}

// SetQueuedPrompts updates the queued prompt count.
func SetQueuedPrompts(n int) {
	UpdateCurrent(func(c *Current) {
		c.QueuedPrompts = n
	})
}

// ── Display ────────────────────────────────────────────────────────────

// Status returns a compact status line for HUD/display.
func (c *Current) Status() string {
	parts := []string{}
	if c.Busy {
		parts = append(parts, "busy")
	} else {
		parts = append(parts, "idle")
	}
	parts = append(parts, fmt.Sprintf("%dt", c.TotalTokens))
	if c.TokensPerSec > 0 {
		parts = append(parts, fmt.Sprintf("%.0f/s", c.TokensPerSec))
	}
	if c.CacheHitPct > 0 {
		parts = append(parts, fmt.Sprintf("cache:%.0f%%", c.CacheHitPct))
	}
	parts = append(parts, fmt.Sprintf("%dMB", c.MemoryMB))
	if c.CostUSD > 0 {
		parts = append(parts, fmt.Sprintf("$%.4f", c.CostUSD))
	}
	return fmt.Sprintf("current: %s", joinParts(parts, " · "))
}

func joinParts(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 { result += sep }
		result += p
	}
	return result
}
