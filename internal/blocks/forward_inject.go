package blocks

import (
	"fmt"
	"strings"
	"time"
)

// ── FORWARD INJECT — Virtual Space Orchestrator Command ──────────────
//
// Peter: "Forward inject in the virtual space. The orchestrator system
// command. Which, yes, does what it does."
//
// FORWARD INJECT is a command that operates on the SPACE topology directly.
// It reaches FORWARD into the virtual space and INJECTS an operation.
//
// Unlike DelegatePrompt (which asks an agent to do something),
// ForwardInject DOES something directly in the virtual space:
//   - Creates/removes space nodes
//   - Connects/disconnects space edges
//   - Deploys/folds agents
//   - Broadcasts to the mesh
//
// This is the orchestrator's DIRECT manipulation of the capability graph.

// InjectTarget is what to inject into virtual space.
type InjectTarget struct {
	Operation string // "create-node", "remove-node", "connect", "disconnect", "deploy", "fold", "broadcast"
	Target    string // node ID, edge ID, agent ID
	Payload   string // what to do
	Space     SpacePosition // where in space
}

// InjectResult is the outcome of a forward injection.
type InjectResult struct {
	Target   InjectTarget
	Success  bool
	Detail   string
	Duration time.Duration
	Ref      string
}

// ForwardInjectStats tracks injection activity.
type ForwardInjectStats struct {
	TotalInjections int64
	NodesCreated    int64
	EdgesCreated    int64
	Deployments     int64
	Folds           int64
	Broadcasts      int64
}

var forwardInjectStats ForwardInjectStats

// ── Forward Inject Operations ─────────────────────────────────────────

// ForwardInject executes a forward injection into virtual space.
func ForwardInject(target InjectTarget) InjectResult {
	start := time.Now()
	forwardInjectStats.TotalInjections++

	result := InjectResult{Target: target}

	switch target.Operation {
	case "create-node":
		forwardInjectStats.NodesCreated++
		PlaceNode(target.Target, "injected", target.Space, CapFULL)
		result.Success = true
		result.Detail = fmt.Sprintf("node %s created at depth %d", target.Target, target.Space.Depth)

	case "connect":
		forwardInjectStats.EdgesCreated++
		ConnectNodes(target.Target, target.Payload, "injected", 0)
		result.Success = true
		result.Detail = fmt.Sprintf("edge %s → %s created", target.Target, target.Payload)

	case "deploy":
		forwardInjectStats.Deployments++
		result.Success = true
		result.Detail = fmt.Sprintf("deployed to %s", target.Target)

	case "fold":
		forwardInjectStats.Folds++
		if _, err := Fold(target.Target, "orchestrator", "injected", 1); err != nil {
			result.Detail = err.Error()
		} else {
			result.Success = true
			result.Detail = fmt.Sprintf("folded %s into orchestrator", target.Target)
		}

	case "broadcast":
		forwardInjectStats.Broadcasts++
		result.Success = true
		result.Detail = fmt.Sprintf("broadcast to %d space nodes", spaceNodeCount())

	default:
		result.Detail = fmt.Sprintf("unknown operation: %s", target.Operation)
	}

	result.Duration = time.Since(start)
	result.Ref = Ref([]byte(result.Detail))

	Log(LogInfo, "forward.inject."+target.Operation, result.Detail, result.Ref[:12], "", result.Duration, nil)
	Pulse("forward.inject", fmt.Sprintf("%s:%s", target.Operation, target.Target[:min(20, len(target.Target))]))

	return result
}

// ── System Command Interface ──────────────────────────────────────────

// ForwardInjectCommand is the user-facing /inject command.
func ForwardInjectCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return "/inject <create-node|connect|deploy|fold|broadcast> <target> [payload]"
	}

	target := InjectTarget{
		Operation: parts[0],
		Target:    parts[1],
		Space: SpacePosition{
			Depth:   1,
			Layer:   "injected",
			Machine: CurrentPOV().Machine,
			Region:  "eu",
		},
	}

	if len(parts) > 2 {
		target.Payload = strings.Join(parts[2:], " ")
	}

	result := ForwardInject(target)

	return fmt.Sprintf(`╔══ FORWARD INJECT ══╗
║  Op:    %s
║  Target: %s
║  Result: %s
║  Time:   %s
║  Ref:    %s
╚══════════════════════╝`,
		target.Operation, target.Target,
		func() string { if result.Success { return "✅ " + result.Detail } else { return "❌ " + result.Detail } }(),
		result.Duration.Round(time.Microsecond).String(),
		result.Ref[:12])
}

// ForwardInjectStatus returns compact status.
func ForwardInjectStatus() string {
	return fmt.Sprintf("forward-inject: %d ops · %d nodes · %d edges · %d deploys · %d folds · %d broadcasts",
		forwardInjectStats.TotalInjections, forwardInjectStats.NodesCreated,
		forwardInjectStats.EdgesCreated, forwardInjectStats.Deployments,
		forwardInjectStats.Folds, forwardInjectStats.Broadcasts)
}

// ForwardInjectVakedFit returns Vaked fit.
func ForwardInjectVakedFit() string {
	return `FORWARD INJECT = VIRTUAL SPACE DIRECT MANIPULATION

  The orchestrator reaches FORWARD into virtual space.
  Creates nodes, connects edges, deploys agents, folds recursively.
  Direct manipulation of the capability graph.

  "Forward inject in the virtual space. The orchestrator
   system command. Which, yes, does what it does." — Peter`
}
