package agent

import (
	"testing"

	"github.com/usewhale/whale/internal/core"
)

func TestEligibleParallelSubagentGroupsConsecutiveSpawnSubagents(t *testing.T) {
	calls := []core.ToolCall{
		{ID: "1", Name: "spawn_subagent"},
		{ID: "2", Name: "spawn_subagent"},
		{ID: "3", Name: "spawn_subagent"},
	}

	groups := eligibleParallelSubagentGroups(calls)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Start != 0 {
		t.Fatalf("expected group start 0, got %d", groups[0].Start)
	}
	if len(groups[0].Calls) != 3 {
		t.Fatalf("expected 3 calls in group, got %d", len(groups[0].Calls))
	}
}

func TestEligibleParallelSubagentGroupsMixedToolsCreateBoundaries(t *testing.T) {
	calls := []core.ToolCall{
		{ID: "1", Name: "spawn_subagent"},
		{ID: "2", Name: "read_file"},
		{ID: "3", Name: "spawn_subagent"},
		{ID: "4", Name: "spawn_subagent"},
		{ID: "5", Name: "shell"},
		{ID: "6", Name: "spawn_subagent"},
		{ID: "7", Name: "spawn_subagent"},
		{ID: "8", Name: "apply_patch"},
		{ID: "9", Name: "todo_add"},
		{ID: "10", Name: "request_user_input"},
		{ID: "11", Name: "spawn_subagent"},
		{ID: "12", Name: "spawn_subagent"},
	}

	groups := eligibleParallelSubagentGroups(calls)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Start != 2 || len(groups[0].Calls) != 2 {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	if groups[1].Start != 5 || len(groups[1].Calls) != 2 {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
	if groups[2].Start != 10 || len(groups[2].Calls) != 2 {
		t.Fatalf("unexpected third group: %+v", groups[2])
	}
}

func TestEligibleParallelSubagentGroupsParallelReasonIsBoundary(t *testing.T) {
	calls := []core.ToolCall{
		{ID: "1", Name: "spawn_subagent"},
		{ID: "2", Name: "parallel_reason"},
		{ID: "3", Name: "spawn_subagent"},
		{ID: "4", Name: "spawn_subagent"},
	}

	groups := eligibleParallelSubagentGroups(calls)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Start != 2 || len(groups[0].Calls) != 2 {
		t.Fatalf("unexpected group: %+v", groups[0])
	}
}

func TestEligibleParallelSubagentGroupsRequiresAtLeastTwoReadyCalls(t *testing.T) {
	calls := []core.ToolCall{
		{ID: "1", Name: "spawn_subagent"},
		{ID: "2", Name: "read_file"},
		{ID: "3", Name: "spawn_subagent"},
	}

	groups := eligibleParallelSubagentGroups(calls)
	if len(groups) != 0 {
		t.Fatalf("expected no groups for single ready spawn_subagent calls, got %+v", groups)
	}
}
