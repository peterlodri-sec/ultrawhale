package repomap

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileCacheEntry stores the SHA256 and parsed symbols for a single file.
type fileCacheEntry struct {
	SHA     string
	Symbols []Symbol
	ModTime time.Time
}

// Cache is a thread-safe in-memory cache for parsed file symbols.
// It invalidates entries when the file's SHA256 changes.
type Cache struct {
	mu    sync.RWMutex
	root  string
	files map[string]fileCacheEntry // relative path -> entry
}

// NewCache creates a cache rooted at the given workspace directory.
func NewCache(root string) *Cache {
	return &Cache{
		root:  root,
		files: make(map[string]fileCacheEntry, 256),
	}
}

// Get returns cached symbols for a file if the SHA matches, or nil if stale/missing.
func (c *Cache) Get(relPath string) ([]Symbol, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.files[relPath]
	if !ok {
		return nil, false
	}

	sha, err := fileSHA(filepath.Join(c.root, relPath))
	if err != nil || sha != entry.SHA {
		return nil, false
	}

	return entry.Symbols, true
}

// Put stores symbols for a file keyed by its current SHA256.
func (c *Cache) Put(relPath string, symbols []Symbol) error {
	sha, err := fileSHA(filepath.Join(c.root, relPath))
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.files[relPath] = fileCacheEntry{
		SHA:     sha,
		Symbols: symbols,
		ModTime: time.Now(),
	}
	return nil
}

// Invalidate removes a file from the cache.
func (c *Cache) Invalidate(relPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.files, relPath)
}

// InvalidateAll clears the entire cache.
func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files = make(map[string]fileCacheEntry, 256)
}

// Size returns the number of cached files.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.files)
}

// fileSHA returns the hex-encoded SHA256 of a file's contents.
func fileSHA(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

// FilesChanged returns the set of files that have changed since they were cached.
func (c *Cache) FilesChanged(paths []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var changed []string
	for _, relPath := range paths {
		entry, ok := c.files[relPath]
		if !ok {
			changed = append(changed, relPath)
			continue
		}
		sha, err := fileSHA(filepath.Join(c.root, relPath))
		if err != nil || sha != entry.SHA {
			changed = append(changed, relPath)
		}
	}
	return changed
}
