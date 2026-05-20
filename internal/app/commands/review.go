package commands

import (
	"fmt"
	"strings"
)

func ReviewPromptFromArgs(args string) (string, error) {
	args = strings.TrimSpace(args)
	kind := "local"
	target := ""
	custom := ""

	if args != "" {
		fields := strings.Fields(args)
		switch fields[0] {
		case "local", "changes":
			if len(fields) > 1 {
				return "", fmt.Errorf("usage: /review local")
			}
			kind = "local"
		case "branch":
			if len(fields) > 2 {
				return "", fmt.Errorf("usage: /review branch [base]")
			}
			kind = "branch"
			if len(fields) == 2 {
				target = fields[1]
			}
		case "pr":
			if len(fields) != 2 {
				return "", fmt.Errorf("usage: /review pr <number-or-url>")
			}
			kind = "pr"
			target = fields[1]
		case "commit":
			if len(fields) != 2 {
				return "", fmt.Errorf("usage: /review commit <sha>")
			}
			kind = "commit"
			target = fields[1]
		default:
			kind = "custom"
			custom = args
		}
	}

	return buildReviewPrompt(kind, target, custom), nil
}

func buildReviewPrompt(kind, target, custom string) string {
	var targetBlock string
	switch kind {
	case "branch":
		if strings.TrimSpace(target) == "" {
			targetBlock = `Target: current branch vs the repository default branch.

Determine the default branch using the local git configuration or remote refs, then review:
- git status --short
- git diff <base>...HEAD`
		} else {
			quotedRange := shellQuoteArg(target + "...HEAD")
			targetBlock = fmt.Sprintf(`Target: current branch vs %s.

Review:
- git status --short
- git diff %s`, target, quotedRange)
		}
	case "pr":
		quotedTarget := shellQuoteArg(target)
		targetBlock = fmt.Sprintf(`Target: pull request %s.

Review:
- gh pr view %s
- gh pr diff %s`, target, quotedTarget, quotedTarget)
	case "commit":
		quotedTarget := shellQuoteArg(target)
		targetBlock = fmt.Sprintf(`Target: commit %s.

Review:
- git show --stat --patch %s`, target, quotedTarget)
	case "custom":
		targetBlock = fmt.Sprintf(`Target: custom review request.

User request:
%s

Determine the relevant files, diff, or pull request before reviewing.`, custom)
	default:
		targetBlock = `Target: local changes.

Review:
- git status --short
- git diff --cached
- git diff

If git status shows untracked files with ??, inspect the contents of each relevant untracked file before reporting findings. Use read-only file inspection or git diff --no-index /dev/null <file> for those files.

If both staged and unstaged diffs are empty but the branch has committed changes, compare the current branch to the default branch with git diff <base>...HEAD.`
	}

	return strings.TrimSpace(fmt.Sprintf(`You are an expert code reviewer. Perform a read-only review and do not modify files.

%s

Review rules:
- Focus on correctness, security, hidden behavior changes, missing tests, and meaningful maintainability issues.
- Prefer project conventions over generic style rules.
- Do not report speculative issues. If evidence is weak, omit the finding.
- Do not include praise sections by default.
- Shell commands already run from the workspace root. Do not prefix commands with cd; use the shell_run cwd parameter for subdirectories.

Output format:
- Start with findings, ordered by severity.
- Each finding must include file/line, the problem, impact, and concrete fix direction.
- If you find no actionable issues, say that clearly and mention exactly what you reviewed.
- Keep the review concise.`, targetBlock))
}

func shellQuoteArg(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
}
