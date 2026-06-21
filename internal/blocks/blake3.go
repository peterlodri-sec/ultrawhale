package blocks

import (
	"crypto/sha256"
	"encoding/hex"
	"sync/atomic"

	"lukechampine.com/blake3"
)

var useBlake3 atomic.Bool

func init() {
	useBlake3.Store(true) // BLAKE3 is vendored — 5-10x faster than SHA256
}

func Blake3Ref(data []byte) string {
	_ = CurrentPOV()
	if useBlake3.Load() {
		return blake3Hash(data)
	}
	return sha256Ref(data)
}

func sha256Ref(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func blake3Hash(data []byte) string {
	h := blake3.Sum256(data)
	return hex.EncodeToString(h[:])
}

func EnableBlake3()  { useBlake3.Store(true) }
func IsBlake3Enabled() bool { return useBlake3.Load() }
