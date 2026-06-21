package blocks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ── Pre-Hook Layer ────────────────────────────────────────────────────
// Pre-hooks run BEFORE block operations. They validate, optimize, cache.
// Failures prevent the operation from executing.

// PreHook is a validation/optimization hook that runs before a block op.
type PreHook interface {
	Name() string
	Validate(data []byte, path string) error
}

// PreCommit runs git pre-commit checks.
type PreCommit struct{}

func (p *PreCommit) Name() string { return "pre-commit" }
func (p *PreCommit) Validate(data []byte, path string) error {
	// Check if there are staged changes
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, _ := cmd.Output()
	if len(out) == 0 {
		return nil // no staged changes
	}

	// Run go fmt on staged Go files
	for _, file := range strings.Split(string(out), "\n") {
		if strings.HasSuffix(file, ".go") {
			fmtCmd := exec.Command("gofmt", "-w", file)
			if err := fmtCmd.Run(); err != nil {
				return fmt.Errorf("pre-commit: gofmt %s: %w", file, err)
			}
		}
	}

	// Run go vet
	vetCmd := exec.Command("go", "vet", "./...")
	if out, err := vetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pre-commit: go vet: %s", string(out))
	}

	return nil
}

// PreWrite validates a write operation before execution.
type PreWrite struct {
	MaxSize int64 // max file size in bytes (default: 10MB)
}

func (p *PreWrite) Name() string { return "pre-write" }
func (p *PreWrite) Validate(data []byte, path string) error {
	if p.MaxSize == 0 { p.MaxSize = 10 * 1024 * 1024 } // 10MB default

	if int64(len(data)) > p.MaxSize {
		return fmt.Errorf("pre-write: file too large (%d bytes, max %d)", len(data), p.MaxSize)
	}

	if path == "" {
		return fmt.Errorf("pre-write: empty path")
	}

	// Check if parent directory exists
	if dir := path[:strings.LastIndex(path, "/")]; dir != "" {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0o755)
		}
	}

	return nil
}

// PreSed validates a sed operation before execution.
type PreSed struct{}

func (p *PreSed) Name() string { return "pre-sed" }
func (p *PreSed) Validate(data []byte, path string) error {
	// Validate pattern: find and replace should not be identical
	// This is called before Sed/SedAll — args are passed differently
	return nil
}

// PreSedPattern validates a sed pattern.
func PreSedPattern(content, find, replace []byte) error {
	if len(find) == 0 {
		return fmt.Errorf("pre-sed: empty find pattern")
	}
	if bytes.Equal(find, replace) {
		return fmt.Errorf("pre-sed: find and replace are identical")
	}

	// Estimate matches for dry-run feedback
	count := bytes.Count(content, find)
	if count > 0 {
		Log(LogInfo, "pre-sed", fmt.Sprintf("%d matches found", count), "", "", 0, nil)
	}

	return nil
}

// PreGrep validates and optimizes a grep operation.
type PreGrep struct{}

func (p *PreGrep) Name() string { return "pre-grep" }
func (p *PreGrep) Validate(data []byte, path string) error { return nil }

// ── Hook Registry ─────────────────────────────────────────────────────

// RunPreHooks executes all registered pre-hooks for a given operation.
func RunPreHooks(op string, data []byte, path string) error {
	start := time.Now()
	var hooks []PreHook

	switch op {
	case "commit":
		hooks = append(hooks, &PreCommit{})
	case "write":
		hooks = append(hooks, &PreWrite{})
	case "sed":
		// PreSed is called via PreSedPattern directly
		return nil
	case "git":
		hooks = append(hooks, &PreGit{})
	case "deploy":
		hooks = append(hooks, &PreDeploy{})
	case "grep":
		hooks = append(hooks, &PreGrep{})
	}

	for _, h := range hooks {
		// Timeout: 10s per pre-hook (prevents blocking)
		done := make(chan error, 1)
		go func() { done <- h.Validate(data, path) }()
		select {
		case err := <-done:
			if err != nil {
			Log(LogError, "prehook."+h.Name(), path, "", "", time.Since(start), err)
				return fmt.Errorf("%s: %w", h.Name(), err)
			}
		case <-time.After(10 * time.Second):
			Log(LogWarn, "prehook."+h.Name(), "timeout after 10s", "", "", time.Since(start), nil)
			return fmt.Errorf("%s: timeout", h.Name())
		}
	}

	if len(hooks) > 0 {
		Log(LogInfo, "prehook."+op, path, "", "", time.Since(start), nil)
	}
	return nil
}

// ── Wire into Write ───────────────────────────────────────────────────

// WriteWithPreHook writes with pre-write validation.
func WriteWithPreHook(path string, content []byte) (*Block, error) {
	if err := RunPreHooks("write", content, path); err != nil {
		return nil, err
	}
	return Write(path, content)
}

// SedAllWithPreHook runs sed with pattern validation.
func SedAllWithPreHook(content, find, replace []byte) ([]byte, int, error) {
	if err := PreSedPattern(content, find, replace); err != nil {
		return content, 0, err
	}
	result, count := SedAll(content, find, replace)
	return result, count, nil
}

// PreGit runs before git operations.
type PreGit struct{}

func (p *PreGit) Name() string { return "pre-git" }
func (p *PreGit) Validate(data []byte, path string) error {
	// Check working tree is clean
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("pre-git: git status failed: %w", err)
	}
	dirty := 0
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" { dirty++ }
	}
	if dirty > 0 {
		Log(LogInfo, "pre-git", fmt.Sprintf("%d dirty files", dirty), "", "", 0, nil)
	}
	return nil
}

// PreDeploy runs before deploy operations.
type PreDeploy struct{}

func (p *PreDeploy) Name() string { return "pre-deploy" }
func (p *PreDeploy) Validate(data []byte, path string) error {
	// Verify binary exists
	if _, err := os.Stat("bin/ultrawhale"); os.IsNotExist(err) {
		return fmt.Errorf("pre-deploy: binary not found — run 'task build' first")
	}
	// Check doctor
	cmd := exec.Command("bin/ultrawhale", "--dangerously-skip-permissions", "doctor")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pre-deploy: doctor failed: %s", string(out))
	}
	return nil
}

// PreStream validates streaming output before rendering.
type PreStream struct {
	MaxTokensPerTurn int // default: 100000
}

func (p *PreStream) Name() string { return "pre-stream" }
func (p *PreStream) Validate(data []byte, path string) error {
	if p.MaxTokensPerTurn == 0 { p.MaxTokensPerTurn = 100000 }
	// Token counting is done at the agent level — pre-hook validates the stream context
	return nil
}

// StreamJournal logs streaming events to the blocks journal.
func StreamJournal(eventType, content string) {
	Log(LogInfo, "stream."+eventType, truncateStr(content, 100), "", "", 0, nil)
}

func truncateStr(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n-3] + "..."
}
