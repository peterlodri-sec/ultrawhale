package blocks

import "fmt"

type TestifyEngine struct {
	Name    string
	Version string
	Stats   TestifyEngineStats
}

type TestifyEngineStats struct {
	ProbesRun   int64
	Predictions int64
	Lessons     int64
}

var testifyEngine = &TestifyEngine{Name: "testify-engine", Version: CurrentVersion()}

func TestifyEngineStatus() string {
	return fmt.Sprintf("testify: %s · %s · %s · %s",
		testifyEngine.Name, ProbeStatus(), PredictStatus(), LearnStatus())
}

func TestifyEngineVakedFit() string {
	return `Vaked:  ... → Enforce → TESTIFY-ENGINE → Index → ...
                                   ↑
                            probe + predict + learn
                            Evidence of what happened.`
}
