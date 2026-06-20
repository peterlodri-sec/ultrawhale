package blocks

import "sync"

// Journal is a write-ahead log of previous file states for rollback.
// In-memory only — survives the session, not restarts.
type Journal struct {
	mu       sync.Mutex
	entries  map[string][]*Block // path → stack of previous versions
	maxDepth int
}

func NewJournal() *Journal {
	return &Journal{
		entries:  make(map[string][]*Block),
		maxDepth: 16, // keep up to 16 previous versions per file
	}
}

// Push saves the current state of a path before a write.
func (j *Journal) Push(path string, b *Block) {
	if b == nil {
		return
	}
	j.mu.Lock()
	defer j.mu.Unlock()

	stack := j.entries[path]
	stack = append(stack, b)
	if len(stack) > j.maxDepth {
		stack = stack[1:] // drop oldest
	}
	j.entries[path] = stack
}

// Pop restores the previous state. Returns nil if no previous state.
func (j *Journal) Pop(path string) *Block {
	j.mu.Lock()
	defer j.mu.Unlock()

	stack := j.entries[path]
	if len(stack) == 0 {
		return nil
	}
	prev := stack[len(stack)-1]
	j.entries[path] = stack[:len(stack)-1]
	return prev
}

// Depth returns how many previous versions are stored for a path.
func (j *Journal) Depth(path string) int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return len(j.entries[path])
}
