// Package asm provides multi-platform SIMD-accelerated primitives.
// amd64: AVX2 + SHA-NI + POPCNT
// arm64: NEON + SHA256 (ARMv8 crypto extensions)
package asm

import "fmt"

// ArchCapabilities describes which hardware features are available.
type ArchCapabilities struct {
	Name    string // "Apple M1 Max", "AMD EPYC", "Intel Xeon"
	Arch    string // "arm64", "amd64"
	AVX2    bool
	NEON    bool
	SHA_NI  bool
	SHA_ARM bool
	POPCNT  bool
	BLAKE3  bool
}

// DetectArch returns the current architecture capabilities.
func DetectArch() ArchCapabilities {
	return detectArchNative()
}

// ArchStatus returns compact architecture status.
func ArchStatus() string {
	c := DetectArch()
	return fmt.Sprintf("arch: %s · %s · SIMD: %v · SHA: %v · BLAKE3: %v",
		c.Name, c.Arch, c.hasSIMD(), c.hasSHA(), c.BLAKE3)
}

func (c ArchCapabilities) hasSIMD() bool { return c.AVX2 || c.NEON }
func (c ArchCapabilities) hasSHA() bool  { return c.SHA_NI || c.SHA_ARM }
