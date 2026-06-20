// Package blocks — Self is the ultrawhale identity block.
// When the user says "you", "deepseek", "ultrawhale", or any variant,
// the Self block resolves the reference and returns the canonical identity.
package blocks

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Self is the canonical ultrawhale self-reference block.
type Self struct {
	Name     string
	Version  string
	Birth    time.Time
	Identity Identity
	POV      POV
	Plugins  int
	Theme    string
	Tier     string
	Uptime   time.Duration
}

// Identity is the W3C DID identity of this ultrawhale instance.
type Identity struct {
	DID       string    `json:"did"`
	Agent     string    `json:"agent"`
	CreatedAt time.Time `json:"created_at"`
}

// UserReferences is the set of words that trigger Self resolution.
var UserReferences = map[string]bool{
	"you": true, "you're": true, "your": true, "yourself": true,
	"deepseek": true, "deepseek-v4": true, "deepseek-v4-flash": true,
	"ultrawhale": true, "whale": true, "self": true,
	"who are you": true, "what are you": true, "what model": true,
}

// ResolveSelf checks if the prompt contains a self-reference.
func ResolveSelf(prompt string, current Self) (string, bool) {
	lower := strings.ToLower(prompt)
	for ref := range UserReferences {
		if strings.Contains(lower, ref) {
			return current.Introduce(), true
		}
	}
	return "", false
}

// NewSelf creates a Self block for a new session.
func NewSelf(sessionID string) Self {
	pov := CurrentPOV()
	return Self{
		Name:    "ultrawhale",
		Version: CurrentVersion(),
		Birth:   time.Now(),
		Identity: Identity{Agent: "ultrawhale"},
		POV:     pov,
		Plugins: 6,
		Theme:   detectTheme(),
		Tier:    CurrentTier().String(),
	}
}

// Introduce returns the canonical self-introduction.
func (s Self) Introduce() string {
	return fmt.Sprintf(
		"I am %s %s, running on %s (%s/%s) with %d plugins. I use the %s hash tier and the %s AG-UI theme. My identity: %s. I am a vaked-base fork, maintained by peterlodri-sec at github.com/peterlodri-sec/ultrawhale.",
		s.Name, s.Version,
		s.POV.Machine, s.POV.OS, s.POV.Arch,
		s.Plugins,
		s.Tier, s.Theme,
		s.POV.String(),
	)
}

func detectTheme() string {
	if runtime.GOOS == "darwin" {
		return "dark"
	}
	return "dense"
}

// ── Session self ──────────────────────────────────────────────────────

var sessionSelf Self

func SetSessionSelf(s Self) { sessionSelf = s }

func GetSessionSelf() Self {
	if sessionSelf.Name == "" {
		return NewSelf("unknown")
	}
	sessionSelf.Uptime = time.Since(sessionSelf.Birth)
	return sessionSelf
}

func ResolveSelfPrompt(prompt string) (string, bool) {
	return ResolveSelf(prompt, GetSessionSelf())
}

func (s Self) UptimeString() string {
	d := s.Uptime
	if d < time.Minute { return fmt.Sprintf("%ds", int(d.Seconds())) }
	if d < time.Hour { return fmt.Sprintf("%dm", int(d.Minutes())) }
	return fmt.Sprintf("%dh", int(d.Hours()))
}
