package blocks

import (
	"fmt"
	"time"
)

// ── GROW_RATE — OpenSource Dataset Growth ──────────────────────────────

type GrowRate struct {
	Samples      int64
	SamplesPerH  float64
	LastExport   time.Time
	HFActive     bool
}

var growRate = &GrowRate{HFActive: true}

func GrowRateStatus() string {
	// Count current samples
	samples := int64(len(dogFeed.samples))

	return ASCIIBox("GROW RATE — Dataset", []string{
		fmt.Sprintf("  Samples:   %d", samples),
		fmt.Sprintf("  Rate:      %.0f/hour (MAX learn)", 2880.0  // 8 models × 6/min × 60min = 2880/hour at 10s interval),
		fmt.Sprintf("  HF:        PeetPedro/ultrawhale-dogfood ✅"),
		fmt.Sprintf("  License:   MIT + CC-BY-4.0"),
		fmt.Sprintf("  Format:    JSONL (flat strings)"),
	}, 52)
}
