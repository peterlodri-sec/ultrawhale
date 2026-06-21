package blocks

import "fmt"

type EnforceEngine struct {
	Name    string
	Version string
	Stats   EnforceEngineStats
}

type EnforceEngineStats struct {
	PrehooksRun    int64
	BlocksRejected int64
	PermissionsDenied int64
}

var enforceEngine = &EnforceEngine{Name: "enforce-engine", Version: CurrentVersion()}

func EnforceEngineStatus() string {
	return fmt.Sprintf("enforce: %s · %s · %s",
		enforceEngine.Name, PermissionStatus(), SacredStatus())
}

func EnforceEngineVakedFit() string {
	return `Vaked:  ... → Supervise → ENFORCE-ENGINE → Testify → ...
                                    ↑
                             pre-hooks + permission + sacred
                             Gatekeeper of every operation.`
}
