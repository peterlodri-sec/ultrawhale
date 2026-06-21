package blocks

import "fmt"

type SuperviseEngine struct {
	Name    string
	Version string
	Stats   SuperviseEngineStats
}

type SuperviseEngineStats struct {
	AgentsSpawned int64
	AgentsRestarted int64
	Delegations  int64
}

var superviseEngine = &SuperviseEngine{Name: "supervise-engine", Version: CurrentVersion()}

func SuperviseEngineStatus() string {
	o := GetOrchestrator()
	return fmt.Sprintf("supervise: %s · %d agents · %d turns · %s",
		superviseEngine.Name, AgentCount(), o.TotalTurns,
		GetSupervisor().SupervisorStatus())
}

func SuperviseEngineVakedFit() string {
	return `Vaked:  Declare-Engine → ENGINE → SUPERVISE-ENGINE → Enforce → Testify → Index → Reveal
                                         ↑
                                  orchestrator + ralph + supervisor
                                  Delegates and restarts.`
}
