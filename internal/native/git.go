package native

import (
	"os"
	"os/exec"
	"strings"
)

// ── Git (minimal native operations) ────────────────────────────────────
// Full git operations still use shell_exec for complex commands (rebase, merge).
// But the most common read operations are native Go: status, log, branch, diff.

// GitStatus returns the working tree status.
func GitStatus(workspace string) (string, error) {
	return gitCmd(workspace, "status", "--short")
}

// GitBranch returns the current branch name.
func GitBranch(workspace string) (string, error) {
	out, err := gitCmd(workspace, "branch", "--show-current")
	return strings.TrimSpace(out), err
}

// GitLog returns recent commit log.
func GitLog(workspace string, n int) ([]Commit, error) {
	out, err := gitCmd(workspace, "log", "--oneline", "-n", itoa(n))
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			commits = append(commits, Commit{SHA: parts[0], Message: parts[1]})
		}
	}
	return commits, nil
}

type Commit struct {
	SHA     string
	Message string
}

// GitDiff returns the diff of unstaged changes.
func GitDiff(workspace string) (string, error) {
	return gitCmd(workspace, "diff")
}

// GitDiffStaged returns the diff of staged changes.
func GitDiffStaged(workspace string) (string, error) {
	return gitCmd(workspace, "diff", "--staged")
}

// GitDiffBranch returns diff between current branch and another.
func GitDiffBranch(workspace, branch string) (string, error) {
	return gitCmd(workspace, "diff", branch+"...HEAD")
}

// GitCreateBranch creates and checks out a new branch.
func GitCreateBranch(workspace, name string) (string, error) {
	return gitCmd(workspace, "checkout", "-b", name)
}

// GitCommit commits staged changes.
func GitCommit(workspace, message string) (string, error) {
	return gitCmd(workspace, "commit", "-m", message)
}

// GitPush pushes the current branch.
func GitPush(workspace, remote, branch string) (string, error) {
	return gitCmd(workspace, "push", remote, branch)
}

// GitPushForce pushes with force.
func GitPushForce(workspace, remote, branch string) (string, error) {
	return gitCmd(workspace, "push", "--force-with-lease", remote, branch)
}

// ── Helpers ────────────────────────────────────────────────────────────

func gitCmd(workspace string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workspace
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out), &GitError{Msg: string(exitErr.Stderr), ExitCode: exitErr.ExitCode()}
		}
		return string(out), err
	}
	return string(out), nil
}

type GitError struct {
	Msg      string
	ExitCode int
}

func (e *GitError) Error() string { return e.Msg }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte(0 + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
