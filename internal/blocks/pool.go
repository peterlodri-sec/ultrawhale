package blocks

import "sync"

// ── Buffer Pool ───────────────────────────────────────────────────────
// Reusable byte buffers for blocks engine hot paths.
// Reduces GC pressure on write/sed/diff operations.

var (
	// Small pool: <4KB buffers (log messages, cache keys)
	smallPool = sync.Pool{New: func() any { return make([]byte, 0, 4096) }}

	// Medium pool: 4KB-64KB (sed operations, diff)
	mediumPool = sync.Pool{New: func() any { return make([]byte, 0, 65536) }}

	// Large pool: >64KB (block reads, batch writes)
	largePool = sync.Pool{New: func() any { return make([]byte, 0, 262144) }}
)

// GetBuffer returns a buffer from the appropriate pool.
func GetBuffer(size int) []byte {
	_ = CurrentPOV()
	switch {
	case size <= 4096:
		return smallPool.Get().([]byte)[:0]
	case size <= 65536:
		return mediumPool.Get().([]byte)[:0]
	default:
		return largePool.Get().([]byte)[:0]
	}
}

// PutBuffer returns a buffer to its pool.
func PutBuffer(buf []byte) {
	c := cap(buf)
	switch {
	case c <= 4096:
		smallPool.Put(buf[:0])
	case c <= 65536:
		mediumPool.Put(buf[:0])
	default:
		largePool.Put(buf[:0])
	}
}
