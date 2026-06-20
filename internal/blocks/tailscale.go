package blocks

import (
	"fmt"
	"os/exec"
	"strings"
)

// ── Tailscale Primitive ───────────────────────────────────────────────
// First-class Tailscale integration. Auto-discovers nodes, resolves
// MagicDNS names, provides status for HUD/sidepanel.

// TailscaleStatus returns the Tailscale network status.
func TailscaleStatus() string {
	out, err := exec.Command("tailscale", "status").Output()
	if err != nil {
		return "tailscale: not available"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	online := 0
	for _, line := range lines[1:] { // skip header
		if strings.Contains(line, "active") || strings.Contains(line, "idle") {
			online++
		}
	}
	return fmt.Sprintf("tailscale: %d nodes online (%d total)", online, len(lines)-1)
}

// TailscaleNodes returns a list of online Tailscale nodes.
func TailscaleNodes() []string {
	out, err := exec.Command("tailscale", "status").Output()
	if err != nil { return nil }
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var nodes []string
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			nodes = append(nodes, fields[1]) // hostname
		}
	}
	return nodes
}

// TailscaleIP returns the Tailscale IP for a hostname.
func TailscaleIP(host string) string {
	out, err := exec.Command("tailscale", "ip", host).Output()
	if err != nil { return "" }
	return strings.TrimSpace(string(out))
}

// IsTailscaleUp returns true if Tailscale is running and connected.
func IsTailscaleUp() bool {
	return exec.Command("tailscale", "status").Run() == nil
}
