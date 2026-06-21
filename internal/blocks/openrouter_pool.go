package blocks

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ── OpenRouter FREE Model Pool — Round-Robin + Stats ──────────────────
//
// Cycle through FREE OpenRouter models to maximize diversity
// and avoid rate limits. Append-only stats track every call.
//
// FREE models (always $0):
//   google/gemma-3-4b-it:free
//   mistralai/mistral-7b-instruct:free
//   meta-llama/llama-3.2-3b-instruct:free
//   huggingfaceh4/zephyr-7b-beta:free

// FreeModel is one FREE OpenRouter model.
type FreeModel struct {
	ID       string
	Provider string
	Calls    int64 // append-only
	Tokens   int64 // append-only
}

// FreeModelPool cycles through free models.
type FreeModelPool struct {
	mu         sync.Mutex
	Models     []*FreeModel
	CurrentIdx int
	Stats      FreeModelStats
}

// FreeModelStats tracks pool activity.
type FreeModelStats struct {
	TotalCalls  int64
	TotalTokens int64
	RoundRobins int64
}

var freeModelPool = &FreeModelPool{
	Models: []*FreeModel{
		{ID: "google/gemma-3-4b-it:free", Provider: "Google"},
		{ID: "mistralai/mistral-7b-instruct:free", Provider: "Mistral"},
		{ID: "meta-llama/llama-3.2-3b-instruct:free", Provider: "Meta"},
		{ID: "huggingfaceh4/zephyr-7b-beta:free", Provider: "HuggingFace"},
	},
}

// NextFreeModel returns the next free model in round-robin order.
func NextFreeModel() *FreeModel {
	freeModelPool.mu.Lock()
	defer freeModelPool.mu.Unlock()

	idx := freeModelPool.CurrentIdx
	model := freeModelPool.Models[idx]
	freeModelPool.CurrentIdx = (idx + 1) % len(freeModelPool.Models)
	freeModelPool.Stats.RoundRobins++

	return model
}

// RecordFreeCall records a call to a free model.
func RecordFreeCall(modelID string, tokens int64) {
	freeModelPool.mu.Lock()
	defer freeModelPool.mu.Unlock()

	for _, m := range freeModelPool.Models {
		if m.ID == modelID {
			atomic.AddInt64(&m.Calls, 1)
			atomic.AddInt64(&m.Tokens, tokens)
			break
		}
	}

	atomic.AddInt64(&freeModelPool.Stats.TotalCalls, 1)
	atomic.AddInt64(&freeModelPool.Stats.TotalTokens, tokens)

	Pulse("openrouter.free.call", fmt.Sprintf("%s: %d tokens", modelID, tokens))
}

// FreeModelStatsReport returns append-only stats for all free models.
func FreeModelStatsReport() string {
	freeModelPool.mu.Lock()
	defer freeModelPool.mu.Unlock()

	out := "╔══════════════════════════════════════════════════╗\n"
	out += "║  OpenRouter FREE Model Pool — Append-Only Stats  ║\n"
	out += "╠══════════════════════════════════════════════════╣\n"

	for i, m := range freeModelPool.Models {
		marker := "  "
		if i == freeModelPool.CurrentIdx { marker = "→ " }
		out += fmt.Sprintf("║ %s%-38s │ %6d calls │ %8d tok ║\n",
			marker, m.Provider[:min(12, len(m.Provider))],
			atomic.LoadInt64(&m.Calls), atomic.LoadInt64(&m.Tokens))
	}

	out += "╠══════════════════════════════════════════════════╣\n"
	out += fmt.Sprintf("║  TOTAL: %d calls · %d tokens · %d round-robins    ║\n",
		freeModelPool.Stats.TotalCalls, freeModelPool.Stats.TotalTokens,
		freeModelPool.Stats.RoundRobins)
	out += "╚══════════════════════════════════════════════════╝"

	return out
}

// FreeModelPoolStatus returns compact pool status.
func FreeModelPoolStatus() string {
	return fmt.Sprintf("openrouter-free: %d models · %d calls · %d tokens · next: %s",
		len(freeModelPool.Models), freeModelPool.Stats.TotalCalls,
		freeModelPool.Stats.TotalTokens, freeModelPool.Models[freeModelPool.CurrentIdx].Provider)
}

// FreeModelPoolVakedFit returns the pool's Vaked fit.
func FreeModelPoolVakedFit() string {
	return `OPENROUTER FREE POOL = TESTIFIES + DOGFEED

  4 FREE models. Round-robin. Append-only stats.
  Every call recorded. Every token counted.
  $0 cost. Infinite diversity.

  Google Gemma · Mistral · Meta Llama · HuggingFace Zephyr`
}
