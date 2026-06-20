package repomap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
)

const PluginID = "repomap"

// Plugin is the repo map plugin. It implements the Plugin interface
// expected by internal/plugins, but does not import it to avoid cycles.
// Registration is done via a thin wrapper in internal/plugins/repomap_plugin.go.
type Plugin struct {
	cache     *Cache
	graph     *Graph
	mu        sync.RWMutex
	root      string
	ready     bool
	lastBuild time.Time
}

func NewPlugin(workspaceRoot string) *Plugin {
	return &Plugin{
		cache: NewCache(workspaceRoot),
		root:  workspaceRoot,
	}
}

func (p *Plugin) ID() string       { return PluginID }
func (p *Plugin) Name() string     { return "Repo Map" }
func (p *Plugin) Version() string  { return "0.1.0" }
func (p *Plugin) Description() string {
	return "Builds a dependency graph of the workspace and injects it into the LLM context."
}

func (p *Plugin) StartupContext(ctx context.Context, workspaceRoot string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.ready || p.graph == nil {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("[REPO_MAP]\n")
	sb.WriteString(fmt.Sprintf("Files: %d  Symbols: %d  Edges: %d\n",
		len(p.graph.Files), len(p.graph.Symbols), len(p.graph.Edges)))
	sb.WriteString(p.graph.ToJSON())
	sb.WriteString("\n[/REPO_MAP]")
	return sb.String(), nil
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{
		{
			Event:       agent.HookEventPostToolUse,
			Name:        "repomap.invalidate-on-write",
			Source:      "plugin:repomap",
			Description: "Invalidates cached symbols when a file is written or edited.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				return p.onPostToolUse(payload)
			},
		},
		{
			Event:       agent.HookEventSessionStart,
			Name:        "repomap.build-on-start",
			Source:      "plugin:repomap",
			Description: "Builds the initial repo map when a session starts.",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				return p.onSessionStart()
			},
		},
		{
			Event:       agent.HookEventPrePromptSubmit,
			Name:        "repomap.inject-context",
			Source:      "plugin:repomap",
			Description: "Injects repo map summary before prompt submission.",
			Priority:    50,
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				return p.onPrePromptSubmit(payload)
			},
		},
		{
			Event:       agent.HookEventError,
			Name:        "repomap.on-error",
			Source:      "plugin:repomap",
			Description: "Logs parse failures and cache errors.",
			Priority:    10,
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				return agent.HookResult{Decision: agent.HookDecisionPass}
			},
		},
	}
}

func (p *Plugin) onPostToolUse(payload agent.HookPayload) agent.HookResult {
	switch payload.ToolName {
	case "write", "edit", "multi_edit":
		if args, ok := payload.ToolArgs.(map[string]any); ok {
			if filePath, ok := args["file_path"].(string); ok {
				relPath, err := filepath.Rel(p.root, filePath)
				if err == nil && !strings.HasPrefix(relPath, "..") {
					p.cache.Invalidate(relPath)
					go p.rebuildFile(relPath)
				}
			}
		}
	}
	return agent.HookResult{Decision: agent.HookDecisionPass}
}

func (p *Plugin) onSessionStart() agent.HookResult {
	go p.buildFull()
	return agent.HookResult{Decision: agent.HookDecisionPass}
}

func (p *Plugin) buildFull() {
	p.mu.Lock()
	p.ready = false
	p.mu.Unlock()

	root := p.root
	var paths []string
	_ = filepath.Walk(root, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(fullPath)
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == "target" ||
				base == ".git" || base == "zig-cache" || base == "zig-out" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, fullPath)
		ext := filepath.Ext(rel)
		if detectExt(ext) {
			paths = append(paths, rel)
		}
		return nil
	})

	graph, err := BuildGraphSIMD(root, paths, p.cache)
	if err != nil {
		return
	}

	p.mu.Lock()
	p.graph = graph
	p.lastBuild = time.Now()
	p.ready = true
	p.mu.Unlock()
}

func (p *Plugin) rebuildFile(relPath string) {
	fullPath := filepath.Join(p.root, relPath)
	src, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}
	syms := ExtractSymbolsSIMD(relPath, src)
	_ = p.cache.Put(relPath, syms)

	p.mu.Lock()
	defer p.mu.Unlock()
	for key, sym := range p.graph.Symbols {
		if sym.File == relPath {
			delete(p.graph.Symbols, key)
		}
	}
	for _, sym := range syms {
		p.graph.AddSymbol(sym)
	}
}

func (p *Plugin) Doctor() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.ready {
		return "repo map: not yet built"
	}
	return fmt.Sprintf("repo map: %d files, %d symbols, %d edges (built %s)",
		len(p.graph.Files), len(p.graph.Symbols), len(p.graph.Edges),
		p.lastBuild.Format(time.RFC3339))
}

func (p *Plugin) SetRoot(root string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.root = root
	p.cache = NewCache(root)
	p.ready = false
}

// ForceRebuild clears the cache and rebuilds the repo map immediately.
// Useful for /reload repomap and after large-scale refactors.
func (p *Plugin) ForceRebuild() {
	p.cache.InvalidateAll()
	p.buildFull()
}

// ── Additional hooks ──────────────────────────────────────────────────

func (p *Plugin) onPrePromptSubmit(payload agent.HookPayload) agent.HookResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.ready {
		return agent.HookResult{Decision: agent.HookDecisionPass}
	}
	return agent.HookResult{
		Decision:          agent.HookDecisionPass,
		AdditionalContext: p.graphSummary(),
	}
}

func (p *Plugin) graphSummary() string {
	return fmt.Sprintf("[REPO_MAP: %d files, %d symbols]", len(p.graph.Files), len(p.graph.Symbols))
}
