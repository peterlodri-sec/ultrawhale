package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ── Disaster Recovery — ultrawhale doctor learns from PROBLEMS ────────
//
// The ASCII-stream gap taught us: rendered HTML docs can go missing.
// The doctor now detects and auto-heals missing docs.
//
// Template: the HTML wrapper that ALL docs use.
// Recovery: if a doc is missing, regenerate from its .md source.
// Doctor: /doctor now includes rendered-doc health check.

// DocTemplate is the sacred HTML template for all docs.
const DocTemplate = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s — ultrawhale docs</title>
<style>
:root { --bg: #0a0a14; --fg: #e0e8f5; --accent: #00d4ff; --green: #00e660; --dim: #6878a0; --card: #141420; }
* { margin:0; padding:0; box-sizing:border-box; }
body { background:var(--bg); color:var(--fg); font-family: system-ui, monospace; padding:2rem; max-width:800px; margin:0 auto; line-height:1.8; }
h1,h2 { color:var(--accent); } a { color:var(--accent); }
pre { background:var(--card); padding:1rem; border-radius:6px; overflow-x:auto; }
nav { margin-bottom:2rem; } nav a { margin-right:1rem; }
footer { margin-top:3rem; color:var(--dim); font-size:0.8rem; border-top:1px solid var(--card); padding-top:1rem; }
</style></head><body>
<nav><a href="/ultrawhale/">Home</a> <a href="/ultrawhale/docs/">Docs</a> <a href="/ultrawhale/book/">Book</a></nav>
<pre style="white-space:pre-wrap;font-family:system-ui,monospace;background:none;padding:0">
%s
</pre>
<footer>ultrawhale · %s · <a href="/ultrawhale/">Home</a></footer>
</body></html>`

// DisasterStats tracks recovery activity.
type DisasterStats struct {
	DocsChecked   int64
	DocsMissing   int64
	DocsRecovered int64
	DocsFailed    int64
}

var disasterStats DisasterStats

// ── Doctor Check: Rendered Docs ──────────────────────────────────────

// DoctorCheckDocs verifies all rendered HTML docs exist.
func DoctorCheckDocs() string {
	var report []string
	report = append(report, "DOCTOR: rendered-docs check")

	mdDocs := findMarkdownDocs()
	disasterStats.DocsChecked = int64(len(mdDocs))

	for _, mdFile := range mdDocs {
		name := strings.TrimSuffix(filepath.Base(mdFile), ".md")
		htmlFile := filepath.Join("rendered-docs", name+".html")

		if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
			disasterStats.DocsMissing++
			report = append(report, fmt.Sprintf("  ❌ %s.html MISSING — attempting recovery...", name))

			// Recovery: generate from template
			if recoverDoc(name, mdFile, htmlFile) {
				disasterStats.DocsRecovered++
				report = append(report, fmt.Sprintf("  ✅ %s.html RECOVERED", name))
			} else {
				disasterStats.DocsFailed++
				report = append(report, fmt.Sprintf("  ❌ %s.html RECOVERY FAILED", name))
			}
		}
	}

	if disasterStats.DocsMissing == 0 {
		report = append(report, fmt.Sprintf("  ✅ All %d docs present", len(mdDocs)))
	}

	return strings.Join(report, "\n")
}

func findMarkdownDocs() []string {
	var docs []string
	filepath.Walk("docs", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") { return nil }
		docs = append(docs, path)
		return nil
	})
	// Also check root .md files
	rootDocs, _ := filepath.Glob("*.md")
	for _, d := range rootDocs {
		docs = append(docs, d)
	}
	return docs
}

func recoverDoc(name, mdPath, htmlPath string) bool {
	content, err := os.ReadFile(mdPath)
	if err != nil { return false }

	title := strings.ReplaceAll(name, "-", " ")
	title = strings.Title(title)

	html := fmt.Sprintf(DocTemplate, title, string(content), CurrentVersion())

	if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil { return false }

	Log(LogInfo, "disaster.recover", name, "", "", 0, nil)
	return true
}

// ── Doctor Status ─────────────────────────────────────────────────────

// DisasterStatus returns compact disaster recovery status.
func DisasterStatus() string {
	return fmt.Sprintf("disaster-recovery: %d checked · %d missing · %d recovered · %d failed",
		disasterStats.DocsChecked, disasterStats.DocsMissing,
		disasterStats.DocsRecovered, disasterStats.DocsFailed)
}

// DisasterVakedFit returns disaster recovery Vaked fit.
func DisasterVakedFit() string {
	return `DISASTER RECOVERY = HEAL LAYER FOR DOCS

  The doctor learns from PROBLEMS.
  Missing HTML docs → auto-regenerated from .md template.
  The SACRED template IS the recovery spec.

  "Fold it into ultrawhale doctor. Learn from PROBLEMS." — Peter`
}
