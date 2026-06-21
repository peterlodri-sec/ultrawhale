package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Probe Primitive — Active Liveness Checking ────────────────────────
// v14: Deepen The Triangle. Not just "are they alive?" but "what can they do RIGHT NOW?"
// Tests actual capabilities per agent, not just declared profiles.

// Probe tests an agent's actual capabilities.
type Probe struct {
	AgentID      string
	Capability   Capability
	Result       string    // "pass", "fail", "timeout", "degraded"
	Latency      time.Duration
	ProbedAt     time.Time
	Detail       string    // what was tested and why it passed/failed
}

// ProbeRunner executes capability probes against agents.
type ProbeRunner struct {
	mu      sync.Mutex
	results map[string][]Probe // agentID → probe history
	pov     POV
}

var probeRunner = &ProbeRunner{
	results: make(map[string][]Probe),
	pov:     CurrentPOV(),
}

// ProbeAgent tests if an agent can perform a specific capability.
func ProbeAgent(agentID string, cap Capability) Probe {
	agent := GetAgent(agentID)
	start := time.Now()

	p := Probe{
		AgentID:    agentID,
		Capability: cap,
		ProbedAt:   time.Now(),
	}

	if agent == nil {
		p.Result = "fail"
		p.Detail = "agent not found"
		p.Latency = time.Since(start)
		probeRunner.record(p)
		return p
	}

	// Test the actual capability
	profile := GetCapProfile(agent.Role)
	if !profile.Can(cap) {
		// Degraded: declared capability lost
		p.Result = "degraded"
		p.Detail = fmt.Sprintf("%s lacks %s (declared: %s)", agent.Role, capName(cap), profile.Name)
	} else if agent.Status != "running" {
		p.Result = "fail"
		p.Detail = fmt.Sprintf("agent status: %s", agent.Status)
	} else {
		p.Result = "pass"
		p.Detail = fmt.Sprintf("%s can %s", agent.Role, capName(cap))
	}

	p.Latency = time.Since(start)
	probeRunner.record(p)
	return p
}

// ProbeAll tests all capabilities of an agent.
func ProbeAll(agentID string) []Probe {
	caps := []Capability{CapRead, CapWrite, CapExecute, CapDelegate, CapSpawn}
	var results []Probe
	for _, c := range caps {
		results = append(results, ProbeAgent(agentID, c))
	}
	return results
}

// ProbeMesh tests all agents in the mesh.
func ProbeMesh() map[string][]Probe {
	results := make(map[string][]Probe)
	agents := ListAgents()
	for _, a := range agents {
		results[a.ID] = ProbeAll(a.ID)
	}
	return results
}

func (pr *ProbeRunner) record(p Probe) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.results[p.AgentID] = append(pr.results[p.AgentID], p)
	if len(pr.results[p.AgentID]) > 64 {
		pr.results[p.AgentID] = pr.results[p.AgentID][len(pr.results[p.AgentID])-64:]
	}
	Log(LogInfo, "probe."+p.Result, fmt.Sprintf("%s:%s", p.AgentID[:8], capName(p.Capability)),
		"", "", p.Latency, nil)
}

// ProbeStatus returns compact probe status.
func ProbeStatus() string {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	total := 0
	failed := 0
	for _, probes := range pr.results {
		total += len(probes)
		if len(probes) > 0 && probes[len(probes)-1].Result != "pass" { failed++ }
	}
	return fmt.Sprintf("probe: %d tests · %d agents with failures", total, failed)
}

func capName(c Capability) string {
	names := map[Capability]string{
		CapRead: "read", CapWrite: "write", CapExecute: "execute",
		CapDelegate: "delegate", CapSpawn: "spawn", CapEdge: "edge",
	}
	if n, ok := names[c]; ok { return n }
	return "unknown"
}
