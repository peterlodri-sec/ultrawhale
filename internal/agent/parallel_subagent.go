package agent

import "github.com/usewhale/whale/internal/core"

const parallelSubagentToolName = "spawn_subagent"

type parallelSubagentGroup struct {
	Start int
	Calls []core.ToolCall
}

func eligibleParallelSubagentGroups(calls []core.ToolCall) []parallelSubagentGroup {
	var groups []parallelSubagentGroup
	for i := 0; i < len(calls); {
		if calls[i].Name != parallelSubagentToolName {
			i++
			continue
		}

		start := i
		for i < len(calls) && calls[i].Name == parallelSubagentToolName {
			i++
		}
		if i-start < 2 {
			continue
		}

		groupCalls := make([]core.ToolCall, i-start)
		copy(groupCalls, calls[start:i])
		groups = append(groups, parallelSubagentGroup{
			Start: start,
			Calls: groupCalls,
		})
	}
	return groups
}
