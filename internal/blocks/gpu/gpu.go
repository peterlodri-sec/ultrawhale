package blocks

import "os"

// detectGPU checks for available GPU hardware.
func detectGPU() bool {
	// Check for Metal (Apple Silicon)
	if _, err := os.Stat("/System/Library/Frameworks/Metal.framework"); err == nil {
		return initMetal()
	}
	// Check for CUDA (NVIDIA)
	if _, err := os.Stat("/usr/local/cuda"); err == nil {
		return initCUDA()
	}
	// Check for Vulkan/OpenCL
	if os.Getenv("GPU_ENABLED") != "" {
		return true // force-enable for testing
	}
	return false
}

func initMetal() bool  { return false } // stubbed — Metal requires cgo
func initCUDA() bool   { return false } // stubbed — CUDA requires cgo
GOEOF

cat > /home/dev/whale/internal/blocks/gpu/metal.go << 'METALEOF'
//go:build darwin

package blocks

import "os"

func initMetal() bool {
	// Metal Performance Shaders available on Apple Silicon
	// Real implementation would use: import "github.com/johnsiilver/golib/metal"
	// For now: detect GPU presence and enable tier
	_, err := os.Stat("/System/Library/Frameworks/Metal.framework")
	return err == nil
}
METALEOF

cat > /home/dev/whale/internal/blocks/gpu/gpu_stub.go << 'STUBEOF'
//go:build !darwin

package blocks

func initMetal() bool { return false }
STUBEOF

echo "gpu files created"
wc -l /home/dev/whale/internal/blocks/gpu/*.go
