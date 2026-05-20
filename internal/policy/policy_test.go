package policy

import (
	"strconv"
	"testing"

	"github.com/usewhale/whale/internal/core"
	whaleTools "github.com/usewhale/whale/internal/tools"
)

func TestDefaultToolPolicyPrefixRulesApplyToShellRunCommand(t *testing.T) {
	p := DefaultToolPolicy{
		Mode:          ApprovalModeOnRequest,
		AllowPrefixes: []string{"git status"},
		DenyPrefixes:  []string{"rm -rf"},
	}
	spec := core.ToolSpec{Name: "shell_run"}

	allow := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":"git status --short"}`})
	if !allow.Allow || allow.RequiresApproval || allow.Code != "allow_prefix" || allow.MatchedRule != "git status" {
		t.Fatalf("expected allow-prefix decision for shell_run.command: %+v", allow)
	}

	deny := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":"rm -rf /tmp/x"}`})
	if deny.Allow || deny.Code != "policy_denied" || deny.MatchedRule != "rm -rf" {
		t.Fatalf("expected deny-prefix decision for shell_run.command: %+v", deny)
	}
}

func TestDefaultToolPolicyPrefixRulesRequireTokenBoundary(t *testing.T) {
	p := DefaultToolPolicy{
		Mode:          ApprovalModeOnRequest,
		AllowPrefixes: []string{"git status"},
		DenyPrefixes:  []string{"rm -rf"},
	}
	spec := core.ToolSpec{Name: "shell_run"}

	allow := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":"git   status   --short"}`})
	if !allow.Allow || allow.RequiresApproval || allow.Code != "allow_prefix" {
		t.Fatalf("expected whitespace-normalized allow-prefix decision: %+v", allow)
	}
	notAllow := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":"git statusfoo"}`})
	if !notAllow.Allow || !notAllow.RequiresApproval || notAllow.Code != "approval_required" {
		t.Fatalf("expected statusfoo not to match git status prefix: %+v", notAllow)
	}
	newline := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":"git\nstatus --short"}`})
	if !newline.Allow || !newline.RequiresApproval || newline.Code != "approval_required" {
		t.Fatalf("expected newline-separated command not to match git status prefix: %+v", newline)
	}
	notDeny := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":"rm -rfoo /tmp/x"}`})
	if !notDeny.Allow || !notDeny.RequiresApproval || notDeny.Code != "approval_required" {
		t.Fatalf("expected rm -rfoo not to match rm -rf deny prefix: %+v", notDeny)
	}
}

func TestDefaultToolPolicyAutoAllowsCommonShellChecksInOnRequest(t *testing.T) {
	p := DefaultToolPolicy{Mode: ApprovalModeOnRequest}
	spec := productionShellRunSpec(t)
	for _, command := range []string{
		"git status --short",
		"git status --short 2>&1",
		"git -C internal status --short",
		"git diff",
		"git diff --cached",
		"git diff --cached 2>&1",
		"git diff -- internal/policy/policy_test.go | tail -80",
		"git diff -- internal/tools/catalog_shell.go | head -40",
		"rg whale internal | wc -l",
		"git diff --stat",
		"git diff main...HEAD",
		"git diff --no-index /dev/null internal/app/commands/review.go",
		"git show --stat --patch HEAD",
		"git log --oneline -5",
		"git branch --show-current",
		"git branch -a",
		"git remote -v",
		"git remote get-url origin",
		"git rev-parse --abbrev-ref HEAD",
		"git config --get remote.origin.url",
		"rg whale internal",
		"ls -u",
		"uptime",
		"cal",
		"id -u",
		"uname -a",
		"whoami",
		"df -h",
		"du -sh .",
		"locale",
		"groups",
		"nproc",
		"stat internal/policy/policy.go",
		"strings bin/whale",
		"hexdump -C internal/policy/policy.go",
		"od -c internal/policy/policy.go",
		"nl -ba internal/policy/policy.go",
		"basename internal/policy/policy.go",
		"dirname internal/policy/policy.go",
		"realpath internal/policy/policy.go",
		"readlink bin/whale",
		"cut -d : -f 1 internal/policy/policy.go",
		"paste internal/policy/policy.go internal/policy/policy_test.go",
		"tr a-z A-Z",
		"column -t internal/policy/policy.go",
		"tac internal/policy/policy.go",
		"rev internal/policy/policy.go",
		"fold -w 80 internal/policy/policy.go",
		"expand internal/policy/policy.go",
		"unexpand internal/policy/policy.go",
		"comm internal/policy/policy.go internal/policy/policy_test.go",
		"cmp internal/policy/policy.go internal/policy/policy_test.go",
		"numfmt --to=iec 1024",
		"true",
		"false",
		"which whale",
		"type whale",
		"expr 1 + 1",
		"test -f internal/policy/policy.go",
		"getconf ARG_MAX",
		"seq 1 3",
		"tsort internal/policy/policy.go",
		"pr internal/policy/policy.go",
		"make test",
		"make test-tui",
		"make build",
		"go test ./...",
		"go vet ./...",
		"go vet ./internal/app/commands/... ./internal/app/... ./internal/policy/... ./internal/tui/... 2>&1",
		"npm run test -- --runInBand",
		"npm run typecheck",
		"python -m pytest tests",
		"cargo check --workspace",
	} {
		decision := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":` + strconv.Quote(command) + `}`})
		if !decision.Allow || decision.RequiresApproval {
			t.Fatalf("expected no approval for %q: %+v", command, decision)
		}
	}
}

func TestDefaultToolPolicyDoesNotAutoAllowUnsafeShellVariants(t *testing.T) {
	p := DefaultToolPolicy{Mode: ApprovalModeOnRequest}
	spec := productionShellRunSpec(t)
	for _, command := range []string{
		"make test clean",
		"make build clean",
		"npm run lint -- --fix",
		"npx jest --updateSnapshot",
		"npx jest -u",
		"npx vitest run --update",
		"find . -delete",
		"find . -exec rm {} +",
		"find . -fprint out",
		"git diff --output=out.patch",
		"git diff --output=out.patch 2>&1",
		"git diff --output out.patch",
		"git diff --no-index /dev/null /etc/passwd",
		"git diff --no-index /dev/null ../secret.txt",
		"git diff 'feature;$(touch-pwn)...HEAD'",
		"git -c core.pager=cat diff",
		"git -C /tmp status --short",
		"git -C ../private diff",
		"git -C.. status --short",
		"git -C status --short",
		"cd /Users/goranka/Engineer/ai/dsk/whale-review-command && git status --short",
		"cd /Users/goranka/Engineer/ai/dsk/whale-review-command && git status --short 2>&1",
		"git branch -d feature",
		"git remote add origin git@example.com:repo.git",
		"git show --ext-diff HEAD",
		"git log --textconv",
		"git diff -- internal/policy/policy_test.go | sh",
		"git diff -- internal/policy/policy_test.go || tail -80",
		"git diff -- internal/policy/policy_test.go | tail -80 > out.txt",
		"git diff --output=out.patch | tail -80",
		"cd /Users/goranka/Engineer/ai/dsk/whale-review-command && git diff | tail -80",
		"rg --pre ./danger pattern",
		"go test ./... > out.txt",
		"go test ./... > out.txt 2>&1",
		"go test ./... 2> out.txt",
		"go test ./... '2>&1'",
		"go test ./...\nrm -rf /tmp/x",
	} {
		decision := p.Decide(spec, core.ToolCall{Name: "shell_run", Input: `{"command":` + strconv.Quote(command) + `}`})
		if !decision.Allow || !decision.RequiresApproval || decision.Code != "approval_required" {
			t.Fatalf("expected approval_required for %q: %+v", command, decision)
		}
	}
}

func TestDefaultToolPolicyNeverSkipsApprovalForMutatingTools(t *testing.T) {
	p := DefaultToolPolicy{Mode: ApprovalModeNever}
	tests := []struct {
		name string
		spec core.ToolSpec
		call core.ToolCall
	}{
		{
			name: "write",
			spec: core.ToolSpec{Name: "write"},
			call: core.ToolCall{Name: "write", Input: `{"file_path":"a.txt","content":"x"}`},
		},
		{
			name: "apply_patch",
			spec: core.ToolSpec{Name: "apply_patch"},
			call: core.ToolCall{Name: "apply_patch", Input: `{"patch":"*** Begin Patch\n*** End Patch\n"}`},
		},
		{
			name: "shell_run",
			spec: core.ToolSpec{Name: "shell_run"},
			call: core.ToolCall{Name: "shell_run", Input: `{"command":"go test ./..."}`},
		},
		{
			name: "mcp",
			spec: core.ToolSpec{Name: "mcp__github__create_issue"},
			call: core.ToolCall{Name: "mcp__github__create_issue", Input: `{}`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decision := p.Decide(tc.spec, tc.call)
			if !decision.Allow || decision.RequiresApproval || decision.Code != "auto_allow" {
				t.Fatalf("decision: %+v", decision)
			}
		})
	}
}

func TestDefaultToolPolicyNeverStillHonorsDenyPrefixes(t *testing.T) {
	p := DefaultToolPolicy{
		Mode:         ApprovalModeNever,
		DenyPrefixes: []string{"rm -rf"},
	}
	for _, command := range []string{
		"rm -rf /tmp/x",
		"rm -rf /tmp/x\necho done",
		"echo before\nrm -rf /tmp/x",
	} {
		decision := p.Decide(
			core.ToolSpec{Name: "shell_run"},
			core.ToolCall{Name: "shell_run", Input: `{"command":` + strconv.Quote(command) + `}`},
		)
		if decision.Allow || decision.Code != "policy_denied" || decision.MatchedRule != "rm -rf" {
			t.Fatalf("expected deny prefix for %q, got %+v", command, decision)
		}
	}
}

func TestShellCommandFromInput(t *testing.T) {
	if got := shellCommandFromInput(`{"command":" echo hi "}`); got != "echo hi" {
		t.Fatalf("shellCommandFromInput = %q, want %q", got, "echo hi")
	}
	if got := shellCommandFromInput(`{`); got != "" {
		t.Fatalf("shellCommandFromInput malformed = %q, want empty", got)
	}
}

func TestDefaultToolPolicyRequiresApprovalForMutatingCapability(t *testing.T) {
	decision := DefaultToolPolicy{Mode: ApprovalModeOnRequest}.Decide(
		core.ToolSpec{Name: "remember", Capabilities: []string{"mutates_state"}},
		core.ToolCall{Name: "remember", Input: `{}`},
	)
	if !decision.Allow || !decision.RequiresApproval || decision.Code != "approval_required" {
		t.Fatalf("decision: %+v", decision)
	}
}

func TestDefaultToolPolicyNeverAllowsMutatingCapability(t *testing.T) {
	decision := DefaultToolPolicy{Mode: ApprovalModeNever}.Decide(
		core.ToolSpec{Name: "remember", Capabilities: []string{"mutates_state"}},
		core.ToolCall{Name: "remember", Input: `{}`},
	)
	if !decision.Allow || decision.RequiresApproval || decision.Code != "auto_allow" {
		t.Fatalf("decision: %+v", decision)
	}
}

func TestApprovalMetadataPreservesToolPreviewValues(t *testing.T) {
	got := ApprovalMetadata(
		core.ToolCall{Name: "remember", Input: `{"scope":"global","name":"style"}`},
		[]string{"remember|x"},
		map[string]any{
			"approval_kind":          "memory_write",
			"approval_session_scope": "global memory: style",
			"memory_name":            "style",
		},
	)
	if got["approval_kind"] != "memory_write" {
		t.Fatalf("approval_kind overwritten: %+v", got)
	}
	if got["approval_session_scope"] != "global memory: style" {
		t.Fatalf("approval_session_scope overwritten: %+v", got)
	}
	if got["approval_scope"] != "workspace" {
		t.Fatalf("approval_scope default not set: %+v", got)
	}
	if got["memory_name"] != "style" {
		t.Fatalf("preview metadata lost: %+v", got)
	}
}

func productionShellRunSpec(t *testing.T) core.ToolSpec {
	t.Helper()
	ts, err := whaleTools.NewToolset(t.TempDir())
	if err != nil {
		t.Fatalf("new toolset: %v", err)
	}
	for _, tool := range ts.Tools() {
		if tool.Name() == "shell_run" {
			return core.DescribeTool(tool)
		}
	}
	t.Fatal("production shell_run tool not found")
	return core.ToolSpec{}
}
