//go:build arm64

package blocks

// ARMv8 crypto extensions are auto-used by Go stdlib.
// No manual assembly needed — crypto/sha256 dispatches to SHA256H instruction.
func asmSHA256Init() {}
