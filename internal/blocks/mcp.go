package blocks

import (
	"fmt"
)

// ── MCP Server — Model Context Protocol ───────────────────────────────
// v20: Industry standard protocol for tool/resource/prompt exposure.
// Agents expose their capabilities as MCP tools.

// MCPServer exposes ultrawhale blocks as MCP tools.
type MCPServer struct {
	Name    string
	Version string
	Tools   []MCPTool
}

// MCPTool is a tool exposed via MCP.
type MCPTool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// MCPListTools returns all registered MCP tools.
func MCPListTools() []MCPTool {
	return []MCPTool{
		{Name: "blocks_write", Description: "Write a file (journaled, rollback-able)", InputSchema: map[string]any{"path": "string", "content": "string"}},
		{Name: "blocks_read", Description: "Read a file with ref verification", InputSchema: map[string]any{"path": "string"}},
		{Name: "blocks_sed", Description: "SIMD-accelerated find-and-replace", InputSchema: map[string]any{"find": "string", "replace": "string"}},
		{Name: "orchestrator_delegate", Description: "Delegate a task to a subagent", InputSchema: map[string]any{"prompt": "string"}},
		{Name: "brain_recall", Description: "Recall short-term memory", InputSchema: map[string]any{"n": "integer"}},
		{Name: "vfs_ls", Description: "List VFS directory", InputSchema: map[string]any{"path": "string"}},
		{Name: "vaked_parse", Description: "Parse a .vaked file", InputSchema: map[string]any{"path": "string"}},
	}
}

// MCPStatus returns compact MCP status.
func MCPStatus() string {
	return fmt.Sprintf("mcp: %d tools exposed", len(MCPListTools()))
}

// MCPInitialize returns the server capabilities.
func MCPInitialize() map[string]any {
	return map[string]any{
		"protocol": "2024-11-05",
		"server":   map[string]string{"name": "ultrawhale", "version": CurrentVersion()},
		"capabilities": map[string]any{
			"tools":     map[string]bool{"listChanged": true},
			"resources": map[string]bool{"subscribe": false},
		},
	}
}
