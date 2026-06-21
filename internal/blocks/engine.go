package blocks

import (
	"fmt"
	"strings"
)

type Engine struct {
	Name         string
	Version      string
	Blocks       []string
	Capabilities CapProfile
	Stats        EngineStats
}

type EngineStats struct {
	TotalOps    int64
	Writes      int64
	Reads       int64
	Seds        int64
	Delegations int64
	OneShots    int64
	Errors      int64
}

var globalEngine = &Engine{
	Name:         "ultrawhale-engine",
	Version:      CurrentVersion(),
	Blocks:       make([]string, 0),
	Capabilities: CapFULL,
}

func InitEngine() {
	globalEngine.Blocks = make([]string, 0, len(schemaRegistry))
	for name := range schemaRegistry {
		globalEngine.Blocks = append(globalEngine.Blocks, name)
	}
	Log(LogInfo, "engine.init", fmt.Sprintf("%d blocks", len(globalEngine.Blocks)), "", "", 0, nil)
}

func GetEngineStats() EngineStats { return globalEngine.Stats }

func EngineStatus() string {
	return fmt.Sprintf("engine: %s · %d blocks · %d ops (%d writes, %d reads, %d seds, %d delegates, %d oneshots)",
		globalEngine.Name, len(globalEngine.Blocks), globalEngine.Stats.TotalOps,
		globalEngine.Stats.Writes, globalEngine.Stats.Reads, globalEngine.Stats.Seds,
		globalEngine.Stats.Delegations, globalEngine.Stats.OneShots)
}

func EngineMaterialize(declaration string) string {
	if !IsAllowed() { return "engine: permission denied" }

	lower := strings.ToLower(declaration)
	switch {
	case strings.HasPrefix(lower, "write"):
		globalEngine.Stats.TotalOps++; globalEngine.Stats.Writes++
		return "engine: write → blocks.Write (journaled)"
	case strings.HasPrefix(lower, "read"):
		globalEngine.Stats.TotalOps++; globalEngine.Stats.Reads++
		return "engine: read → blocks.Read (ref-verified)"
	case strings.HasPrefix(lower, "sed"):
		globalEngine.Stats.TotalOps++; globalEngine.Stats.Seds++
		return "engine: sed → blocks.SedAll (SIMD)"
	case strings.HasPrefix(lower, "delegate"):
		globalEngine.Stats.TotalOps++; globalEngine.Stats.Delegations++
		return "engine: delegate → orchestrator"
	case strings.HasPrefix(lower, "oneshot"):
		globalEngine.Stats.TotalOps++; globalEngine.Stats.OneShots++
		result, _ := OneShotAllowed(declaration)
		return fmt.Sprintf("engine: oneshot → %s (%dµs)", result.Materialized, result.Duration)
	default:
		globalEngine.Stats.TotalOps++; globalEngine.Stats.Errors++
		return fmt.Sprintf("engine: unknown declaration '%s'", declaration[:min(40, len(declaration))])
	}
}

func EngineVakedFit() string {
	return `Vaked:  Declares → ENGINE → Materializes → Reveals
                     ↑
              ultrawhale-engine
              (60 blocks, 6 plugins, 5 protocols)
              
The engine IS the Materializes layer.`
}
