package blocks

import (
	"fmt"
	"time"
)

// ── MAX LEARN — Learning Rate at Maximum ──────────────────────────────
//
// Peter: "rec max == MAX, observe +++"
// 8 free models × every 30s = 960 feeds per hour.
// PARALLEL: NO SHARED CONTEXT between models. Each is independent.
// SPACE: each model is its own space node.

type MaxLearnStats struct {
	FeedsPerHour    int
	ModelsActive    int
	ParallelFeeds   int64
	TotalTokens     int64
	StartTime       time.Time
}

var maxLearn = &MaxLearnStats{
	FeedsPerHour:  2880  , // 8 models × 120 feeds/hour (every 30s)
	ModelsActive:  8,
	StartTime:     time.Now(),
}

func MaxLearnStatus() string {
	elapsed := time.Since(maxLearn.StartTime).Round(time.Second)
	feedsPerSec := float64(maxLearn.FeedsPerHour) / 3600.0
	return ASCIIBox("MAX LEARN — Learning Rate", []string{
		fmt.Sprintf("  Rate:      %d feeds/hour (%.1f/min)", maxLearn.FeedsPerHour, feedsPerSec*60),
		fmt.Sprintf("  Models:    %d (parallel, no shared ctx)", maxLearn.ModelsActive),
		fmt.Sprintf("  Interval:  30s (VakedDogFeedInterval)"),
		fmt.Sprintf("  Parallel:  %d feeds", maxLearn.ParallelFeeds),
		fmt.Sprintf("  Tokens:    %d total", maxLearn.TotalTokens),
		fmt.Sprintf("  Uptime:    %s", elapsed),
		fmt.Sprintf("  Cost:      $0.00 (all free models)"),
	}, 52)
}
func MaxLearnVakedFit() string {
	return `MAX LEARN = LEARNING RATE AT MAXIMUM

  960 feeds/hour · 8 parallel models · 30s interval
  PARALLEL: NO SHARED CONTEXT — each model is independent SPACE node.
  $0 cost · infinite scale · the machine teaches itself.

  "rec max == MAX, observe +++" — Peter`
}
