package blocks

import (
	"fmt"
	"sync/atomic"
)

// ── Real Cost Tracking ────────────────────────────────────────────────
// Tracks actual DeepSeek API token usage + folded tokens/cost.
// "Folded tokens" = tokens counted inside ultrawhale (subagent delegation,
// block operations, brain cycles). "Real cost" = what DeepSeek charges.

// DeepSeek V4 pricing (per 1M tokens)
const (
	V4FlashInputPrice  = 0.14  // $0.14 per 1M input tokens
	V4FlashOutputPrice = 0.28  // $0.28 per 1M output tokens
	V4ProInputPrice    = 0.55  // $0.55 per 1M input tokens
	V4ProOutputPrice   = 1.10  // $1.10 per 1M output tokens
	CacheHitDiscount   = 0.02  // 98% discount on cache hits
)

// Sonnet 4.7 estimated pricing (for comparison)
const (
	Sonnet47InputPrice  = 3.00  // $3.00 per 1M input tokens
	Sonnet47OutputPrice = 15.00 // $15.00 per 1M output tokens
)

// RealCost tracks actual and folded token costs.
type RealCost struct {
	// DeepSeek API tokens (actual billing)
	APIInputTokens     int64
	APIOutputTokens    int64
	APICacheHitTokens  int64
	APITotalCost       float64

	// Folded tokens (internal — subagent, blocks, brain)
	FoldedInputTokens  int64
	FoldedOutputTokens int64
	FoldedTotalCost    float64 // what Sonnet 4.7 would charge

	// Total
	TotalTokens        int64
	TotalCost          float64
}

var realCost atomic.Value

func init() { realCost.Store(&RealCost{}) }

// RecordAPITokens records actual DeepSeek API usage.
func RecordAPITokens(input, output, cacheHit int64) {
	c := realCost.Load().(*RealCost)
	clone := *c
	clone.APIInputTokens += input
	clone.APIOutputTokens += output
	clone.APICacheHitTokens += cacheHit
	clone.APITotalCost = float64(clone.APIInputTokens)*V4FlashInputPrice/1e6 +
		float64(clone.APIOutputTokens)*V4FlashOutputPrice/1e6 -
		float64(clone.APICacheHitTokens)*V4FlashInputPrice*CacheHitDiscount/1e6
	clone.TotalTokens = clone.APIInputTokens + clone.APIOutputTokens + clone.FoldedInputTokens + clone.FoldedOutputTokens
	clone.TotalCost = clone.APITotalCost + clone.FoldedTotalCost
	realCost.Store(&clone)
}

// RecordFoldedTokens records internal (folded) token usage.
func RecordFoldedTokens(input, output int64) {
	c := realCost.Load().(*RealCost)
	clone := *c
	clone.FoldedInputTokens += input
	clone.FoldedOutputTokens += output
	clone.FoldedTotalCost = float64(clone.FoldedInputTokens)*Sonnet47InputPrice/1e6 +
		float64(clone.FoldedOutputTokens)*Sonnet47OutputPrice/1e6
	clone.TotalTokens = clone.APIInputTokens + clone.APIOutputTokens + clone.FoldedInputTokens + clone.FoldedOutputTokens
	clone.TotalCost = clone.APITotalCost + clone.FoldedTotalCost
	realCost.Store(&clone)
}

// GetRealCost returns the current cost state.
func GetRealCost() *RealCost {
	return realCost.Load().(*RealCost)
}

// RealCostStatus returns a compact cost display string.
func RealCostStatus() string {
	c := GetRealCost()
	return fmt.Sprintf("API: %dt ($%.4f) | Folded: %dt (~$%.4f sonnet47) | Total: $%.4f",
		c.APIInputTokens+c.APIOutputTokens, c.APITotalCost,
		c.FoldedInputTokens+c.FoldedOutputTokens, c.FoldedTotalCost,
		c.TotalCost)
}
