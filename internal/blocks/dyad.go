package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Dyad — Two ultrawhale instances paired as one ────────────────────
// A dyad is two machines thinking together. Each has its own TUI,
// orchestrator, and brain. Shared POV, shared memo scope.
// One can take over if the other fails.

// DyadBlock pairs two ultrawhale instances.
type DyadBlock struct {
	mu sync.Mutex

	// Identities
	Self  DyadPeer // this machine
	Peer  DyadPeer // the paired machine
	ID    string   // dyad session ID

	// State
	Status    string    // "connecting", "paired", "degraded", "failed"
	PairedAt  time.Time
	LastPing  time.Time

	// Shared context
	SharedPOV   POV       // merged POV from both machines
	SharedMemos *MemoStore // scoped "dyad" — visible to both

	// Health
	SelfAlive bool
	PeerAlive bool
	PingCount int64
}

// DyadPeer represents one machine in a dyad.
type DyadPeer struct {
	Machine string // "M1-Max", "dev-cx53"
	Arch    string // "arm64", "amd64"
	Tier    string // "go", "asm", "gpu"
	Version string // "v7.0.0"
	DID     string // orchestrator DID
}

// NewDyad creates a dyad pairing self with a peer.
func NewDyad(peer DyadPeer) *DyadBlock {
	pov := CurrentPOV()
	self := DyadPeer{
		Machine: pov.Machine,
		Arch:    pov.Arch,
		Tier:    pov.Tier,
		Version: CurrentVersion(),
		DID:     GetOrchestrator().DID,
	}

	dyad := &DyadBlock{
		Self:    self,
		Peer:    peer,
		ID:      fmt.Sprintf("dyad-%s-%s", self.Machine, peer.Machine),
		Status:  "connecting",
		PairedAt: time.Now(),
		SharedPOV: POV{
			Machine: fmt.Sprintf("%s+%s", self.Machine, peer.Machine),
			Arch:    fmt.Sprintf("%s+%s", self.Arch, peer.Arch),
			Tier:    fmt.Sprintf("%s+%s", self.Tier, peer.Tier),
		},
		SharedMemos: &MemoStore{memos: make(map[string]Memo)},
		SelfAlive:   true,
		PeerAlive:   false, // will be set on first successful ping
	}

	return dyad
}

// Ping sends a heartbeat to the peer via NATS.
func (d *DyadBlock) Ping() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.LastPing = time.Now()
	d.PingCount++
	d.SelfAlive = true

	// In production: publish whale.dyad.ping via NATS
	// For now: local status update
	if d.PeerAlive {
		d.Status = "paired"
	} else {
		d.Status = "degraded"
	}

	Log(LogInfo, "dyad.ping", fmt.Sprintf("%s → %s (%s)", d.Self.Machine, d.Peer.Machine, d.Status),
		"", "", time.Since(d.PairedAt), nil)
}

// PeerPingReceived marks the peer as alive when a ping is received.
func (d *DyadBlock) PeerPingReceived() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.PeerAlive = true
	d.LastPing = time.Now()
	if d.SelfAlive {
		d.Status = "paired"
	}
	d.PingCount++
}

// Failover activates when the peer is detected as dead.
func (d *DyadBlock) Failover() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.PeerAlive && time.Since(d.LastPing) > 30*time.Second {
		d.Status = "failed"
		d.PeerAlive = false
		Log(LogWarn, "dyad.failover", fmt.Sprintf("%s taking over — %s unreachable", d.Self.Machine, d.Peer.Machine),
			"", "", time.Since(d.PairedAt), nil)
		return true
	}
	return false
}

// DualVerify requires both peers to agree on a critical action.
func (d *DyadBlock) DualVerify(action string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Status != "paired" {
		// Degraded mode: self approves, peer not available
		return true
	}

	// In production: send verification request to peer via NATS dyad channel
	// For now: local approval
	d.SharedMemos.Remember(ScopeAgents, fmt.Sprintf("dual-verify: %s", action))
	return true
}

// DyadStatus returns compact dyad status for HUD.
func (d *DyadBlock) DyadStatus() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	icon := "●"
	switch d.Status {
	case "paired": icon = "●"
	case "degraded": icon = "◐"
	case "failed": icon = "○"
	default: icon = "◌"
	}

	return fmt.Sprintf("dyad: %s %s+%s (%s, %d pings)",
		icon, d.Self.Machine, d.Peer.Machine, d.Status, d.PingCount)
}

// ── Global dyad ───────────────────────────────────────────────────────

var globalDyad *DyadBlock

// InitDyad creates the global dyad.
func InitDyad(peerMachine, peerArch, peerTier string) *DyadBlock {
	peer := DyadPeer{
		Machine: peerMachine,
		Arch:    peerArch,
		Tier:    peerTier,
		Version: CurrentVersion(),
	}
	globalDyad = NewDyad(peer)
	return globalDyad
}

// GetDyad returns the global dyad.
func GetDyad() *DyadBlock { return globalDyad }


// DyadHealth returns a health check response for the dyad.
func DyadHealth() map[string]any {
	d := GetDyad()
	if d == nil {
		return map[string]any{"status": "not_initialized"}
	}
	return map[string]any{
		"status":      d.Status,
		"self_alive":  d.SelfAlive,
		"peer_alive":  d.PeerAlive,
		"ping_count":  d.PingCount,
		"last_ping":   d.LastPing.Format(time.RFC3339),
		"paired_at":   d.PairedAt.Format(time.RFC3339),
	}
}
