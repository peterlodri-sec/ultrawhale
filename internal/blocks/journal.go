package blocks

import "sync"

const journalShards = 16

type Journal struct {
	shards   [journalShards]journalShard
	maxDepth int
}

type journalShard struct {
	mu      sync.Mutex
	entries map[string][]*Block
}

func NewJournal() *Journal {
	j := &Journal{maxDepth: 16}
	for i := range j.shards {
		j.shards[i].entries = make(map[string][]*Block)
	}
	return j
}

func (j *Journal) shard(path string) *journalShard {
	h := 0
	for _, c := range path { h = h*31 + int(c) }
	if h < 0 { h = -h }; return &j.shards[h%journalShards]
}

func (j *Journal) Push(path string, b *Block) {
	if b == nil { return }
	s := j.shard(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	stack := s.entries[path]
	stack = append(stack, b)
	if len(stack) > j.maxDepth { stack = stack[1:] }
	s.entries[path] = stack
}

func (j *Journal) Pop(path string) *Block {
	s := j.shard(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	stack := s.entries[path]
	if len(stack) == 0 { return nil }
	prev := stack[len(stack)-1]
	s.entries[path] = stack[:len(stack)-1]
	return prev
}

func (j *Journal) Depth(path string) int {
	s := j.shard(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries[path])
}
