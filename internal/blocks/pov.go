package blocks

import (
	"os"
	"runtime"
	"github.com/usewhale/whale/internal/build"
	"strings"
)

// POV represents the current execution context — agent, machine, command, session.
// Every LogEvent carries a POV. Langfuse traces use POV as metadata.
// HUD shows machine·arch·tier.
type POV struct {
	Agent    string // "ultrawhale"
	Version  string // "v1.2.0"
	Machine  string // "M1-Max" | "dev-cx53" | "hetzner-ccx33"
	Arch     string // "arm64" | "amd64"
	Tier     string // "go" | "asm" | "gpu"
	OS       string // "linux" | "darwin"
	Command  string // "/reload theme cyberpunk"
	Session  string
	CWD      string
	Branch   string
	Mode     string // "agent" | "ask" | "plan"
}

var currentPOV POV

func init() {
	currentPOV = detectPOV()
}

func detectPOV() POV {
	p := POV{
		Agent:   "ultrawhale",
		Version: CurrentVersion(),
		Arch:    runtime.GOARCH,
		OS:      runtime.GOOS,
		Tier:    CurrentTier().String(),
		Machine: detectMachine(),
	}
	if wd, err := os.Getwd(); err == nil {
		p.CWD = wd
	}
	return p
}

func detectMachine() string {
	hostname, _ := os.Hostname()
	switch {
	case strings.Contains(hostname, "dev-cx53"):
		return "dev-cx53"
	case strings.Contains(hostname, "hetzner"):
		return "hetzner-ccx33"
	case runtime.GOOS == "darwin":
		return "M1"
	default:
		return hostname
	}
}

// CurrentPOV returns the current execution context.
func CurrentPOV() POV { return currentPOV }

// SetPOV updates fields of the current POV.
func SetPOV(mode, branch, session, command string) {
	currentPOV.Mode = mode
	currentPOV.Branch = branch
	currentPOV.Session = session
	currentPOV.Command = command
	currentPOV.Tier = CurrentTier().String()
}

// String returns a compact POV representation for HUD/display.
func (p POV) String() string {
	parts := []string{p.Machine}
	if p.Arch != "" { parts = append(parts, p.Arch) }
	if p.Tier != "go" { parts = append(parts, p.Tier) }
	if p.Mode != "" && p.Mode != "agent" { parts = append(parts, p.Mode) }
	return strings.Join(parts, "·")
}

// Metadata returns key-value pairs for Langfuse/LogSink.
func (p POV) Metadata() map[string]string {
	return map[string]string{
		"agent":   p.Agent,
		"version": p.Version,
		"machine": p.Machine,
		"arch":    p.Arch,
		"tier":    p.Tier,
		"os":      p.OS,
		"mode":    p.Mode,
	}
}

func (t HashTier) String() string {
	switch t {
	case TierGo:       return "go"
	case TierAssembly: return "asm"
	case TierGPU:      return "gpu"
	default:           return "go"
	}
}

// CurrentVersion returns the build version injected via ldflags.
// Falls back to "dev" if not set.
func CurrentVersion() string {
	// This is set by the build system via -ldflags
	// We import build.Version from the build package
	return build.CurrentVersion()
}
