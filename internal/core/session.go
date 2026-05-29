package core

import "strings"

const (
	// ToolInputEventsSuffix is the filename suffix for tool input event logs.
	ToolInputEventsSuffix = ".tool_input_events.jsonl"
	// ApprovalEventsSuffix is the filename suffix for approval event logs.
	ApprovalEventsSuffix = ".approval_events.jsonl"
)

// IsSessionJSONLName reports whether name is a session JSONL file (not tool input or approval events).
func IsSessionJSONLName(name string) bool {
	return strings.HasSuffix(name, ".jsonl") &&
		!strings.HasSuffix(name, ToolInputEventsSuffix) &&
		!strings.HasSuffix(name, ApprovalEventsSuffix)
}
