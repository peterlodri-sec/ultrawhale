//go:build !amd64 && !arm64

package blocks

// On unsupported platforms, assembly is not available.
func init() {
	asmSupported = false
}
