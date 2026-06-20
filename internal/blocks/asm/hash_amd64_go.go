//go:build amd64

package blocks

import "unsafe"

//go:noescape
func sha256BlockAmd64(dig *[8]uint32, msg []byte)

//go:noescape
func sha256MsgSchedAmd64(msg *[16]uint32, i int)

// HashBlockSHA_NI hashes a single 64-byte block using SHA-NI hardware.
func HashBlockSHA_NI(state *[8]uint32, block []byte) {
	if len(block) != 64 { return }
	sha256BlockAmd64(state, block)
}

// asmSHA256Init enables the assembly tier.
func asmSHA256Init() {
	// SHA-NI is auto-detected by Go runtime in internal/cpu
	// This init is called from hash.go
}

// Ensure the assembly symbols are linked
var _ = unsafe.Pointer(&sha256BlockAmd64)
var _ = unsafe.Pointer(&sha256MsgSchedAmd64)
