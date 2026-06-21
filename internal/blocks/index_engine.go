package blocks

import "fmt"

type IndexEngine struct {
	Name    string
	Version string
	Stats   IndexEngineStats
}

type IndexEngineStats struct {
	SpaceNodes   int64
	VFSOperations int64
	CrabCCQueries int64
}

var indexEngine = &IndexEngine{Name: "index-engine", Version: CurrentVersion()}

func IndexEngineStatus() string {
	return fmt.Sprintf("index: %s · %s · %s · %s",
		indexEngine.Name, SpaceStatus(), VFSStatus(), CrabCCStatus())
}

func IndexEngineVakedFit() string {
	return `Vaked:  ... → Testify → INDEX-ENGINE → Reveal
                                   ↑
                            space + vfs + crabcc
                            Where everything is.`
}
