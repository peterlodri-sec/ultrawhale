//go:build windows

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/usewhale/whale/internal/shell"
)

func TestWindowsShellRunForegroundAndBackground(t *testing.T) {
	dir := t.TempDir()
	ts, err := NewToolset(dir)
	if err != nil {
		t.Fatalf("new toolset: %v", err)
	}

	const marker = "whale_windows_shell_tool"
	foreground, err := ts.shellRun(context.Background(), tc("shell_run", map[string]any{
		"command": "echo " + marker,
	}))
	if err != nil || foreground.IsError {
		t.Fatalf("shell_run foreground failed: err=%v res=%+v", err, foreground)
	}
	if !strings.Contains(foreground.Content, marker) {
		t.Fatalf("foreground result missing marker %q: %s", marker, foreground.Content)
	}

	start, err := ts.shellRun(context.Background(), tc("shell_run", map[string]any{
		"command":    "echo " + marker,
		"background": true,
	}))
	if err != nil || start.IsError {
		t.Fatalf("shell_run background failed: err=%v res=%+v", err, start)
	}

	var envelope struct {
		Data struct {
			Payload struct {
				TaskID string `json:"task_id"`
			} `json:"payload"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(start.Content), &envelope); err != nil {
		t.Fatalf("unmarshal background result: %v", err)
	}
	if envelope.Data.Payload.TaskID == "" {
		t.Fatalf("expected task_id, got: %s", start.Content)
	}

	wait, err := ts.shellWait(context.Background(), tc("shell_wait", map[string]any{
		"task_id":    envelope.Data.Payload.TaskID,
		"timeout_ms": 5000,
	}))
	if err != nil || wait.IsError {
		t.Fatalf("shell_wait failed: err=%v res=%+v", err, wait)
	}
	if !strings.Contains(wait.Content, marker) {
		t.Fatalf("background result missing marker %q: %s", marker, wait.Content)
	}
}

func TestWindowsShellRunCancelKillsProcessTree(t *testing.T) {
	dir := t.TempDir()
	ts, err := NewToolset(dir)
	if err != nil {
		t.Fatalf("new toolset: %v", err)
	}

	readyPath := filepath.Join(dir, "ready")
	markerPath := filepath.Join(dir, "marker")
	t.Setenv("WHALE_SHELL_TOOL_PROCESS_TREE_HELPER", "parent")
	t.Setenv("WHALE_SHELL_TOOL_PROCESS_TREE_READY", readyPath)
	t.Setenv("WHALE_SHELL_TOOL_PROCESS_TREE_MARKER", markerPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := currentTestBinaryCommand(t)
	type execResult struct {
		res string
		err error
	}
	done := make(chan execResult, 1)
	go func() {
		res, err := ts.shellRun(ctx, tc("shell_run", map[string]any{
			"command":    command,
			"timeout_ms": 120000,
		}))
		done <- execResult{res: res.Content, err: err}
	}()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("shell_run returned before helper was ready: %v", got.err)
		}
		t.Fatalf("shell_run returned before helper was ready: %s", got.res)
	case ok := <-waitForWindowsFile(readyPath, 5*time.Second):
		if !ok {
			t.Fatalf("timed out waiting for %s", readyPath)
		}
	}
	cancel()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("shell_run returned error: %v", got.err)
		}
		if !strings.Contains(got.res, `"code":"cancelled"`) {
			t.Fatalf("expected cancelled result, got: %s", got.res)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shell_run did not return promptly after cancel")
	}

	time.Sleep(1500 * time.Millisecond)
	if _, err := os.Stat(markerPath); err == nil {
		t.Fatalf("descendant process survived cancellation and wrote %s", markerPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat marker: %v", err)
	}
}

func TestWindowsShellRunKeepsLaunchedChildOnSuccess(t *testing.T) {
	dir := t.TempDir()
	ts, err := NewToolset(dir)
	if err != nil {
		t.Fatalf("new toolset: %v", err)
	}

	markerPath := filepath.Join(dir, "marker")
	t.Setenv("WHALE_SHELL_TOOL_PROCESS_TREE_HELPER", "success-parent")
	t.Setenv("WHALE_SHELL_TOOL_PROCESS_TREE_MARKER", markerPath)

	res, err := ts.shellRun(context.Background(), tc("shell_run", map[string]any{
		"command":    currentTestBinaryCommand(t),
		"timeout_ms": 120000,
	}))
	if err != nil || res.IsError {
		t.Fatalf("shell_run failed: err=%v res=%+v", err, res)
	}

	select {
	case ok := <-waitForWindowsFile(markerPath, 4*time.Second):
		if !ok {
			t.Fatalf("launched child did not survive long enough to write %s", markerPath)
		}
	}
}

func TestWindowsShellRunProcessTreeHelper(t *testing.T) {
	switch os.Getenv("WHALE_SHELL_TOOL_PROCESS_TREE_HELPER") {
	case "parent":
		markerPath := os.Getenv("WHALE_SHELL_TOOL_PROCESS_TREE_MARKER")
		readyPath := os.Getenv("WHALE_SHELL_TOOL_PROCESS_TREE_READY")
		cmd := exec.Command(os.Args[0], "-test.run=TestWindowsShellRunProcessTreeHelper")
		cmd.Env = append(os.Environ(),
			"WHALE_SHELL_TOOL_PROCESS_TREE_HELPER=child",
			"WHALE_SHELL_TOOL_PROCESS_TREE_MARKER="+markerPath,
		)
		if err := cmd.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(readyPath, []byte("ready"), 0o644); err != nil {
			os.Exit(3)
		}
		time.Sleep(10 * time.Second)
		os.Exit(0)
	case "success-parent":
		markerPath := os.Getenv("WHALE_SHELL_TOOL_PROCESS_TREE_MARKER")
		time.Sleep(700 * time.Millisecond)
		cmd := exec.Command(os.Args[0], "-test.run=TestWindowsShellRunProcessTreeHelper")
		cmd.Env = append(os.Environ(),
			"WHALE_SHELL_TOOL_PROCESS_TREE_HELPER=success-child",
			"WHALE_SHELL_TOOL_PROCESS_TREE_MARKER="+markerPath,
		)
		if err := cmd.Start(); err != nil {
			os.Exit(5)
		}
		os.Exit(0)
	case "success-child":
		time.Sleep(700 * time.Millisecond)
		if err := os.WriteFile(os.Getenv("WHALE_SHELL_TOOL_PROCESS_TREE_MARKER"), []byte("survived"), 0o644); err != nil {
			os.Exit(6)
		}
		os.Exit(0)
	case "child":
		time.Sleep(time.Second)
		if err := os.WriteFile(os.Getenv("WHALE_SHELL_TOOL_PROCESS_TREE_MARKER"), []byte("alive"), 0o644); err != nil {
			os.Exit(4)
		}
		os.Exit(0)
	}
}

func currentTestBinaryCommand(t *testing.T) string {
	t.Helper()
	spec, err := shell.Resolve("")
	if err != nil {
		t.Fatalf("resolve shell: %v", err)
	}
	testBinary := os.Args[0]
	switch spec.Kind {
	case shell.KindPowerShell:
		return fmt.Sprintf("& '%s' '-test.run=TestWindowsShellRunProcessTreeHelper'", strings.ReplaceAll(testBinary, "'", "''"))
	case shell.KindCmd:
		if strings.ContainsAny(testBinary, " \t") {
			t.Fatalf("cmd.exe helper path cannot contain spaces: %q", testBinary)
		}
		return testBinary + " -test.run=TestWindowsShellRunProcessTreeHelper"
	default:
		t.Fatalf("unexpected Windows shell kind: %q", spec.Kind)
		return ""
	}
}

func waitForWindowsFile(path string, timeout time.Duration) <-chan bool {
	done := make(chan bool, 1)
	go func() {
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(path); err == nil {
				done <- true
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		done <- false
	}()
	return done
}
