package blocks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ── Export Primitive ──────────────────────────────────────────────────
// Session export to JSON, Markdown, or HTML.

// ExportSession exports the current session state to a file.
func ExportSession(format, dir string) (string, error) {
	ts := time.Now().Format("20060102-150405")

	switch format {
	case "json":
		return exportJSON(dir, ts)
	case "markdown", "md":
		return exportMarkdown(dir, ts)
	case "html":
		return exportHTML(dir, ts)
	default:
		return "", fmt.Errorf("export: unknown format %s (use json, markdown, html)", format)
	}
}

func exportJSON(dir, ts string) (string, error) {
	path := filepath.Join(dir, fmt.Sprintf("ultrawhale-session-%s.json", ts))
	pov := CurrentPOV()
	export := map[string]any{
		"version":   CurrentVersion(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"pov":       pov,
		"brain":     GetBrain().BrainDump(),
		"ralph":     GetRalph().RalphStatus(),
		"tools":     ToolStatus(),
	}
	data, _ := json.MarshalIndent(export, "", "  ")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	Log(LogInfo, "blocks.Export", path, Ref(data), "", 0, nil)
	return path, nil
}

func exportMarkdown(dir, ts string) (string, error) {
	path := filepath.Join(dir, fmt.Sprintf("ultrawhale-session-%s.md", ts))
	pov := CurrentPOV()
	md := fmt.Sprintf(`# ultrawhale Session Export

**Date:** %s
**Version:** %s
**POV:** %s

## Brain
%s

## Tools
%s
`, time.Now().UTC().Format(time.RFC3339), CurrentVersion(), pov.String(),
		GetBrain().BrainDump(), ToolStatus())
	if err := os.WriteFile(path, []byte(md), 0o644); err != nil {
		return "", err
	}
	Log(LogInfo, "blocks.Export", path, Ref([]byte(md)), "", 0, nil)
	return path, nil
}

func exportHTML(dir, ts string) (string, error) {
	path := filepath.Join(dir, fmt.Sprintf("ultrawhale-session-%s.html", ts))
	html := fmt.Sprintf(`<!DOCTYPE html><html><head><title>ultrawhale Session</title></head><body><pre>%s</pre></body></html>`,
		GetBrain().BrainDump())
	if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
