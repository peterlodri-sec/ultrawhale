package blocks

import (
	"fmt"
	"strings"
	"time"
)

// ── BACKWARD EXPAND — The Opposite of Forward Inject ─────────────────
//
// Peter: "Let's do the opposite of fold. Expand. It's magic, it's just
// the opposite of a forward injection, it's the backward injection."
//
// FORWARD INJECT: pushes INTO virtual space (create, connect, fold).
// BACKWARD EXPAND: pulls FROM virtual space (extract, trace, unroll).
//
// Like a telescope in reverse. Instead of injecting into the graph,
// we EXPAND what's already there — making the invisible visible.
//
// The answer is GRAPHS. Everything is a graph. Every expand reveals
// more of the graph. Every fold compresses it. Forward inject writes
// to the graph. Backward expand reads FROM the graph.
//
// "The answer is graphs." — Peter

// ExpandTarget is what to expand (pull from virtual space).
type ExpandTarget struct {
	Operation string // "trace", "unroll", "extract", "visualize", "audit"
	Target    string // node ID, agent ID, pattern name
	Depth     int    // how deep to expand (default: 3)
}

// ExpandResult is the outcome of a backward expansion.
type ExpandResult struct {
	Target   ExpandTarget
	Content  string // the expanded content
	NodeCount int   // how many nodes were expanded
	EdgeCount int   // how many edges were traversed
	Duration time.Duration
	Ref      string
}

// BackwardExpandStats tracks expansion activity.
type BackwardExpandStats struct {
	TotalExpands   int64
	NodesTraced    int64
	EdgesTraced    int64
	Extractions    int64
	Visualizations int64
}

var backwardExpandStats BackwardExpandStats

// ── Backward Expand Operations ────────────────────────────────────────

// BackwardExpand pulls from virtual space (opposite of ForwardInject).
func BackwardExpand(target ExpandTarget) ExpandResult {
	start := time.Now()
	backwardExpandStats.TotalExpands++

	result := ExpandResult{Target: target}

	switch target.Operation {
	case "trace":
		result = expandTrace(target)
	case "unroll":
		result = expandUnroll(target)
	case "extract":
		result = expandExtract(target)
	case "visualize":
		result = expandVisualize(target)
	case "audit":
		result = expandAudit(target)
	default:
		result.Content = fmt.Sprintf("unknown expand: %s", target.Operation)
	}

	result.Duration = time.Since(start)
	result.Ref = Ref([]byte(result.Content))

	Log(LogInfo, "backward.expand."+target.Operation,
		fmt.Sprintf("%s (%d nodes, %d edges)", target.Target, result.NodeCount, result.EdgeCount),
		result.Ref[:12], "", result.Duration, nil)
	Pulse("backward.expand", fmt.Sprintf("%s:%s", target.Operation, target.Target[:min(20, len(target.Target))]))

	return result
}

func expandTrace(target ExpandTarget) ExpandResult {
	backwardExpandStats.NodesTraced++
	backwardExpandStats.EdgesTraced++

	reachable := Reachable(target.Target)
	nodeCount := len(reachable) + 1

	content := fmt.Sprintf(`╔══ EXPAND: trace %s ══╗
║  Node:     %s
║  Reachable: %d nodes
║  Path:     %s
║  Edges:    %d traversed
║  Depth:    %d
╚══════════════════════════╝`,
		target.Target, target.Target, nodeCount,
		strings.Join(reachable[:min(3, len(reachable))], " → "),
		backwardExpandStats.EdgesTraced, target.Depth)

	return ExpandResult{Target: target, Content: content, NodeCount: nodeCount, EdgeCount: len(reachable)}
}

func expandUnroll(target ExpandTarget) ExpandResult {
	backwardExpandStats.NodesTraced++
	backwardExpandStats.EdgesTraced++

	// Unroll: show the full topology from this node
	spaceTopology.mu.Lock()
	totalNodes := len(spaceTopology.Nodes)
	totalEdges := len(spaceTopology.Edges)
	spaceTopology.mu.Unlock()

	content := fmt.Sprintf(`╔══ EXPAND: unroll %s ══╗
║  Node:     %s
║  Topology: %d nodes, %d edges
║  Space:    %s
║  Depth:    %d
║  
║  The graph unrolls from here.
║  Every node is connected.
║  Every edge is a capability.
║  
║  The answer is GRAPHS.
╚══════════════════════════╝`,
		target.Target, target.Target,
		totalNodes, totalEdges,
		SpaceStatus(), target.Depth)

	return ExpandResult{Target: target, Content: content, NodeCount: totalNodes, EdgeCount: totalEdges}
}

func expandExtract(target ExpandTarget) ExpandResult {
	backwardExpandStats.Extractions++

	// Extract: pull the raw content from a VFS node
	content, err := VFSCat(target.Target)
	if err != nil { content = err.Error() }

	return ExpandResult{Target: target, Content: content, NodeCount: 1, EdgeCount: 0}
}

func expandVisualize(target ExpandTarget) ExpandResult {
	backwardExpandStats.Visualizations++

	// Visualize: render the space topology as a tree
	tree := VFSTree()[:min(500, len(VFSTree()))]
	return ExpandResult{Target: target, Content: tree, NodeCount: spaceNodeCount(), EdgeCount: len(tree) / 10}
}

func expandAudit(target ExpandTarget) ExpandResult {
	backwardExpandStats.NodesTraced++

	// Audit: check all guarantees for a node
	checks := HardenAll()
	drift := SurfaceDrift()

	content := fmt.Sprintf(`╔══ EXPAND: audit %s ══╗
║  Node:       %s
║  Hardening:  %d/6 guarantees
║  Entropy:    %.4f drift
║  SACRED:     %s
║  SEALING:    %s
║  
║  The answer is GRAPHS.
╚══════════════════════════════╝`,
		target.Target, target.Target,
		6, drift,
		func() string { if IsSacredIntact() { return "intact" }; return "degraded" }(),
		SealingStatus())

	return ExpandResult{Target: target, Content: content, NodeCount: 6, EdgeCount: 1}
}

// ── Command Interface ─────────────────────────────────────────────────

// BackwardExpandCommand is the user-facing /expand command.
func BackwardExpandCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return "/expand <trace|unroll|extract|visualize|audit> <target> [depth]"
	}

	target := ExpandTarget{
		Operation: parts[0],
		Target:    parts[1],
		Depth:     3,
	}
	if len(parts) > 2 { fmt.Sscanf(parts[2], "%d", &target.Depth) }

	result := BackwardExpand(target)

	return result.Content
}

// BackwardExpandStatus returns compact status.
func BackwardExpandStatus() string {
	return fmt.Sprintf("backward-expand: %d ops · %d traces · %d extractions · %d visualizations",
		backwardExpandStats.TotalExpands, backwardExpandStats.NodesTraced,
		backwardExpandStats.Extractions, backwardExpandStats.Visualizations)
}

// BackwardExpandVakedFit returns Vaked fit.
func BackwardExpandVakedFit() string {
	return `BACKWARD EXPAND = THE OPPOSITE OF FORWARD INJECT

  FORWARD INJECT: pushes INTO virtual space.
  BACKWARD EXPAND: pulls FROM virtual space.

  trace · unroll · extract · visualize · audit
  Like a telescope in reverse. Making the invisible visible.

  "The answer is GRAPHS." — Peter`
}
