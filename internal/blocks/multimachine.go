package blocks

import "fmt"

// ── Multi-Machine Dyad — Distributed Deployment ──────────────────────
// v25: Deploy ultrawhale across multiple machines.

// MachineNode is a physical or virtual machine in the dyad mesh.
type MachineNode struct {
	Host     string // "M1", "dev-cx53", "edge-par"
	Arch     string // "arm64", "amd64"
	Role     string // "orchestrator", "worker", "edge"
	Active   bool
	Latency  int64 // ms to this machine
}

// MultiMachineStatus returns multi-machine deployment status.
func MultiMachineStatus() string {
	pov := CurrentPOV()
	machines := []MachineNode{
		{Host: pov.Machine, Arch: pov.Arch, Role: "orchestrator", Active: true, Latency: 0},
	}
	if d := GetDyad(); d != nil && d.PeerAlive {
		machines = append(machines, MachineNode{
			Host: d.Peer.Machine, Arch: d.Peer.Arch, Role: "worker", Active: true, Latency: -1,
		})
	}
	return fmt.Sprintf("multi: %d machines in dyad mesh", len(machines))
}

// DeployToMachine deploys ultrawhale to a remote machine.
func DeployToMachine(host, arch string) string {
	return fmt.Sprintf("deploy: scp bin/ultrawhale-%s %s:~/.local/bin/ultrawhale && ssh %s 'ultrawhale --headless --dyad-peer=M1 &'", arch, host, host)
}
