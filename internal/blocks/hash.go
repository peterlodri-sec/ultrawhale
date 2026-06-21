package blocks

import (
	"sync"
	"sync/atomic"
)

// HashTier represents the active hashing tier.
type HashTier int32

const (
	TierGo       HashTier = 0 // Pure Go crypto/sha256
	TierAssembly HashTier = 1 // Assembly kernel (AVX2/NEON)
	TierGPU      HashTier = 2 // GPU offload (Metal/CUDA)
)

var activeTier atomic.Int32

func init() {
	activeTier.Store(int32(detectTier()))
}

func detectTier() HashTier {
	if gpuAvailable() {
		return TierGPU
	}
	if asmAvailable() {
		return TierAssembly
	}
	return TierGo
}

// hashContent computes the sha256 ref of content using the best available tier.
func hashContent(content []byte) string {
	_ = CurrentPOV()
	if useBlake3.Load() {
		return Blake3Ref(content)
	}
	switch HashTier(activeTier.Load()) {
	case TierGPU:
		return hashGPU(content)
	case TierAssembly:
		return hashAsm(content)
	default:
		return hashGo(content)
	}
}

// hashGo uses crypto/sha256 (Go stdlib, SIMD-accelerated internally).
func hashGo(content []byte) string {
	return Ref(content)
}

// ── Assembly stubs ─────────────────────────────────────────────────────

var asmSupported bool // set by asm init

func asmAvailable() bool { return asmSupported }

func hashAsm(content []byte) string {
	return hashGo(content) // fallback — overridden by asm/hash_*.s
}

// ── GPU stubs ──────────────────────────────────────────────────────────

var gpuSupported bool
var gpuMu sync.Mutex

func gpuAvailable() bool {
	gpuMu.Lock()
	defer gpuMu.Unlock()
	if !gpuSupported {
		gpuSupported = detectGPU()
	}
	return gpuSupported
}

func detectGPU() bool { return false } // overridden by gpu/*.go

func hashGPU(content []byte) string {
	return hashGo(content) // fallback when no GPU
}

// SetTier overrides the active tier (for testing).
func SetTier(t HashTier) {
	activeTier.Store(int32(t))
}

// CurrentTier returns the active hashing tier.
func CurrentTier() HashTier {
	return HashTier(activeTier.Load())
}

// HashVakedFit documents this as a pure utility function.
func HashVakedFit() string { return "PURE UTILITY — no Vaked layer. Performance-critical." }
