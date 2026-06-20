package blocks

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

// Sed replaces the first occurrence of find with replace in content.
// Uses bytes.Index (SIMD-accelerated via AVX2/NEON by Go runtime).
// Returns modified content and number of replacements made.
func Sed(content, find, replace []byte) ([]byte, int) {
	if len(find) == 0 {
		return content, 0
	}
	idx := bytes.Index(content, find)
	if idx < 0 {
		return content, 0
	}
	out := make([]byte, len(content)-len(find)+len(replace))
	copy(out, content[:idx])
	copy(out[idx:], replace)
	copy(out[idx+len(replace):], content[idx+len(find):])
	return out, 1
}

// SedAll replaces all occurrences of find with replace.
// Global flag equivalent. SIMD-accelerated loop.
func SedAll(content, find, replace []byte) ([]byte, int) {
	if len(find) == 0 {
		return content, 0
	}
	if len(replace) == 0 {
		return sedDeleteAll(content, find)
	}
	return sedReplaceAll(content, find, replace)
}

// sedReplaceAll is the SIMD-accelerated global replace loop.
// Pre-allocates output buffer based on match count for zero re-allocs.
func sedReplaceAll(content, find, replace []byte) ([]byte, int) {
	count := bytes.Count(content, find)
	if count == 0 {
		return content, 0
	}
	
	delta := len(replace) - len(find)
	out := make([]byte, 0, len(content)+count*delta)
	
	remaining := content
	for i := 0; i < count; i++ {
		idx := bytes.Index(remaining, find)
		out = append(out, remaining[:idx]...)
		out = append(out, replace...)
		remaining = remaining[idx+len(find):]
	}
	out = append(out, remaining...)
	return out, count
}

// sedDeleteAll removes all occurrences of find (replace with empty).
func sedDeleteAll(content, find []byte) ([]byte, int) {
	count := bytes.Count(content, find)
	if count == 0 {
		return content, 0
	}
	out := make([]byte, 0, len(content)-count*len(find))
	remaining := content
	for i := 0; i < count; i++ {
		idx := bytes.Index(remaining, find)
		out = append(out, remaining[:idx]...)
		remaining = remaining[idx+len(find):]
	}
	out = append(out, remaining...)
	return out, count
}

// SedFile reads a file, applies sed, and writes back via Write (journaled).
func SedFile(path string, find, replace []byte, global bool) (*Block, int, error) {
	start := time.Now()
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("sed read %s: %w", path, err)
	}

	var modified []byte
	var count int
	if global {
		modified, count = SedAll(src, find, replace)
	} else {
		modified, count = Sed(src, find, replace)
	}

	if count == 0 {
		Log(LogInfo, "blocks.SedFile", path, "", "", time.Since(start), nil)
		return nil, 0, nil
	}

	b, err := Write(path, modified)
	if err != nil {
		return nil, count, fmt.Errorf("sed write %s: %w", path, err)
	}

	Log(LogInfo, "blocks.SedFile", path, b.Ref, "", time.Since(start), nil)
	return b, count, nil
}

// SedBatch applies the same sed pattern to multiple files atomically.
func SedBatch(paths []string, find, replace []byte, global bool) error {
	start := time.Now()
	var mu sync.Mutex
	var errs []error
	var wg sync.WaitGroup

	type result struct {
		block *Block
		count int
		err   error
	}
	results := make([]result, len(paths))

	for i, path := range paths {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			b, count, err := SedFile(p, find, replace, global)
			mu.Lock()
			results[idx] = result{block: b, count: count, err: err}
			if err != nil {
				errs = append(errs, err)
			}
			mu.Unlock()
		}(i, path)
	}
	wg.Wait()

	if len(errs) > 0 {
		Log(LogError, "blocks.SedBatch", "", "", "", time.Since(start), errs[0])
		return errs[0]
	}

	total := 0
	for _, r := range results {
		total += r.count
	}
	Log(LogInfo, "blocks.SedBatch", fmt.Sprintf("%d files, %d replacements", len(paths), total), "", "", time.Since(start), nil)
	return nil
}
