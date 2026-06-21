package blocks

import (
	"fmt"
	"sync"
)

// ── Agent Capabilities ────────────────────────────────────────────────
// Vaked philosophy: either FULL (read/write/execute) or OBSERVE (read only).
// No intermediate modes — materializes Vaked's binary capability model.

// Capability is what an agent can do.
type Capability int

const (
	CapNone    Capability = 0
	CapRead    Capability = 1 << iota // 1: read files
	CapWrite                          // 2: write files
	CapExecute                        // 4: execute commands
	CapDelegate                       // 8: delegate to other agents
	CapSpawn                           // 16: spawn subagents/swarms
	CapEdge                            // 32: deploy to edge
)

// CapProfile defines an agent's capability set.
type CapProfile struct {
	Name string     // "FULL" or "OBSERVE"
	Caps Capability
}

var (
	// FULL: read + write + execute + delegate + spawn + edge
	CapFULL = CapProfile{Name: "FULL", Caps: CapRead | CapWrite | CapExecute | CapDelegate | CapSpawn | CapEdge}

	// OBSERVE: read only
	CapOBSERVE = CapProfile{Name: "OBSERVE", Caps: CapRead}
)

// Can checks if a profile has a specific capability.
func (c CapProfile) Can(cap Capability) bool {
	return c.Caps&cap != 0
}

// CapRegistry maps agent roles to capability profiles.
type CapRegistry struct {
	mu    sync.Mutex
	roles map[string]CapProfile
}

var capRegistry = &CapRegistry{roles: map[string]CapProfile{
	"swe":     CapFULL,
	"explore": CapOBSERVE,
	"review":  CapOBSERVE,
}}

// GetCapProfile returns the capability profile for a role.
func GetCapProfile(role string) CapProfile {
	_ = CurrentPOV()
	capRegistry.mu.Lock()
	defer capRegistry.mu.Unlock()
	if p, ok := capRegistry.roles[role]; ok { return p }
	return CapOBSERVE // default: observe only
}

// SetCapProfile sets the capability profile for a role.
func SetCapProfile(role string, profile CapProfile) {
	capRegistry.mu.Lock()
	defer capRegistry.mu.Unlock()
	capRegistry.roles[role] = profile
}

// CapStatus returns compact capability registry status.
func CapStatus() string {
	capRegistry.mu.Lock()
	defer capRegistry.mu.Unlock()
	return fmt.Sprintf("caps: %d roles (%d FULL, %d OBSERVE)",
		len(capRegistry.roles),
		countByCap(CapFULL.Caps),
		countByCap(CapOBSERVE.Caps))
}

func countByCap(c Capability) int {
	count := 0
	for _, p := range capRegistry.roles {
		if p.Caps == c { count++ }
	}
	return count
}
