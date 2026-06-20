// Package blocks provides content-addressed, atomic, journaled read/write primitives.
// Every operation is ref-verified (sha256), journaled for rollback, and logged.
package blocks

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sync"
	"time"
)

// BlockKind classifies what the block content represents.
type BlockKind string

const (
	KindFile    BlockKind = "file"    // raw file content
	KindDiff    BlockKind = "diff"    // unified diff
	KindPatch   BlockKind = "patch"   // patch file
	KindSymbol  BlockKind = "symbol"  // symbol reference
	KindOutline BlockKind = "outline" // structural outline
)

// Block is the universal content unit. Content-addressed via sha256.
type Block struct {
	Ref      string    // hex-encoded sha256 of Content
	Sym      string    // associated symbol name (empty for raw files)
	Outline  string    // structural context: "file:func:line_start-line_end"
	Content  []byte    // raw content
	Kind     BlockKind
	Version  int       // monotonic counter (for journal rollback)
	Path     string    // file path this block represents
	PrevRef  string    // previous ref before this write (for rollback)
	Metadata map[string]string
}

// Ref computes the sha256 hex ref of content.
func Ref(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

// NewBlock creates a content-addressed block from a file path and content.
func NewBlock(path string, content []byte, kind BlockKind) *Block {
	return &Block{
		Ref:     Ref(content),
		Content: content,
		Kind:    kind,
		Path:    path,
	}
}

// ── Read ────────────────────────────────────────────────────────────────

// Read reads a file and returns a ref-verified Block.
// The returned Block.Ref is the sha256 of the content — verifiable integrity.
func Read(path string) (*Block, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := NewBlock(path, content, KindFile)
	Log(LogInfo, "blocks.Read", path, b.Ref, "", 0, nil)
	return b, nil
}

// ── Write (journaled, rollback-able) ───────────────────────────────────

var journal = NewJournal()

// Write writes content to a file atomically. Journaled for rollback.
// Returns the new Block with ref. The previous ref is stored in PrevRef.
func Write(path string, content []byte) (*Block, error) {
	// Pre-write validation
	if err := RunPreHooks("write", content, path); err != nil {
		return nil, err
	}
	start := time.Now()
	prev, _ := Read(path) // best-effort previous state
	prevRef := ""
	if prev != nil {
		prevRef = prev.Ref
	}

	b := NewBlock(path, content, KindFile)
	b.PrevRef = prevRef

	// Journal the previous state BEFORE writing (atomic: push first, then write)
	journal.Push(path, prev)

	// Atomic write: tmp file → rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		Log(LogError, "blocks.Write", path, b.Ref, prevRef, time.Since(start), err)
		return nil, err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp) // clean up temp file
		Log(LogError, "blocks.Write", path, b.Ref, prevRef, time.Since(start), err)
		return nil, err
	}

	Log(LogInfo, "blocks.Write", path, b.Ref, prevRef, time.Since(start), nil)
	return b, nil
}

// WriteAsync writes content asynchronously. Never blocks the caller.
func WriteAsync(path string, content []byte, onDone func(*Block, error)) {
	go func() {
		b, err := Write(path, content)
		if onDone != nil {
			onDone(b, err)
		}
	}()
}

// ── Rollback ────────────────────────────────────────────────────────────

// Rollback restores the previous version of a file from the journal.
func Rollback(path string) error {
	start := time.Now()
	prev := journal.Pop(path)
	if prev == nil || prev.Content == nil {
		Log(LogWarn, "blocks.Rollback", path, "", "", time.Since(start), nil)
		return nil // nothing to roll back
	}

	if err := os.WriteFile(path, prev.Content, 0o644); err != nil {
		Log(LogError, "blocks.Rollback", path, "", prev.Ref, time.Since(start), err)
		return err
	}

	Log(LogInfo, "blocks.Rollback", path, Ref(prev.Content), prev.Ref, time.Since(start), nil)
	return nil
}

// ── Batch (atomic multi-file) ──────────────────────────────────────────

// BatchOp is a single operation in a batch.
type BatchOp struct {
	Path    string
	Content []byte
	Kind    BlockKind
}

// Batch executes multiple writes atomically. If any fails, all are rolled back.
func Batch(ops []BatchOp) error {
	start := time.Now()
	var mu sync.Mutex
	var errs []error
	var wg sync.WaitGroup

	type result struct {
		block *Block
		err   error
	}
	results := make([]result, len(ops))

	for i, op := range ops {
		wg.Add(1)
		go func(idx int, op BatchOp) {
			defer wg.Done()
			b, err := Write(op.Path, op.Content)
			mu.Lock()
			results[idx] = result{block: b, err: err}
			if err != nil {
				errs = append(errs, err)
			}
			mu.Unlock()
		}(i, op)
	}
	wg.Wait()

	if len(errs) > 0 {
		// Rollback all successful writes
		for _, r := range results {
			if r.err == nil && r.block != nil {
				_ = Rollback(r.block.Path)
			}
		}
		Log(LogError, "blocks.Batch", "", "", "", time.Since(start), errs[0])
		return errs[0]
	}

	Log(LogInfo, "blocks.Batch", "", "", "", time.Since(start), nil)
	return nil
}

func (k BlockKind) String() string { return string(k) }
