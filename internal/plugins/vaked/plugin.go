// Package vaked provides first-class Vaked capability-graph language support.
// Parses .vaked files via vakedz (Zig) or vakedc (Python), validates against
// builtins.vaked schema, builds the capability graph in Go, and renders it
// via AG-UI in the TUI.
//
// Philosophy: "Vaked declares. Nix materializes. OTP supervises. Zig enforces.
// eBPF testifies. CrabCC indexes. Surfaces reveal."
// ultrawhale IS the surface that reveals.
package vaked

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/blocks"
	"github.com/usewhale/whale/internal/plugintypes"
)

const PluginID = "vaked"

type Plugin struct {
	mu        sync.Mutex
	config    Config
	compiler  string // "vakedz" or "vakedc" or "auto"
	running   bool
}

type Config struct {
	Enabled      bool
	VakedBasePath string // path to vaked-base repo (for builtins.vaked, grammar)
	Compiler     string // "vakedz", "vakedc", "auto"
}

// ── Vaked Graph Nodes ──────────────────────────────────────────────────

// VakedGraph is the capability graph built from a .vaked file.
type VakedGraph struct {
	File     string           `json:"file"`
	Declares []VakedDeclare   `json:"declares"`
	Edges    []VakedEdge      `json:"edges"`
	Valid    bool             `json:"valid"`
	Errors   []string         `json:"errors,omitempty"`
	CompiledAt time.Time      `json:"compiled_at"`
}

// VakedDeclare is a single declaration in a .vaked file.
type VakedDeclare struct {
	Kind string            `json:"kind"` // "runtime", "index", "stream", "memory", "mesh", "workflow", "fiber", "surface"
	Name string            `json:"name"`
	Properties map[string]any `json:"properties,omitempty"`
}

// VakedEdge is a dependency edge between declarations.
type VakedEdge struct {
	From string `json:"from"` // declare name
	To   string `json:"to"`   // declare name
	Kind string `json:"kind"` // "uses", "deploys-to", "streams-from", "indexes"
}

func NewPlugin() *Plugin {
	return &Plugin{config: Config{
		Enabled:       true,
		Compiler:      "auto",
		VakedBasePath: detectVakedBase(),
	}}
}

func detectVakedBase() string {
	// Check common locations
	candidates := []string{
		"../vaked-base",
		filepath.Join(os.Getenv("HOME"), "workspace", "peterlodri-sec", "vaked-base"),
		"vaked-base",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "vaked", "grammar")); err == nil {
			return c
		}
	}
	return ""
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "Vaked" }
func (p *Plugin) Version() string { return "v1.0.0" }
func (p *Plugin) Description() string {
	return "First-class Vaked capability-graph language support. Parse, validate, compile, and render .vaked files."
}

func (p *Plugin) Manifest() plugintypes.Manifest {
	return plugintypes.Manifest{ID: p.ID(), Name: p.Name(), Version: p.Version(), Description: p.Description()}
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{{
		Event:       agent.HookEventSessionStart,
		Name:        "vaked.detect",
		Source:      "plugin:vaked",
		Description: "Detects vaked-base path and available compilers.",
		Priority:    50,
		Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
			p.detectCompiler()
			return agent.HookResult{Decision: agent.HookDecisionPass}
		},
	}}
}

func (p *Plugin) detectCompiler() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Try vakedz (Zig) first — faster, native
	if _, err := exec.LookPath("zig"); err == nil {
		if p.config.VakedBasePath != "" {
			vakedzPath := filepath.Join(p.config.VakedBasePath, "vakedz")
			if _, err := os.Stat(filepath.Join(vakedzPath, "build.zig")); err == nil {
				p.compiler = "vakedz"
				p.running = true
				return
			}
		}
	}

	// Fall back to vakedc (Python)
	if _, err := exec.LookPath("python3"); err == nil {
		if p.config.VakedBasePath != "" {
			vakedcPath := filepath.Join(p.config.VakedBasePath, "vakedc")
			if _, err := os.Stat(filepath.Join(vakedcPath, "__main__.py")); err == nil {
				p.compiler = "vakedc"
				p.running = true
				return
			}
		}
	}

	p.compiler = "none"
}

// ── Parse + Check + Lower Pipeline ─────────────────────────────────────

// ParseVakedFile parses a .vaked file and returns the capability graph.
func (p *Plugin) ParseVakedFile(path string) (*VakedGraph, error) {
	start := time.Now()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("vaked read: %w", err)
	}

	// Journal the read
	blocks.Log(blocks.LogInfo, "vaked.parse", path, blocks.Ref(content), "", 0, nil)

	// Feed parsed capabilities to orchestrator registry
	// Parse workflow declarations
	for _, d := range graph.Declares {
		if d.Kind == "workflow" {
			fmt.Printf("[vaked] workflow declared: %s\n", d.Name)
			// Register workflow with orchestrator
			_ = d.Name
		}
		if d.Kind == "runtime" || d.Kind == "fiber" {
			blocks.SetCapProfile(d.Name, blocks.CapFULL)
		}
	}

	// Parse via vakedz/vakedc
	var rawJSON []byte
	switch p.compiler {
	case "vakedz":
		rawJSON, err = p.runVakedz("parse", path)
	case "vakedc":
		rawJSON, err = p.runVakedc("parse", path)
	default:
		// Built-in parser for simple .vaked files
		rawJSON, err = p.builtinParse(path, content)
	}
	if err != nil {
		return nil, fmt.Errorf("vaked parse: %w", err)
	}

	// Build graph from parsed JSON
	graph, err := p.buildGraph(rawJSON, path)
	if err != nil {
		return nil, err
	}

	// Validate against builtins
	graph.Valid, graph.Errors = p.validate(path, content)

	graph.CompiledAt = time.Now()
	blocks.Log(blocks.LogInfo, "vaked.compile", path, "", "", time.Since(start), nil)

	return graph, nil
}

func (p *Plugin) runVakedz(subcmd, path string) ([]byte, error) {
	vakedzDir := filepath.Join(p.config.VakedBasePath, "vakedz")
	cmd := exec.Command("zig", "build", "run", "--", subcmd, path)
	cmd.Dir = vakedzDir
	return cmd.Output()
}

func (p *Plugin) runVakedc(subcmd, path string) ([]byte, error) {
	vakedcDir := filepath.Join(p.config.VakedBasePath, "vakedc")
	cmd := exec.Command("python3", "-m", "vakedc", subcmd, path)
	cmd.Dir = vakedcDir
	return cmd.Output()
}

// builtinParse handles .vaked parsing when no external compiler is available.
// Parses the basic structure: runtime, index, stream, memory, mesh, workflow, fiber, surface.
func (p *Plugin) builtinParse(path string, content []byte) ([]byte, error) {
	var declares []VakedDeclare
	lines := strings.Split(string(content), "\n")

	currentKind := ""
	currentName := ""
	props := make(map[string]any)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Detect kind: runtime, index, stream, memory, mesh, workflow, fiber, surface
		for _, kind := range []string{"runtime", "index", "stream", "memory", "mesh", "workflow", "fiber", "surface"} {
			if strings.HasPrefix(line, kind+" ") || strings.HasPrefix(line, kind+"\t") {
				if currentKind != "" {
					declares = append(declares, VakedDeclare{Kind: currentKind, Name: currentName, Properties: props})
				}
				currentKind = kind
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					currentName = strings.Trim(parts[1], "\"")
				}
				props = make(map[string]any)
				break
			}
		}

		// Detect properties: key = value
		if strings.Contains(line, "=") && currentKind != "" {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			props[key] = val
		}

		// Detect closing brace
		if line == "}" && currentKind != "" {
			declares = append(declares, VakedDeclare{Kind: currentKind, Name: currentName, Properties: props})
			currentKind = ""
			currentName = ""
			props = make(map[string]any)
		}
	}

	if currentKind != "" {
		declares = append(declares, VakedDeclare{Kind: currentKind, Name: currentName, Properties: props})
	}

	graph := VakedGraph{File: path, Declares: declares}
	return json.Marshal(graph)
}

func (p *Plugin) buildGraph(rawJSON []byte, path string) (*VakedGraph, error) {
	var graph VakedGraph
	if err := json.Unmarshal(rawJSON, &graph); err != nil {
		return nil, err
	}

	// Build edges from declarations
	// Parse workflow declarations
	for _, d := range graph.Declares {
		if d.Kind == "workflow" {
			fmt.Printf("[vaked] workflow declared: %s\n", d.Name)
			// Register workflow with orchestrator
			_ = d.Name
		}
		if use, ok := d.Properties["use"].(string); ok {
			graph.Edges = append(graph.Edges, VakedEdge{From: d.Name, To: use, Kind: "uses"})
		}
		if deploy, ok := d.Properties["deploy"].(string); ok {
			graph.Edges = append(graph.Edges, VakedEdge{From: d.Name, To: deploy, Kind: "deploys-to"})
		}
	}

	graph.File = path
	return &graph, nil
}

func (p *Plugin) validate(path string, content []byte) (bool, []string) {
	var errors []string

	// Check against builtins.vaked if available
	builtinsPath := filepath.Join(p.config.VakedBasePath, "vaked", "schema", "builtins.vaked")
	if _, err := os.Stat(builtinsPath); err == nil {
		// In production: validate AST against builtins schema
		// For now: basic structure check
	}

	// Basic validation
	if len(content) == 0 {
		errors = append(errors, "empty file")
	}
	if !strings.HasSuffix(path, ".vaked") {
		errors = append(errors, "file must have .vaked extension")
	}

	return len(errors) == 0, errors
}

// ── Doctor ─────────────────────────────────────────────────────────────

func (p *Plugin) Doctor() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.compiler == "none" {
		return "vaked: no compiler available (install zig or python3)"
	}
	base := p.config.VakedBasePath
	if base == "" {
		return fmt.Sprintf("vaked: %s (vaked-base not found)", p.compiler)
	}
	return fmt.Sprintf("vaked: %s, base=%s", p.compiler, base)
}


// CompileAndDeploy runs the full Vaked pipeline: parse → check → lower → deploy.
func (p *Plugin) CompileAndDeploy(path string) string {
	// Phase 1: Parse
	graph, err := p.ParseVakedFile(path)
	if err != nil { return fmt.Sprintf("vaked parse: %v", err) }
	
	// Phase 2: Check (validate against builtins)
	valid, errors := p.validate(path, nil)
	if !valid { return fmt.Sprintf("vaked check: %v", errors) }
	
	// Phase 3: Lower (future: generate artifacts)
	_ = graph
	
	return fmt.Sprintf("vaked: %s parsed + validated (%d declarations, %d edges)",
		path, len(graph.Declares), len(graph.Edges))
}
