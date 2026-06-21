package blocks

import "fmt"

// ── Global Mesh — Multi-Orchestrator Coordination ────────────────────
// v50 timeline: 10+ machines coordinated as one logical orchestrator.

type MeshNode struct {
	ID       string
	Machine  string
	Role     string // "orchestrator", "worker", "edge"
	Status   string // "active", "idle", "offline"
	Load     int    // current task count
	Latency  int64  // ms
}

var meshNodes = make([]MeshNode, 0, 16)

// JoinMesh adds a machine to the global mesh.
func JoinMesh(machine, role string) {
	meshNodes = append(meshNodes, MeshNode{
		ID:      fmt.Sprintf("mesh-%s", machine),
		Machine: machine,
		Role:    role,
		Status:  "active",
	})
	Log(LogInfo, "mesh.join", machine, "", "", 0, nil)
}

// MeshBalance distributes load across the mesh.
func MeshBalance(taskCount int) string {
	active := 0
	for _, n := range meshNodes {
		if n.Status == "active" { active++ }
	}
	if active == 0 { return fmt.Sprintf("mesh: single-node (no peers) — %d tasks", taskCount) }
	
	perNode := taskCount / active
	return fmt.Sprintf("mesh: %d nodes · %d tasks · %d per node", active, taskCount, perNode)
}

func GlobalMeshStatus() string {
	return MeshBalance(AgentCount())
}

func GlobalMeshVakedFit() string {
	return `GLOBAL MESH = SUPERVISES LAYER SCALED

  v40: single orchestrator
  v45: multi-orchestrator mesh (10+ machines)
  v50: global mesh with auto-balancing

  Space topology across machines.
  Dyad becomes mesh. Mesh becomes global.`
}
