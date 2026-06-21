package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── Obsidian — First-Class Vault Integration ─────────────────────────
//
// Obsidian is the human's knowledge graph. ultrawhale is the machine's.
// The auto-dev-wiki bridges them: every session, every brainstorm,
// every architecture decision → Obsidian vault as a markdown note.
//
// Vaked fit: Indexes layer. The Obsidian vault IS the CrabCC for human thought.
// Space: the vault path is a space node in the topology graph.
// Context: every note carries POV + session context.

// ObsidianVault is a connected Obsidian vault.
type ObsidianVault struct {
	Name      string
	Path      string // ~/Documents/Knowledge or iCloud path
	Connected bool
	Stats     ObsidianStats
}

// ObsidianStats tracks vault activity.
type ObsidianStats struct {
	NotesCreated   int64
	NotesUpdated   int64
	LinksCreated   int64
	BacklinksFound int64
	LastSync       time.Time
}

// ObsidianNote is one markdown note in the vault.
type ObsidianNote struct {
	Title      string
	Path       string
	Tags       []string
	Links      []string // [[wiki-links]]
	Backlinks  []string // pages that link here
	Content    string
	ModifiedAt time.Time
}

var obsidianVaults = make(map[string]*ObsidianVault)

// ── Vault Operations ─────────────────────────────────────────────────

// ConnectObsidianVault connects to an Obsidian vault at the given path.
func ConnectObsidianVault(name, path string) (*ObsidianVault, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("obsidian: vault not found at %s", path)
	}

	vault := &ObsidianVault{
		Name:      name,
		Path:      path,
		Connected: true,
	}
	obsidianVaults[name] = vault

	Log(LogInfo, "obsidian.connect", fmt.Sprintf("%s → %s", name, path), "", "", 0, nil)

	// Auto-discover: scan for .md files
	go vault.scanVault()

	return vault, nil
}

// WriteNote creates or updates a note in the vault.
func (v *ObsidianVault) WriteNote(title, content string, tags, links []string) (*ObsidianNote, error) {
	v.Stats.NotesCreated++

	// Build markdown with frontmatter
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: \"%s\"\n", title))
	sb.WriteString(fmt.Sprintf("created: \"%s\"\n", time.Now().Format("2006-01-02T15:04:05")))
	sb.WriteString(fmt.Sprintf("source: \"ultrawhale %s\"\n", CurrentVersion()))
	sb.WriteString(fmt.Sprintf("pov: \"%s/%s/%s\"\n", CurrentPOV().Machine, CurrentPOV().Arch, CurrentPOV().Tier))
	if len(tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(tags, ", ")))
	}
	sb.WriteString("---\n\n")
	sb.WriteString(content)

	if len(links) > 0 {
		sb.WriteString("\n\n## Related\n\n")
		for _, link := range links {
			sb.WriteString(fmt.Sprintf("- [[%s]]\n", link))
		}
	}

	// Sanitize filename
	filename := strings.ReplaceAll(strings.ToLower(title), " ", "-")
	filename = strings.ReplaceAll(filename, "/", "-")
	filePath := filepath.Join(v.Path, filename+".md")

	if err := os.WriteFile(filePath, []byte(sb.String()), 0o644); err != nil {
		return nil, fmt.Errorf("obsidian write: %w", err)
	}

	v.Stats.LastSync = time.Now()
	Log(LogInfo, "obsidian.write", filePath, "", "", 0, nil)

	return &ObsidianNote{
		Title:      title,
		Path:       filePath,
		Tags:       tags,
		Links:      links,
		Content:    content,
		ModifiedAt: time.Now(),
	}, nil
}

// AutoDevWiki writes a session summary to the vault.
func (v *ObsidianVault) AutoDevWiki(sessionID, topic, summary string) (*ObsidianNote, error) {
	return v.WriteNote(
		fmt.Sprintf("dev-%s-%s", time.Now().Format("20060102"), topic),
		summary,
		[]string{"ultrawhale", "dev-wiki", "session"},
		[]string{"ultrawhale-architecture", "vaked-pipeline", topic},
	)
}

// scanVault discovers existing notes and backlinks.
func (v *ObsidianVault) scanVault() {
	filepath.Walk(v.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") { return nil }

		content, err := os.ReadFile(path)
		if err != nil { return nil }

		// Count wiki-links [[...]]
		links := countWikiLinks(string(content))
		v.Stats.LinksCreated += int64(links)
		v.Stats.BacklinksFound += int64(links)

		return nil
	})
}

func countWikiLinks(content string) int {
	count := 0
	for i := 0; i < len(content)-3; i++ {
		if content[i:i+2] == "[[" {
			count++
		}
	}
	return count
}

// ── Auto-Dev-Wiki Integration ─────────────────────────────────────────

// StartAutoDevWiki begins automatic session logging to Obsidian.
func StartAutoDevWiki(vaultName, topic string) {
	vault, ok := obsidianVaults[vaultName]
	if !ok {
		Log(LogWarn, "obsidian.autodev", "vault not connected", "", "", 0, nil)
		return
	}

	// Write session start
	vault.AutoDevWiki("session", topic, fmt.Sprintf(
		"# Auto-Dev-Wiki: %s\n\n**Started:** %s\n**Version:** %s\n**POV:** %s/%s",
		topic, time.Now().Format(time.RFC3339),
		CurrentVersion(), CurrentPOV().Machine, CurrentPOV().Arch,
	))

	Log(LogInfo, "obsidian.autodev", topic, "", "", 0, nil)
}

// ── Status ────────────────────────────────────────────────────────────

// ObsidianStatus returns compact vault status.
func ObsidianStatus() string {
	if len(obsidianVaults) == 0 {
		return "obsidian: no vaults connected"
	}

	var lines []string
	for name, v := range obsidianVaults {
		status := "connected"
		if !v.Connected { status = "disconnected" }
		lines = append(lines, fmt.Sprintf("  %s: %s · %d notes · %d links · last: %s",
			name, status, v.Stats.NotesCreated, v.Stats.LinksCreated,
			v.Stats.LastSync.Format("15:04:05")))
	}
	return "obsidian:\n" + strings.Join(lines, "\n")
}

// ObsidianVakedFit returns Obsidian's Vaked fit.
func ObsidianVakedFit() string {
	return `OBSIDIAN = INDEXES LAYER (human knowledge graph)

  The Obsidian vault IS the CrabCC for human thought.
  Space: vault path is a space topology node.
  Context: every note carries POV + session context.
  Time: auto-dev-wiki timestamps every decision.

  Method: claude-obsidian:github — the bridge between
  human knowledge (Obsidian) and machine knowledge (ultrawhale).`
}

// AutoDiscoverObsidian attempts to find Obsidian vaults.
func AutoDiscoverObsidian() string {
	paths := []string{
		os.ExpandEnv("$HOME/Documents/Knowledge"),
		os.ExpandEnv("$HOME/Library/Mobile Documents/iCloud~md~obsidian/Documents"),
		os.ExpandEnv("$HOME/Obsidian"),
	}

	found := 0
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			name := filepath.Base(p)
			ConnectObsidianVault(name, p)
			found++
		}
	}

	return fmt.Sprintf("obsidian: auto-discovered %d vaults", found)
}
