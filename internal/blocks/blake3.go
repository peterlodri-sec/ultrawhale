package blocks

import (
	"crypto/sha256"
	"encoding/hex"
	"sync/atomic"
	"strings"
	"sync"

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


// Blake3TreeRef computes a BLAKE3 hash using tree mode for parallel execution.
// For files >1MB, splits into 1MB chunks and hashes in parallel (O(n/P)).
func Blake3TreeRef(data []byte) string {
	const chunkSize = 1 << 20 // 1MB chunks
	
	if len(data) <= chunkSize || !useBlake3.Load() {
		return Blake3Ref(data)
	}
	
	// Parallel tree hashing: split into chunks, hash each, then hash the tree
	numChunks := (len(data) + chunkSize - 1) / chunkSize
	chunkHashes := make([]string, numChunks)
	
	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) { end = len(data) }
		
		wg.Add(1)
		go func(idx int, chunk []byte) {
			defer wg.Done()
			chunkHashes[idx] = Blake3Ref(chunk)
		}(i, data[start:end])
	}
	wg.Wait()
	
	// Hash the tree root
	root := strings.Join(chunkHashes, "")
	return Blake3Ref([]byte(root))
}
