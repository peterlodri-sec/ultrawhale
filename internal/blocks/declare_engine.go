package blocks

import "fmt"

type DeclareEngine struct {
	Name    string
	Version string
	Stats   DeclareEngineStats
}

type DeclareEngineStats struct {
	Parsed    int64
	Validated int64
	Rejected  int64
}

var declareEngine = &DeclareEngine{Name: "declare-engine", Version: CurrentVersion()}

func DeclareEngineStatus() string {
	return fmt.Sprintf("declare: %s · %d parsed, %d valid, %d rejected · %s",
		declareEngine.Name, declareEngine.Stats.Parsed,
		declareEngine.Stats.Validated, declareEngine.Stats.Rejected,
		SchemaStatus())
}

func DeclareEngineVakedFit() string {
	return `Vaked:  DECLARE-ENGINE → Engine → Materializes → UI-Engine → Reveals
                   ↑
            schema + contract + capabilities
            Validates what is declared.`
}
