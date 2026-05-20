package app

import "strings"

type HelpCommand struct {
	Name        string
	Description string
}

func HelpCommands() []HelpCommand {
	return []HelpCommand{
		{Name: "/agent", Description: "Switch to agent mode"},
		{Name: "/ask [prompt]", Description: "Switch to ask mode, optionally submit a prompt"},
		{Name: "/clear", Description: "Clear the visible conversation"},
		{Name: "/compact", Description: "Compact the current conversation"},
		{Name: "/exit", Description: "Exit Whale"},
		{Name: "/feedback", Description: "Open the Whale issue tracker"},
		{Name: "/focus", Description: "Toggle focus view"},
		{Name: "/help", Description: "Show help and available commands"},
		{Name: "/init", Description: "Create AGENTS.md from repository context"},
		{Name: "/mcp", Description: "Show MCP server status"},
		{Name: "/memory", Description: "Manage memory entries"},
		{Name: "/model", Description: "Choose model, effort, and thinking settings"},
		{Name: "/new [id]", Description: "Start a new session"},
		{Name: "/permissions", Description: "Choose tool approval behavior"},
		{Name: "/plan [prompt]", Description: "Switch to plan mode, optionally submit a prompt"},
		{Name: "/plugins", Description: "Manage plugins"},
		{Name: "/resume", Description: "Open the resume picker"},
		{Name: "/review [target]", Description: "Open review mode or review a target"},
		{Name: "/skills", Description: "Show available skills"},
		{Name: "/stats", Description: "Show usage and tool statistics"},
		{Name: "/status", Description: "Show session and configuration status"},
	}
}

func BuildHelpText() string {
	var b strings.Builder
	b.WriteString("Whale help\n\n")
	b.WriteString("Browse default commands:\n\n")
	for i, cmd := range HelpCommands() {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- `")
		b.WriteString(cmd.Name)
		b.WriteString("`\n")
		b.WriteString("  ")
		b.WriteString(cmd.Description)
	}
	b.WriteString("\n\nFor more help: https://github.com/usewhale/whale")
	return b.String()
}
