//go:build amd64

package asm

import (
	"fmt"
	"golang.org/x/sys/cpu"
)

func detectArchNative() ArchCapabilities {
	c := ArchCapabilities{
		Arch: "amd64",
		AVX2: cpu.X86.HasAVX2,
		SHA_NI: cpu.X86.HasSHA,
		POPCNT: cpu.X86.HasPOPCNT,
		BLAKE3: true,
	}
	if c.AVX2 && c.SHA_NI { c.Name = "AMD/Intel x86_64 (AVX2 + SHA-NI)" } else { c.Name = "amd64 (base)" }
	return c
}
