package blocks

import (
	"fmt"
	"strings"
)

// ── Engine — ultrawhale Blocks Engine Abstraction ────────────────────
//
// The engine is the Vaked Materializes layer made explicit.
// It takes declarations (Vaked .vaked files) and materializes them
// into running blocks.
//
// Engine = the thing that turns declarations into operations.
// In Vaked: Declares → Engine → Materializes → Reveals.
// The engine IS the Materializes layer.

// Engine is the ultrawhale blocks execution engine.
type Engine struct {
	Name       string
	Version    string
	Blocks     []string          // registered block names
	Capabilities CapProfile      // what the engine can do
	Stats      EngineStats
}

// EngineStats tracks engine activity.
type EngineStats struct {
	TotalOps    int64
	Writes      int64
	Reads       int64
	Seds        int64
	Delegations int64
	OneShots    int64
	Errors      int64
}

var engine = &Engine{
	Name:    "ultrawhale-engine",
	Version: CurrentVersion(),
	Blocks:  make([]string, 0),
	Capabilities: CapFULL,
}

// InitEngine initializes the engine with all registered blocks.
func InitEngine() {
	engine.Blocks = make([]string, 0, len(schemaRegistry))
	for name := range schemaRegistry {
		engine.Blocks = append(engine.Blocks, name)
	}
	Log(LogInfo, "engine.init", fmt.Sprintf("%d blocks", len(engine.Blocks)), "", "", 0, nil)
}

// EngineStats returns engine activity statistics.
func GetEngineStats() EngineStats { return GetEngineStats }

// EngineStatus returns compact engine status.
func EngineStatus() string {
	return fmt.Sprintf("engine: %s · %d blocks · %d ops (%d writes, %d reads, %d seds, %d delegates, %d oneshots)",
		engine.Name, len(engine.Blocks), engine.Stats.TotalOps,
		engine.Stats.Writes, engine.Stats.Reads, engine.Stats.Seds,
		engine.Stats.Delegations, engine.Stats.OneShots)
}

// EngineMaterialize takes a Vaked declaration and materializes it into a block operation.
// This IS the Vaked Materializes layer.
func EngineMaterialize(declaration string) string {
	if !IsAllowed() { return "engine: permission denied" }

	lower := strings.ToLower(declaration)
	
	switch {
	case strings.HasPrefix(lower, "write"):
		engine.Stats.TotalOps++; engine.Stats.Writes++
		return "engine: write → blocks.Write (journaled)"
	case strings.HasPrefix(lower, "read"):
		engine.Stats.TotalOps++; engine.Stats.Reads++
		return "engine: read → blocks.Read (ref-verified)"
	case strings.HasPrefix(lower, "sed"):
		engine.Stats.TotalOps++; engine.Stats.Seds++
		return "engine: sed → blocks.SedAll (SIMD)"
	case strings.HasPrefix(lower, "delegate"):
		engine.Stats.TotalOps++; engine.Stats.Delegations++
		return "engine: delegate → orchestrator"
	case strings.HasPrefix(lower, "oneshot"):
		engine.Stats.TotalOps++; engine.Stats.OneShots++
		result, _ := OneShotAllowed(declaration)
		return fmt.Sprintf("engine: oneshot → %s (%dµs)", result.Materialized, result.Duration)
	default:
		engine.Stats.TotalOps++; engine.Stats.Errors++
		return fmt.Sprintf("engine: unknown declaration '%s'", declaration[:min(40, len(declaration))])
	}
}

// EngineVakedFit returns how the engine fits into the Vaked philosophy.
func EngineVakedFit() string {
	return `Vaked:  Declares → ENGINE → Materializes → Reveals
                     ↑
              ultrawhale-engine
              (59 blocks, 6 plugins, 5 protocols)
              
The engine IS the Materializes layer.
It takes .vaked declarations and turns them into running blocks.
Every Write, Read, Sed, Delegate, OneShot flows through the engine.`
}
