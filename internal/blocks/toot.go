package blocks

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// ── TOOT — Mastodon Bot Integration ──────────────────────────────────
//
// Posts to Mastodon via the vakedbot account.
// Token can come from env var (CI) or CF Pages secret (webhook).

type Toot struct {
	Content   string
	MediaURLs []string
	Timestamp time.Time
	Posted    bool
	ID        string
}

var tootHistory = make([]Toot, 0, 32)

// TootPost sends a status update to Mastodon via vakedbot.
func TootPost(content string) string {
	token := os.Getenv("MASTODON_ACCESS_TOKEN")
	if token == "" {
		return "❌ MASTODON_ACCESS_TOKEN not set"
	}

	// Truncate to Mastodon's 500 char limit
	if len(content) > 470 {
		content = content[:467] + "..."
	}

	// Add hashtags
	if !strings.Contains(content, "#ultrawhale") {
		content += "\n\n#ultrawhale #vaked #opensource"
	}

	// Post via Mastodon API
	url := "https://social.crabcc.app/api/v1/statuses"
	body := fmt.Sprintf(`{"status":%q}`, content)
	
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", fmt.Sprintf("ultrawhale-%d", time.Now().Unix()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("❌ Toot failed: %v", err)
	}
	defer resp.Body.Close()

	toot := Toot{
		Content:   content[:min(40, len(content))],
		Timestamp: time.Now(),
		Posted:    resp.StatusCode == 200,
	}
	tootHistory = append(tootHistory, toot)

	if resp.StatusCode == 200 {
		return fmt.Sprintf("✅ Tooted: %s... (#%d)", content[:40], len(tootHistory))
	}
	return fmt.Sprintf("❌ HTTP %d", resp.StatusCode)
}

// TootRelease posts a release announcement.
func TootRelease(version string) string {
	msg := fmt.Sprintf("🚀 ultrawhale %s released!\n\n148 blocks · 7 recursions · 14 protocols\nSPACE+TIME PROOF · RE+Audit Pipeline\n\nhttps://github.com/peterlodri-sec/ultrawhale/releases/tag/%s", version, version)
	return TootPost(msg)
}

// TootStatus returns toot history.
func TootStatus() string {
	return fmt.Sprintf("toot: %d posted · last: %s",
		len(tootHistory),
		func() string {
			if len(tootHistory) > 0 {
				return tootHistory[len(tootHistory)-1].Timestamp.Format("15:04:05")
			}
			return "never"
		}())
}
