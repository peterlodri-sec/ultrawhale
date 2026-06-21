package blocks

import (
	"fmt"
	"strings"
)

// ── PULSING DNS — Swarm Topology as Live DNS Tree ──────────────────

func PulsingDNSTree() string {
	nodes := spaceNodeCount()
	agents := AgentCount()

	var sb strings.Builder
	sb.WriteString("╔══ PULSING DNS — Swarm Topology ══╗\n")
	sb.WriteString(fmt.Sprintf("║  🌐 vaked.dev           [LIVE]\n"))
	sb.WriteString(fmt.Sprintf("║  ├─ 🖥️  %s            [PULSING]\n", CurrentPOV().Machine))
	sb.WriteString(fmt.Sprintf("║  │   ├─ blocks: %d     [ACTIVE]\n", len(schemaRegistry)))
	sb.WriteString(fmt.Sprintf("║  │   ├─ agents: %d     [%s]\n", agents, func() string { if agents > 0 { return "PULSING" }; return "IDLE" }()))
	sb.WriteString(fmt.Sprintf("║  │   └─ nodes: %d      [%s]\n", nodes, func() string { if nodes > 0 { return "CONNECTED" }; return "SOLO" }()))
	if d := GetDyad(); d != nil && d.PeerAlive {
		sb.WriteString(fmt.Sprintf("║  └─ 🖥️  %s        [PULSING]\n", d.Peer.Machine))
	}
	sb.WriteString("╚══════════════════════════════════════╝")
	return sb.String()
}
