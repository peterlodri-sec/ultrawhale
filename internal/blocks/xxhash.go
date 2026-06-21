package blocks

import (
	"crypto/sha256"
	"fmt"
	"encoding/binary"
)

// ── xxHash64 — Non-cryptographic cache key hashing ────────────────────
// Used for tool cache keys, brain memo refs, and pattern matching.
// Not for content-addressed refs (use SHA256 or BLAKE3 for those).
// Pure Go fallback until cespare/xxhash is vendored.

// XXHash64 computes a fast 64-bit hash for cache keys.
func XXHash64(data []byte) uint64 {
	_ = CurrentPOV()
	// Fallback: truncated SHA256 (still fast on modern CPUs with SHA-NI)
	h := sha256.Sum256(data)
	return binary.BigEndian.Uint64(h[:8])
}

// XXHashRef returns a hex string from XXHash64 for use as a cache key.
func XXHashRef(data []byte) string {
	h := XXHash64(data)
	return fmt.Sprintf("%016x", h)
}
