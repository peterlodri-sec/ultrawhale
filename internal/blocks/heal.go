package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Heal Primitive — Self-Repair-and-Healing ──────────────────────────
//
// The system watches itself. When it detects degradation, it heals.
// Self-healing parameters: what to watch, how often, what to do.
//
// Heal = observe → detect → repair → verify → log.

// HealCheck is one self-healing check.
type HealCheck struct {
	Name     string
	Watch    func() bool    // returns true if healthy
	Repair   func()         // called when unhealthy
	Interval time.Duration  // how often to check
	Cooldown time.Duration  // minimum time between repairs
	Enabled  bool
}

// Healer runs self-healing checks on a schedule.
type Healer struct {
	mu      sync.Mutex
	checks  map[string]*HealCheck
	stats   HealStats
	running bool
}

// HealStats tracks healing activity.
type HealStats struct {
	ChecksRun    int64
	FaultsFound  int64
	RepairsDone  int64
	RepairFails  int64
	LastHeal     time.Time
}

var healer = &Healer{checks: make(map[string]*HealCheck)}

// RegisterHealCheck adds a self-healing check.
func RegisterHealCheck(name string, watch func() bool, repair func(), interval, cooldown time.Duration) {
	healer.mu.Lock()
	defer healer.mu.Unlock()
	healer.checks[name] = &HealCheck{
		Name: name, Watch: watch, Repair: repair,
		Interval: interval, Cooldown: cooldown, Enabled: true,
	}
}

// StartHealing begins the self-healing loop.
func StartHealing() {
	healer.mu.Lock()
	if healer.running { healer.mu.Unlock(); return }
	healer.running = true
	healer.mu.Unlock()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			healer.mu.Lock()
			now := time.Now()
			for _, c := range healer.checks {
				if !c.Enabled { continue }
				healer.stats.ChecksRun++
				if !c.Watch() {
					healer.stats.FaultsFound++
					if now.Sub(healer.stats.LastHeal) > c.Cooldown {
						c.Repair()
						healer.stats.RepairsDone++
						healer.stats.LastHeal = now
						Log(LogWarn, "heal."+c.Name, "repair triggered", "", "", 0, nil)
					}
				}
			}
			healer.mu.Unlock()
		}
	}()

	Log(LogInfo, "heal.start", fmt.Sprintf("%d checks", len(healer.checks)), "", "", 0, nil)
}

// Built-in self-healing checks
func init() {
	// Heal: restart supervisor if no agents are running
	RegisterHealCheck("supervisor-liveness",
		func() bool { return AgentCount() > 0 || GetOrchestrator().TotalTurns == 0 },
		func() { GetSupervisor().ReportFailure("system") },
		30*time.Second, 60*time.Second,
	)

	// Heal: clear brainstorm sessions stuck for >1h
	RegisterHealCheck("brainstorm-gc",
		func() bool {
			sessions := ListBrainstorms()
			for _, s := range sessions {
				if s.Status == "active" && time.Since(s.UpdatedAt) > 1*time.Hour {
					return false
				}
			}
			return true
		},
		func() { StartBrainstormGC() },
		5*time.Minute, 10*time.Minute,
	)

	// Heal: permission re-check — if revoked, don't auto-heal
	RegisterHealCheck("permission-integrity",
		func() bool { return IsAllowed() || PermissionState(permissionGate.Load()) == PermUnset },
		func() {}, // no auto-heal for permissions
		60*time.Second, 0,
	)
}

// HealStatus returns compact healing status.
func HealStatus() string {
	healer.mu.Lock()
	defer healer.mu.Unlock()
	return fmt.Sprintf("heal: %d checks · %d run · %d faults · %d repairs · last: %s",
		len(healer.checks), healer.stats.ChecksRun,
		healer.stats.FaultsFound, healer.stats.RepairsDone,
		healer.stats.LastHeal.Format("15:04:05"))
}

// HealEngineVakedFit returns the healing engine's Vaked fit.
func HealEngineVakedFit() string {
	return `Heal IS the self-repair layer.
It watches ALL 7 engines and repairs degradation.

  Declares → Engine → Supervise → Enforce → Testify → Index → Reveal
     ↑         ↑         ↑          ↑         ↑        ↑       ↑
     └─────────┴─────────┴──────────┴─────────┴────────┴───────┘
                              HEAL watches all`
}
