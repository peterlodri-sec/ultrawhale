// Package blocks — codewhale primitive: brain + memo system.
// brain: short-term (session turns) + long-term (disk-persisted)
// memo: scoped internal notes (ultrawhale+USER, self, agents)
package blocks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── Brain ──────────────────────────────────────────────────────────────

// Brain holds ultrawhale session cognitive state.
type Brain struct {
	mu        sync.Mutex
	shortTerm []string   // last 32 turns
	longTerm  *os.File   // jsonl append file
	memos     *MemoStore
}

var sessionBrain *Brain

func init() {
	sessionBrain = NewBrain()
}

func NewBrain() *Brain {
	b := &Brain{
		shortTerm: make([]string, 0, 32),
		memos:     NewMemoStore(),
	}
	b.openLongTerm()
	return b
}

func (b *Brain) openLongTerm() {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".whale", "brain")
	os.MkdirAll(dir, 0o700)
	path := filepath.Join(dir, "long-term.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err == nil {
		b.longTerm = f
	}
}

// RememberTurn records a user/agent turn in short-term memory.
func (b *Brain) RememberTurn(turn string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.shortTerm = append(b.shortTerm, turn)
	if len(b.shortTerm) > 32 {
		b.shortTerm = b.shortTerm[len(b.shortTerm)-32:]
	}
}

// RememberLongTerm persists a fact to long-term storage.
func (b *Brain) RememberLongTerm(fact map[string]string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.longTerm != nil {
		data, _ := json.Marshal(map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"fact":      fact,
		})
		b.longTerm.Write(append(data, '\n'))
	}
}

// RecallShortTerm returns the last N turns.
func (b *Brain) RecallShortTerm(n int) []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n <= 0 || n > len(b.shortTerm) {
		n = len(b.shortTerm)
	}
	result := make([]string, n)
	copy(result, b.shortTerm[len(b.shortTerm)-n:])
	return result
}

// BrainDump returns the complete brain state for debugging.
func (b *Brain) BrainDump() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return fmt.Sprintf("brain: %d short-term turns, %d memos",
		len(b.shortTerm), b.memos.Count())
}

// ── Memo ───────────────────────────────────────────────────────────────

// MemoScope defines who can read a memo.
type MemoScope string

const (
	ScopeInternal MemoScope = "internal" // ultrawhale + USER only
	ScopeSelf     MemoScope = "self"     // this ultrawhale instance only
	ScopeAgents   MemoScope = "agents"   // shared with subagents
)

// Memo is a scoped internal note.
type Memo struct {
	Ref       string    `json:"ref"`
	Content   string    `json:"content"`
	Scope     MemoScope `json:"scope"`
	Timestamp time.Time `json:"timestamp"`
}

// MemoStore persists memos in memory + disk.
type MemoStore struct {
	mu    sync.Mutex
	memos map[string]Memo // keyed by ref
	dir   string
}

func NewMemoStore() *MemoStore {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".whale", "memos")
	os.MkdirAll(dir, 0o700)
	ms := &MemoStore{
		memos: make(map[string]Memo),
		dir:   dir,
	}
	ms.loadFromDisk()
	return ms
}

func (ms *MemoStore) loadFromDisk() {
	entries, _ := os.ReadDir(ms.dir)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(ms.dir, e.Name()))
			if err != nil {
				continue
			}
			var m Memo
			if json.Unmarshal(data, &m) == nil {
				ms.memos[m.Ref] = m
			}
		}
	}
}

// Remember stores a memo with the given scope.
func (ms *MemoStore) Remember(scope MemoScope, content string) Memo {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	m := Memo{
		Ref:       Ref([]byte(content + string(scope) + time.Now().String()))[:16],
		Content:   content,
		Scope:     scope,
		Timestamp: time.Now(),
	}
	ms.memos[m.Ref] = m

	// Persist
	path := filepath.Join(ms.dir, m.Ref+".json")
	data, _ := json.MarshalIndent(m, "", "  ")
	os.WriteFile(path, data, 0o600)

	return m
}

// Recall returns memos for the given scope.
func (ms *MemoStore) Recall(scope MemoScope) []Memo {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	var result []Memo
	for _, m := range ms.memos {
		if m.Scope == scope {
			result = append(result, m)
		}
	}
	return result
}

// RecallAll returns all memos.
func (ms *MemoStore) RecallAll() []Memo {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := make([]Memo, 0, len(ms.memos))
	for _, m := range ms.memos {
		result = append(result, m)
	}
	return result
}

// Forget removes a memo by ref.
func (ms *MemoStore) Forget(ref string) bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.memos[ref]; ok {
		delete(ms.memos, ref)
		os.Remove(filepath.Join(ms.dir, ref+".json"))
		return true
	}
	return false
}

// Count returns the total number of memos.
func (ms *MemoStore) Count() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.memos)
}

// ── Session API ────────────────────────────────────────────────────────

// RememberSessionMemo stores a memo scoped to the current session.
func RememberSessionMemo(content string) Memo {
	return sessionBrain.memos.Remember(ScopeInternal, content)
}

// RecallSessionMemos returns all internal memos.
func RecallSessionMemos() []Memo {
	return sessionBrain.memos.Recall(ScopeInternal)
}

// RecallAgentMemos returns agents-scoped memos.
func RecallAgentMemos() []Memo {
	return sessionBrain.memos.Recall(ScopeAgents)
}

// BrainStatus returns a compact brain+memory status.
func BrainStatus() string {
	return sessionBrain.BrainDump()
}

// GetBrain returns the session brain.
func GetBrain() *Brain { return sessionBrain }

func (b *Brain) ForgetMemo(ref string) bool { return b.memos.Forget(ref) }
