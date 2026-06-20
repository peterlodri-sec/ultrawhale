//go:build arm64

package blocks

func init() {
	// ARMv8 crypto extensions (FEAT_SHA256) available on Apple Silicon M1+
	asmSupported = true
}

func hashAsm(data []byte) string {
	return hashGo(data)
}
