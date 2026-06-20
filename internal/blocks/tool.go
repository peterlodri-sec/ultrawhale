package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ToolCache caches subagent tool call results. SHA256-keyed, 5-min TTL.
// Each subagent has its own cache. Orchestrator NEVER caches.
// Invalidated on file writes for related paths.
type ToolCache struct {
	mu       sync.RWMutex
	entries  map[string]CacheEntry
	maxSize  int
	hits     int64
	misses   int64
}

// CacheEntry is a cached tool result.
type CacheEntry struct {
	ToolName  string
	ArgsHash  string
	Result    string
	Ref       string
	CachedAt  time.Time
	TTL       time.Duration
	Path      string // file path (if file operation)
}

// NewToolCache creates a tool cache for a subagent.
func NewToolCache() *ToolCache {
	return &ToolCache{
		entries: make(map[string]CacheEntry),
		maxSize: 256,
	}
}

// Key generates a cache key from tool name + normalized args.
func Key(toolName, args string) string {
	return Ref([]byte(toolName + ":" + args))[:32]
}

// Get retrieves a cached result. Returns empty string if not found or expired.
func (tc *ToolCache) Get(toolName, args string) (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	key := Key(toolName, args)
	entry, ok := tc.entries[key]
	if !ok {
		tc.misses++
		return "", false
	}

	if time.Since(entry.CachedAt) > entry.TTL {
		delete(tc.entries, key)
		tc.misses++
		return "", false
	}

	tc.hits++
	return entry.Result, true
}

// Set stores a tool result in the cache.
func (tc *ToolCache) Set(toolName, args, result, path string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if len(tc.entries) >= tc.maxSize {
		// Evict oldest entry
		var oldestKey string
		var oldestTime time.Time
		for k, v := range tc.entries {
			if oldestKey == "" || v.CachedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CachedAt
			}
		}
		delete(tc.entries, oldestKey)
	}

	key := Key(toolName, args)
	tc.entries[key] = CacheEntry{
		ToolName: toolName,
		ArgsHash: key,
		Result:   result,
		Ref:      Ref([]byte(result))[:12],
		CachedAt: time.Now(),
		TTL:      5 * time.Minute,
		Path:     path,
	}
}

// Invalidate removes cache entries for a specific file path.
func (tc *ToolCache) Invalidate(path string) int {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	count := 0
	for k, v := range tc.entries {
		if v.Path == path || (path != "" && len(v.Path) > 0 && v.Path == path) {
			delete(tc.entries, k)
			count++
		}
	}
	return count
}

// Clear removes all cache entries.
func (tc *ToolCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.entries = make(map[string]CacheEntry)
	tc.hits = 0
	tc.misses = 0
}

// Stats returns cache statistics.
func (tc *ToolCache) Stats() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	total := tc.hits + tc.misses
	rate := 0.0
	if total > 0 {
		rate = float64(tc.hits) / float64(total) * 100
	}
	return fmt.Sprintf("tool-cache: %d entries, %d hits/%d misses (%.0f%%), max %d",
		len(tc.entries), tc.hits, tc.misses, rate, tc.maxSize)
}

// ── Per-agent tool caches ──────────────────────────────────────────────

var agentCaches sync.Map // map[agentID]*ToolCache

// GetAgentCache returns the tool cache for a specific agent.
func GetAgentCache(agentID string) *ToolCache {
	c, _ := agentCaches.LoadOrStore(agentID, NewToolCache())
	return c.(*ToolCache)
}

// InvalidatePathAcrossAgents invalidates cache entries for a path across all agents.
func InvalidatePathAcrossAgents(path string) int {
	count := 0
	agentCaches.Range(func(key, value any) bool {
		c := value.(*ToolCache)
		count += c.Invalidate(path)
		return true
	})
	return count
}

func AgentCachesRange(fn func(agentID string, cache *ToolCache)) {
	agentCaches.Range(func(key, value any) bool {
		fn(key.(string), value.(*ToolCache))
		return true
	})
}

func ClearAllAgentCaches() {
	agentCaches.Range(func(key, value any) bool {
		agentCaches.Delete(key)
		return true
	})
}
