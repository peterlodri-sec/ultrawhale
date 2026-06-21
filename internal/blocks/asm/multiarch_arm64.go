//go:build arm64

package asm

func detectArchNative() ArchCapabilities {
	// Apple Silicon: NEON + ARMv8 crypto always available
	return ArchCapabilities{
		Name: "Apple Silicon (NEON + ARMv8 Crypto)",
		Arch: "arm64",
		NEON: true,
		SHA_ARM: true,
		BLAKE3: true,
	}
}
