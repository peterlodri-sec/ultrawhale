// Package runner is the unified workflow+script execution engine.
// Replaces scattered script_runner.go with a first-class module.
// Wired to: NATS (events), blocks (journal), POV (context), brain (memos).
package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/blocks"
)

// Runner executes workflows and scripts with full block integration.
type Runner struct {
	mu       sync.Mutex
	config   Config
	running  map[string]*Run
	pov      blocks.POV
}

// Config holds runner settings.
type Config struct {
	MaxConcurrency int           // max parallel workflows (default: 8)
	DefaultTimeout time.Duration // per-workflow timeout (default: 5m)
	WorkflowDir    string        // .whale/workflows/
}

// Run is an active workflow execution.
type Run struct {
	ID        string
	Name      string
	Script    string
	Status    string // "running", "completed", "failed"
	StartedAt time.Time
	EndedAt   time.Time
	Output    string
	Error     string
	POV       blocks.POV
	MemoRef   string // brain memo ref for results
}

// NewRunner creates a unified runner.
func NewRunner(cfg Config) *Runner {
	if cfg.MaxConcurrency == 0 { cfg.MaxConcurrency = 8 }
	if cfg.DefaultTimeout == 0 { cfg.DefaultTimeout = 5 * time.Minute }
	if cfg.WorkflowDir == "" { cfg.WorkflowDir = ".whale/workflows" }
	return &Runner{
		config:  cfg,
		running: make(map[string]*Run),
		pov:     blocks.CurrentPOV(),
	}
}

// Execute runs a workflow script with full block integration.
func (r *Runner) Execute(name, script string) (*Run, error) {
	r.mu.Lock()
	if len(r.running) >= r.config.MaxConcurrency {
		r.mu.Unlock()
		return nil, fmt.Errorf("max concurrency reached (%d)", r.config.MaxConcurrency)
	}

	run := &Run{
		ID:        fmt.Sprintf("run-%d", time.Now().UnixNano()),
		Name:      name,
		Script:    script,
		Status:    "running",
		StartedAt: time.Now(),
		POV:       blocks.CurrentPOV(),
	}
	r.running[run.ID] = run
	r.mu.Unlock()

	// Run in background
	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.running, run.ID)
			r.mu.Unlock()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), r.config.DefaultTimeout)
		defer cancel()

		output, err := r.runScript(ctx, script)
		run.EndedAt = time.Now()

		if err != nil {
			run.Status = "failed"
			run.Error = err.Error()
		} else {
			run.Status = "completed"
			run.Output = output
		}

		// Memoize result in brain
		memo := blocks.RememberSessionMemo(
			fmt.Sprintf("[%s] %s: %s (%s)", run.Status, run.Name, truncate(output, 200), run.EndedAt.Sub(run.StartedAt).Round(time.Millisecond)))
		run.MemoRef = memo.Ref

		// Journal the run
		blocks.Log(blocks.LogInfo, "runner.Execute", run.Name, run.MemoRef, "", run.EndedAt.Sub(run.StartedAt), err)
	}()

	return run, nil
}

func (r *Runner) runScript(ctx context.Context, script string) (string, error) {
	// Write script to temp file
	tmpFile := fmt.Sprintf("/tmp/ultrawhale-runner-%d.js", time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, []byte(script), 0o700); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)

	cmd := exec.CommandContext(ctx, "node", tmpFile)
	cmd.Env = append(os.Environ(),
		"POV_MACHINE="+r.pov.Machine,
		"POV_ARCH="+r.pov.Arch,
		"POV_TIER="+r.pov.Tier,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%v: %s", err, string(out))
	}
	return string(out), nil
}

// List returns all active runs.
func (r *Runner) List() []*Run {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*Run, 0, len(r.running))
	for _, run := range r.running {
		result = append(result, run)
	}
	return result
}

// Status returns compact runner status.
func (r *Runner) Status() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	active := len(r.running)
	return fmt.Sprintf("runner: %d active, %d max concurrency, %d POV wired",
		active, r.config.MaxConcurrency, 1)
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	if n <= 3 { return s[:n] }
	return s[:n-3] + "..."
}
