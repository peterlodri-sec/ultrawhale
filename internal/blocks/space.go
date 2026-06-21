package blocks

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ── Space Primitive — The Missing Fundamental Atom ────────────────────
//
// Space is WHERE things are in relation to each other.
// Time tells WHEN. Context tells WHAT. Space tells WHERE.
//
// In Vaked: the capability graph IS space. Nodes are declarations.
// Edges are capabilities. Distance is latency. Adjacency is reachability.
//
// In ultrawhale:
//   - Agent mesh topology (who can reach whom)
//   - Orchestration DAG (orchestrator → swarm → subagent)
//   - Capability adjacency (FULL ↔ OBSERVE)
//   - Surface layout (InfraBar above Chat above HUD)
//   - Machine topology (M1 ↔ dev-cx53 via Tailscale)
//   - Block nesting (block → journal → log)
//   - Dyad pairing (two machines, one shared context)

// SpaceNode is a point in the capability graph.
type SpaceNode struct {
	ID       string            // unique node identifier
	Kind     string            // "agent", "block", "widget", "machine", "surface"
	Position SpacePosition     // where in the graph
	Edges    []SpaceEdge       // connections to other nodes
	Capabilities CapProfile    // what this node can do
	State    string            // "active", "idle", "degraded", "offline"
	LastSeen time.Time
}

// SpacePosition places a node in n-dimensional space.
// Dimensions: graph depth, orchestration layer, machine locality, UI layer.
type SpacePosition struct {
	Depth     int    // graph depth (0 = root/orchestrator)
	Layer     string // "blocks", "plugins", "orchestrator", "tui", "surface"
	Machine   string // "M1", "dev-cx53", "edge"
	Region    string // "eu", "us", "apac"
}

// SpaceEdge is a connection between two nodes.
type SpaceEdge struct {
	From      string        // source node ID
	To        string        // target node ID
	Kind      string        // "delegates", "streams", "pings", "contains", "adjacent"
	Latency   time.Duration // measured round-trip time
	Bandwidth int64         // bytes/sec (0 if unknown)
	Active    bool
}

// ── Space Topology ────────────────────────────────────────────────────

// SpaceTopology is the complete capability graph — the Vaked "space".
type SpaceTopology struct {
	mu    sync.Mutex
	Nodes map[string]*SpaceNode
	Edges []SpaceEdge
}

var spaceTopology = &SpaceTopology{Nodes: make(map[string]*SpaceNode)}

// PlaceNode adds a node to the topology at a position.
func PlaceNode(id, kind string, pos SpacePosition, caps CapProfile) *SpaceNode {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()

	node := &SpaceNode{
		ID:           id,
		Kind:         kind,
		Position:     pos,
		Capabilities: caps,
		State:        "active",
		LastSeen:     time.Now(),
	}
	spaceTopology.Nodes[id] = node

	Log(LogInfo, "space.place", fmt.Sprintf("%s (%s) at depth=%d layer=%s machine=%s",
		id, kind, pos.Depth, pos.Layer, pos.Machine),
		"", "", 0, nil)
	return node
}

// ConnectNodes creates a directed edge between two nodes.
func ConnectNodes(from, to, kind string, latency time.Duration) {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()

	edge := SpaceEdge{From: from, To: to, Kind: kind, Latency: latency, Active: true}
	spaceTopology.Edges = append(spaceTopology.Edges, edge)
}

// ── Space Queries ─────────────────────────────────────────────────────

// Distance returns the graph distance between two nodes (shortest path hops).
func Distance(from, to string) int {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()

	// BFS shortest path
	visited := make(map[string]bool)
	queue := []struct {
		id   string
		dist int
	}{{from, 0}}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.id == to { return current.dist }

		for _, edge := range spaceTopology.Edges {
			if edge.From == current.id && edge.Active && !visited[edge.To] {
				visited[edge.To] = true
				queue = append(queue, struct {
					id   string
					dist int
				}{edge.To, current.dist + 1})
			}
		}
	}
	return -1 // unreachable
}

// Reachable returns all nodes reachable from a given node.
func Reachable(from string) []string {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()

	visited := make(map[string]bool)
	queue := []string{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range spaceTopology.Edges {
			if edge.From == current && edge.Active && !visited[edge.To] {
				visited[edge.To] = true
				queue = append(queue, edge.To)
			}
		}
	}

	var result []string
	for id := range visited {
		if id != from { result = append(result, id) }
	}
	return result
}

// TopologyStatus returns compact topology status.
func TopologyStatus() string {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()

	return fmt.Sprintf("space: %d nodes, %d edges", len(spaceTopology.Nodes), len(spaceTopology.Edges))
}

// ── Built-in Topology ────────────────────────────────────────────────

func init() {
	pov := CurrentPOV()

	// Place the orchestrator at the root (depth 0)
	PlaceNode("orchestrator", "orchestrator",
		SpacePosition{Depth: 0, Layer: "orchestrator", Machine: pov.Machine, Region: "eu"},
		CapFULL)

	// Place blocks engine
	PlaceNode("blocks", "engine",
		SpacePosition{Depth: 1, Layer: "blocks", Machine: pov.Machine, Region: "eu"},
		CapFULL)
	ConnectNodes("orchestrator", "blocks", "contains", 0)

	// Place TUI surface
	PlaceNode("tui", "surface",
		SpacePosition{Depth: 1, Layer: "tui", Machine: pov.Machine, Region: "eu"},
		CapOBSERVE)
	ConnectNodes("orchestrator", "tui", "renders", 0)

	// Place dyad peer if available
	if d := GetDyad(); d != nil {
		PlaceNode(d.Peer.Machine, "machine",
			SpacePosition{Depth: 0, Layer: "machine", Machine: d.Peer.Machine, Region: "eu"},
			CapFULL)
		ConnectNodes(pov.Machine, d.Peer.Machine, "dyad", 0)
	}
}

// SpaceStatus returns a compact space status for HUD/display.
func SpaceStatus() string {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()

	nodesByLayer := make(map[string]int)
	nodesByMachine := make(map[string]int)
	for _, n := range spaceTopology.Nodes {
		nodesByLayer[n.Position.Layer]++
		nodesByMachine[n.Position.Machine]++
	}

	return fmt.Sprintf("space: %d nodes · %d edges · %d layers · %d machines",
		len(spaceTopology.Nodes), len(spaceTopology.Edges),
		len(nodesByLayer), len(nodesByMachine))
}

// VakedTriangle returns the three pillars of the system.
func VakedTriangle() string {
	return strings.Join([]string{
		"Context (WHAT: POV, capabilities, brain)",
		"Time   (WHEN: journal, sessions, Ralph versions)",
		"Space  (WHERE: topology, distance, reachability)",
	}, "\n")
}
