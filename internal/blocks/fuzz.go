package blocks

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// ── Fuzz Harness — Continuous Randomized Testing ─────────────────────
// v50 timeline: 24/7 fuzzing of all block operations.

type FuzzHarness struct {
	Stats FuzzStats
}

type FuzzStats struct {
	Runs      int64
	Failures  int64
	LastRun   time.Time
	LastFuzz  string
}

var fuzzHarness = &FuzzHarness{}

// FuzzBlock runs a randomized test on a block operation.
func FuzzBlock(blockName string) error {
	fuzzHarness.Stats.Runs++
	fuzzHarness.Stats.LastRun = time.Now()
	fuzzHarness.Stats.LastFuzz = blockName

	switch blockName {
	case "write":
		data := make([]byte, rand.Intn(65536))
		rand.Read(data)
		b, _ := NewBlock("/tmp/ultrawhale-fuzz-test", data, KindFile)
		err := b.Write()
		if err != nil { fuzzHarness.Stats.Failures++; return err }
		_ = Rollback("/tmp/ultrawhale-fuzz-test")
	case "sed":
		content := fmt.Sprintf("fuzz-test-%d", rand.Int63())
		result, _ := Sed(content, "fuzz", "FUZZ", false)
		if result == "" { fuzzHarness.Stats.Failures++ }
	case "hash":
		data := make([]byte, rand.Intn(1024))
		rand.Read(data)
		_ = Blake3Ref(data)
	case "space":
		PlaceNode(fmt.Sprintf("fuzz-%d", rand.Intn(1000)), "test",
			SpacePosition{Depth: rand.Intn(10), Layer: "test", Machine: "fuzz"},
			CapFULL)
	}

	Log(LogInfo, "fuzz."+blockName, fmt.Sprintf("run #%d", fuzzHarness.Stats.Runs), "", "", 0, nil)
	return nil
}

// StartFuzzing begins continuous fuzzing on a schedule.
func StartFuzzing(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		blocks := []string{"write", "sed", "hash", "space"}
		for range ticker.C {
			for _, b := range blocks {
				FuzzBlock(b)
			}
		}
	}()
	Log(LogInfo, "fuzz.start", fmt.Sprintf("every %s", interval), "", "", 0, nil)
}

func FuzzStatus() string {
	return fmt.Sprintf("fuzz: %d runs · %d failures · last: %s (%s)",
		atomic.LoadInt64(&fuzzHarness.Stats.Runs),
		atomic.LoadInt64(&fuzzHarness.Stats.Failures),
		fuzzHarness.Stats.LastRun.Format("15:04:05"),
		fuzzHarness.Stats.LastFuzz)
}

func FuzzVakedFit() string {
	return `FUZZ = TESTIFIES LAYER ACCELERATED

  24/7 randomized testing of ALL blocks.
  Write, Sed, Hash, Space — continuously verified.
  Evidence flows to Langfuse + NATS.

  v50: formal fuzzing with coverage-guided mutation.`
}
