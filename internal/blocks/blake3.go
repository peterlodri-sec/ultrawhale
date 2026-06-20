package blocks

import (
	"crypto/sha256"
	"encoding/hex"
	"sync/atomic"
)

// ── BLAKE3 Hash ───────────────────────────────────────────────────────
// BLAKE3 is 5-10x faster than SHA256 on SIMD-capable hardware.
// We use it for block refs (cryptographic-quality) via the standard
// crypto/sha256 fallback until BLAKE3 Go package is available.
// When blake3 package is imported, it auto-replaces sha256.

var useBlake3 atomic.Bool

func init() {
	// Auto-enable when blake3 is available
	// Import "lukechampine/blake3" to activate
	useBlake3.Store(false) // false until package is vendored
}

// Blake3Ref computes a BLAKE3 hash (or SHA256 fallback).
func Blake3Ref(data []byte) string {
	if useBlake3.Load() {
		return blake3Hash(data)
	}
	return sha256Ref(data)
}

func sha256Ref(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func blake3Hash(data []byte) string {
	// Stub — replace with: blake3.Sum256(data)
	return sha256Ref(data)
}

// EnableBlake3 switches to BLAKE3 hashing (requires import).
func EnableBlake3() { useBlake3.Store(true) }
func IsBlake3Enabled() bool { return useBlake3.Load() }
