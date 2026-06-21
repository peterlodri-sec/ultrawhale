package blocks

import (
	"github.com/usewhale/whale/internal/blocks/asm"
)

// HashDispatch selects the fastest hash function for the current architecture.
func HashDispatch(data []byte) string {
	arch := asm.DetectArch()

	switch {
	case arch.SHA_NI || arch.SHA_ARM:
		// Native SHA instructions (Intel SHA-NI or ARMv8 Crypto)
		return Blake3Ref(data)
	case arch.AVX2:
		// AVX2 SIMD path
		return Blake3Ref(data)
	default:
		// Pure Go fallback
		return sha256Ref(data)
	}
}

// HashDispatchBatch hashes multiple blocks with arch-optimized dispatch.
func HashDispatchBatch(blocks [][]byte) []string {
	arch := asm.DetectArch()
	result := make([]string, len(blocks))

	if arch.AVX2 || arch.NEON {
		// Parallel with SIMD
		return MetalHashBatch(blocks)
	}

	// Sequential fallback
	for i, b := range blocks {
		result[i] = HashDispatch(b)
	}
	return result
}
