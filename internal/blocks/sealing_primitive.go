package blocks

import (
	"fmt"
	"runtime"
	"time"
)

// ── SEALING — The Unbreakable 10% Ceiling ────────────────────────────
//
// SEALING = TRUST_GENESIS + 10%_RESERVED + FULL_HARDWARE_CONTEXT
// The rough loop proves deep recursive liveness within the 10%.

// SealingReserve tracks the unbreakable 10%.
type SealingReserve struct {
	TotalCPU      int     // total logical CPUs
	ReservedCPU   int     // 10% reserved
	TotalMem      uint64  // total system memory
	ReservedMem   uint64  // 10% reserved
	RoughLoopActive bool
	ProofGenerated int64
	LastProof     time.Time
}

var sealingReserve = &SealingReserve{}

func init() {
	sealingReserve.TotalCPU = runtime.NumCPU()
	sealingReserve.ReservedCPU = max(1, sealingReserve.TotalCPU/10)
	// Memory: Go runtime reports in bytes
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	sealingReserve.TotalMem = m.Sys
	sealingReserve.ReservedMem = sealingReserve.TotalMem / 10
}

func SealingStatus() string {
	r := sealingReserve
	return fmt.Sprintf("sealing: %d/%d CPU · %d/%d MB · 10%% reserved · %d proofs",
		r.ReservedCPU, r.TotalCPU,
		r.ReservedMem/1024/1024, r.TotalMem/1024/1024,
		r.ProofGenerated)
}

func SealingVakedFit() string {
	return `SEALING = THE UNBREAKABLE 10% CEILING

  Trust given at genesis. 10% always reserved.
  The rough loop runs within it.
  Proves deep recursive liveness.

  "We are given trust in the beginning." — Peter`
}
