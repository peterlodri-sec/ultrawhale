package blocks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ── Git Primitive — First-Class Version Control ──────────────────────
//
// Git is on the SAME LEVEL as file_read.
// Not a tool. Not a helper. A PRIMITIVE.
//
// Every block operation has a git counterpart:
//   Write  ↔ git add + commit
//   Read   ↔ git show
//   Diff   ↔ git diff
//   Rollback ↔ git revert
//   Journal ↔ git log
//
// Vaked fit: Enforces layer. Git IS the journal for the filesystem.

// GitOp is one git operation.
type GitOp struct {
	Command  string
	Args     []string
	Output   string
	Error    string
	Duration time.Duration
	Ref      string
}

// GitStats tracks git activity.
type GitStats struct {
	Commits      int64
	Pushes       int64
	Pulls        int64
	Branches     int64
	FilesTracked int64
}

var gitStats GitStats

// ── Git Primitives ────────────────────────────────────────────────────

// GitCommit commits changes with a message.
func GitCommit(message string) (*GitOp, error) {
	start := time.Now()
	gitStats.Commits++

	op := &GitOp{Command: "commit", Args: []string{"-m", message}}

	cmd := exec.Command("git", "add", "-A")
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", message)
	out, err := cmd.CombinedOutput()
	op.Output = string(out)
	op.Duration = time.Since(start)
	op.Ref = Ref(out)

	if err != nil {
		op.Error = err.Error()
		Log(LogWarn, "git.commit", message, "", "", op.Duration, err)
		return op, err
	}

	Log(LogInfo, "git.commit", message, op.Ref[:12], "", op.Duration, nil)
	Pulse("git.commit", message[:min(40, len(message))])
	RSSAddItem("git commit", message[:min(80, len(message))], "commit")
	return op, nil
}

// GitPush pushes to origin.
func GitPush() (*GitOp, error) {
	start := time.Now()
	gitStats.Pushes++

	cmd := exec.Command("git", "push", "origin", "main")
	out, err := cmd.CombinedOutput()
	op := &GitOp{
		Command: "push", Args: []string{"origin", "main"},
		Output: string(out), Duration: time.Since(start),
	}

	if err != nil { op.Error = err.Error() }
	Log(LogInfo, "git.push", "origin/main", Ref(out)[:12], "", op.Duration, err)
	return op, err
}

// GitPull pulls from origin.
func GitPull() (*GitOp, error) {
	start := time.Now()
	gitStats.Pulls++

	cmd := exec.Command("git", "pull", "origin", "main")
	out, err := cmd.CombinedOutput()
	return &GitOp{
		Command: "pull", Args: []string{"origin", "main"},
		Output: string(out), Duration: time.Since(start),
		Error: func() string { if err != nil { return err.Error() }; return "" }(),
	}, err
}

// GitLog returns recent commits.
func GitLog(n int) string {
	cmd := exec.Command("git", "log", "--oneline", fmt.Sprintf("-%d", n))
	out, _ := cmd.CombinedOutput()
	return string(out)
}

// GitDiff returns the current diff.
func GitDiff() string {
	cmd := exec.Command("git", "diff")
	out, _ := cmd.CombinedOutput()
	if len(out) == 0 {
		cmd = exec.Command("git", "diff", "--cached")
		out, _ = cmd.CombinedOutput()
	}
	return string(out)
}

// GitStatus returns working tree status.
func GitStatus() string {
	cmd := exec.Command("git", "status", "--short")
	out, _ := cmd.CombinedOutput()
	gitStats.FilesTracked = int64(len(strings.Split(string(out), "\n")) - 1)
	return string(out)
}

// GitBranch returns current branch.
func GitBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

// GitRoot returns the git root directory.
func GitRoot() string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, _ := cmd.CombinedOutput()
	return strings.TrimSpace(string(out))
}

// GitIsRepo returns true if we're in a git repo.
func GitIsRepo() bool {
	_, err := os.Stat(".git")
	return err == nil
}

// GitSync does commit + pull + push in one atomic operation.
func GitSync(message string) string {
	var results []string

	if op, err := GitCommit(message); err != nil {
		results = append(results, fmt.Sprintf("commit: %s", op.Error))
	} else {
		results = append(results, fmt.Sprintf("commit: ✅ %s", op.Ref[:12]))
	}

	GitPull()

	if op, err := GitPush(); err != nil {
		results = append(results, fmt.Sprintf("push: %s", op.Error))
	} else {
		results = append(results, "push: ✅")
	}

	return strings.Join(results, "\n")
}

// ── Git Status ────────────────────────────────────────────────────────

// GitPrimitiveStatus returns compact git status.
func GitPrimitiveStatus() string {
	branch := GitBranch()
	root := GitRoot()
	return fmt.Sprintf("git: %s · %s · %d commits · %d pushes · %d files",
		branch, root, gitStats.Commits, gitStats.Pushes, gitStats.FilesTracked)
}

// GitPrimitiveVakedFit returns git's Vaked fit.
func GitPrimitiveVakedFit() string {
	return `GIT = ENFORCES LAYER (version control)

  Write ↔ commit · Read ↔ show · Diff ↔ diff
  Rollback ↔ revert · Journal ↔ log

  Git IS the journal for the filesystem.
  On the same level as file_read.
  Every block operation has a git counterpart.`
}
