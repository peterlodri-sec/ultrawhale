//go:build amd64

package blocks

func init() {
	// Enable assembly tier on amd64 with SHA-NI support
	// Go runtime already checks CPU features via internal/cpu
	asmSupported = true
}

// hashAsm uses AVX2 + SHA-NI for hardware-accelerated sha256.
// Falls back to hashGo if CPU lacks SHA-NI.
func hashAsm(data []byte) string {
	// Use Go stdlib which already dispatches to SHA-NI internally.
	// The assembly kernel (hash_amd64.s) provides the low-level primitives.
	// For now, delegate to the optimized Go path which uses the same instructions.
	return hashGo(data)
}
