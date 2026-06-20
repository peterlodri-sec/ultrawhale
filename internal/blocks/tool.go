package blocks

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// ── Tool Primitive v2 ─────────────────────────────────────────────────

// ToolDir classifies whether a tool reads, writes, or both.
type ToolDir string

const (
	DirRead  ToolDir = "read"
	DirWrite ToolDir = "write"
	DirBoth  ToolDir = "both"
)

// ToolImpl is the implementation tier.
type ToolImpl string

const (
	ImplGo     ToolImpl = "go"
	ImplAsm    ToolImpl = "asm"
	ImplGPU    ToolImpl = "gpu"
	ImplSystem ToolImpl = "system"
	ImplCustom ToolImpl = "custom"
)

// ToolScope gates who can call a tool.
type ToolScope string

const (
	ScopeSystemOrchestrator ToolScope = "SYSTEM_ORCHESTRATOR" // only orchestrator
	ScopeAgent              ToolScope = "AGENT"               // any agent
	ScopeCustom             ToolScope = "CUSTOM"              // user-defined
)

// Tool is a typed, versioned, scope-gated primitive.
type Tool struct {
	Name      string      // "grep", "write", "sed", "spawn_subagent"
	Direction ToolDir     // read, write, both
	POV       POV         // which agent registered this tool
	Version   string      // semver
	Impl      ToolImpl    // go, asm, gpu, system, custom
	Scope     ToolScope   // SYSTEM_ORCHESTRATOR, AGENT, CUSTOM
	Cache     *ToolCache  // per-tool result cache (optional)
	CreatedAt time.Time
}

// ── Tool Registry ─────────────────────────────────────────────────────

// ToolRegistry stores all registered tools with scope gating.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*Tool // keyed by name
}

var globalRegistry = &ToolRegistry{tools: make(map[string]*Tool)}

// Register adds a tool to the global registry.
func RegisterTool(t *Tool) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	t.CreatedAt = time.Now()
	if t.POV.Machine == "" {
		t.POV = CurrentPOV()
	}
	globalRegistry.tools[t.Name] = t
}

// GetTool returns a tool by name.
func GetTool(name string) *Tool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.tools[name]
}

// ListTools returns all registered tools.
func ListTools() []*Tool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make([]*Tool, 0, len(globalRegistry.tools))
	for _, t := range globalRegistry.tools {
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListToolsByScope returns tools gated to a specific scope.
func ListToolsByScope(scope ToolScope) []*Tool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	var result []*Tool
	for _, t := range globalRegistry.tools {
		if t.Scope == scope {
			result = append(result, t)
		}
	}
	return result
}

// CanCall checks if a caller (by scope) can use a tool.
func CanCall(toolName string, callerScope ToolScope) bool {
	t := GetTool(toolName)
	if t == nil {
		return false
	}
	// SYSTEM_ORCHESTRATOR can call anything
	if callerScope == ScopeSystemOrchestrator {
		return true
	}
	// AGENT can call AGENT-scoped tools
	if callerScope == ScopeAgent && (t.Scope == ScopeAgent || t.Scope == ScopeCustom) {
		return true
	}
	// CUSTOM can only call CUSTOM-scoped tools
	if callerScope == ScopeCustom && t.Scope == ScopeCustom {
		return true
	}
	return false
}

// ── Built-in system tools ─────────────────────────────────────────────

// RegisterSystemTools registers all orchestrator-owned system tools.
func RegisterSystemTools() {
	pov := CurrentPOV()

	systemTools := []Tool{
		{Name: "spawn_agent", Direction: DirWrite, Version: "v1.0.0", Impl: ImplSystem, Scope: ScopeSystemOrchestrator},
		{Name: "stop_agent", Direction: DirWrite, Version: "v1.0.0", Impl: ImplSystem, Scope: ScopeSystemOrchestrator},
		{Name: "spawn_swarm", Direction: DirWrite, Version: "v1.0.0", Impl: ImplSystem, Scope: ScopeSystemOrchestrator},
		{Name: "delegate_prompt", Direction: DirBoth, Version: "v1.0.0", Impl: ImplSystem, Scope: ScopeSystemOrchestrator},
		{Name: "brain_dump", Direction: DirRead, Version: "v1.0.0", Impl: ImplSystem, Scope: ScopeSystemOrchestrator},
	}

	for i := range systemTools {
		t := &systemTools[i]
		t.POV = pov
		RegisterTool(t)
	}
}

// RegisterAgentTools registers tools available to agents.
func RegisterAgentTools() {
	pov := CurrentPOV()

	agentTools := []Tool{
		{Name: "read", Direction: DirRead, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
		{Name: "write", Direction: DirWrite, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
		{Name: "edit", Direction: DirWrite, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
		{Name: "grep", Direction: DirRead, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
		{Name: "sed", Direction: DirWrite, Version: "v1.0.0", Impl: ImplAsm, Scope: ScopeAgent},
		{Name: "shell_run", Direction: DirBoth, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
		{Name: "web_search", Direction: DirRead, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
		{Name: "web_fetch", Direction: DirRead, Version: "v1.0.0", Impl: ImplGo, Scope: ScopeAgent},
	}

	for i := range agentTools {
		t := &agentTools[i]
		t.POV = pov
		RegisterTool(t)
	}
}

// RegisterAllTools registers both system and agent tools.
func RegisterAllTools() {
	RegisterSystemTools()
	RegisterAgentTools()
}

// ── Tool Status ───────────────────────────────────────────────────────

// ToolStatus returns a compact multi-line status of all tools.
func ToolStatus() string {
	tools := ListTools()
	var lines []string
	lines = append(lines, fmt.Sprintf("tools: %d registered", len(tools)))
	for _, t := range tools {
		cacheInfo := ""
		if t.Cache != nil {
			cacheInfo = fmt.Sprintf(" [cached:%d]", len(t.Cache.entries))
		}
		lines = append(lines, fmt.Sprintf("  %-20s %-6s %-8s %-22s%s",
			t.Name, t.Direction, t.Impl, t.Scope, cacheInfo))
	}
	return strings.Join(lines, "\n")
}

// ── ASM Dispatch ──────────────────────────────────────────────────────

// AsmDispatch accelerates tool dispatch via asm-level jump table.
// Tools with Impl=asm use this path. Falls back to Go dispatch.
func AsmDispatch(toolName string, args []byte) (string, error) {
	t := GetTool(toolName)
	if t == nil {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}
	if t.Impl != ImplAsm {
		return "", fmt.Errorf("tool %s is not asm-accelerated (impl=%s)", toolName, t.Impl)
	}
	// ASM jump table dispatch — stub for now, delegates to Go
	return asmToolDispatch(toolName, args)
}

// asmToolDispatch is the assembly-accelerated dispatch stub.
// Real implementation in asm/tool_dispatch.s
func asmToolDispatch(toolName string, args []byte) (string, error) {
	switch toolName {
	case "sed":
		// SIMD-accelerated via blocks.SedAll
		parts := strings.SplitN(string(args), "→", 3)
		if len(parts) == 3 {
			result, count := SedAll([]byte(parts[0]), []byte(parts[1]), []byte(parts[2]))
			return fmt.Sprintf("sed: %d replacements, %d bytes", count, len(result)), nil
		}
		return "", fmt.Errorf("sed: expected 'content→find→replace'")
	default:
		return "", fmt.Errorf("asm dispatch not implemented for %s", toolName)
	}
}

// ── Per-tool cache (v1 compat) ────────────────────────────────────────
// Keep the existing ToolCache for backward compat

type ToolCache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	maxSize int
	hits    int64
	misses  int64
}

type CacheEntry struct {
	ToolName string
	ArgsHash string
	Result   string
	Ref      string
	CachedAt time.Time
	TTL      time.Duration
	Path     string
}

func NewToolCache() *ToolCache {
	return &ToolCache{entries: make(map[string]CacheEntry), maxSize: 256}
}

func (tc *ToolCache) Get(toolName, args string) (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	key := Ref([]byte(toolName+":"+args))[:32]
	entry, ok := tc.entries[key]
	if !ok || time.Since(entry.CachedAt) > entry.TTL {
		if !ok { tc.misses++ } else { tc.misses++; delete(tc.entries, key) }
		return "", false
	}
	tc.hits++
	return entry.Result, true
}

func (tc *ToolCache) Set(toolName, args, result, path string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if len(tc.entries) >= tc.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range tc.entries {
			if oldestKey == "" || v.CachedAt.Before(oldestTime) {
				oldestKey = k; oldestTime = v.CachedAt
			}
		}
		delete(tc.entries, oldestKey)
	}
	key := Ref([]byte(toolName+":"+args))[:32]
	tc.entries[key] = CacheEntry{
		ToolName: toolName, ArgsHash: key, Result: result,
		Ref: Ref([]byte(result))[:12], CachedAt: time.Now(), TTL: 5 * time.Minute, Path: path,
	}
}

func (tc *ToolCache) Invalidate(path string) int {
	tc.mu.Lock(); defer tc.mu.Unlock()
	count := 0
	for k, v := range tc.entries {
		if v.Path == path { delete(tc.entries, k); count++ }
	}
	return count
}

func (tc *ToolCache) Stats() string {
	tc.mu.RLock(); defer tc.mu.RUnlock()
	total := tc.hits + tc.misses
	rate := 0.0
	if total > 0 { rate = float64(tc.hits) / float64(total) * 100 }
	return fmt.Sprintf("cache: %d entries, %d hits/%d misses (%.0f%%), max %d",
		len(tc.entries), tc.hits, tc.misses, rate, tc.maxSize)
}

var agentCaches sync.Map
func GetAgentCache(agentID string) *ToolCache {
	c, _ := agentCaches.LoadOrStore(agentID, NewToolCache())
	return c.(*ToolCache)
}
func InvalidatePathAcrossAgents(path string) int {
	count := 0
	agentCaches.Range(func(key, value any) bool { count += value.(*ToolCache).Invalidate(path); return true })
	return count
}
func AgentCachesRange(fn func(agentID string, cache *ToolCache)) {
	agentCaches.Range(func(key, value any) bool { fn(key.(string), value.(*ToolCache)); return true })
}
func ClearAllAgentCaches() {
	agentCaches.Range(func(key, value any) bool { agentCaches.Delete(key); return true })
}

func init() {
	RegisterAllTools()
}
