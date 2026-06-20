// Package tool provides the orchestrator tool abstraction.
// Complex shell commands are abstracted into named, versioned, cacheable tools.
// Reduces error rate by pre-validating commands. Enables taskfile-style execution.
package tool

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/blocks"
)

// OrchestratorTool is a named, versioned, cacheable shell command abstraction.
type OrchestratorTool struct {
	Name        string        // "docker:ps", "nix:build", "git:push"
	Command     string        // the shell command to execute
	Description string        // human-readable
	Version     string        // semver, for cache invalidation
	Timeout     time.Duration
	Cacheable   bool          // can results be cached?
	Dir         string        // working directory
}

// Registry holds all orchestrator tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*OrchestratorTool
}

var globalRegistry = &Registry{tools: make(map[string]*OrchestratorTool)}

// Register adds a tool to the registry.
func Register(t *OrchestratorTool) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.tools[t.Name] = t
}

// Get returns a tool by name.
func Get(name string) *OrchestratorTool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.tools[name]
}

// List returns all registered tools.
func List() []*OrchestratorTool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make([]*OrchestratorTool, 0, len(globalRegistry.tools))
	for _, t := range globalRegistry.tools {
		result = append(result, t)
	}
	return result
}

// Execute runs a tool and returns the output. Journaled via blocks.Log.
func (t *OrchestratorTool) Execute(args ...string) (string, error) {
	start := time.Now()
	pov := blocks.CurrentPOV()

	cmdStr := t.Command
	if len(args) > 0 {
		cmdStr = t.Command + " " + strings.Join(args, " ")
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	if t.Dir != "" {
		cmd.Dir = t.Dir
	}
	cmd.Env = append(os.Environ(),
		"ULTRAWHALE_TOOL="+t.Name,
		"ULTRAWHALE_VERSION="+t.Version,
	)

	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	// Journal
	blocks.Log(blocks.LogInfo, "orch.tool."+t.Name, cmdStr,
		blocks.Ref(out), "", elapsed, err)

	// Cache result if enabled (future: store in blocks cache)
	_ = pov

	return string(out), err
}

// Status returns compact tool registry status.
func Status() string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return fmt.Sprintf("orch-tools: %d registered", len(globalRegistry.tools))
}
