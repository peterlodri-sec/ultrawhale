package modes

import (
	"testing"
)

func TestUltracodePhases(t *testing.T) {
	u := NewUltracode("test-session")
	if len(u.phases) != 7 {
		t.Fatalf("expected 7 phases, got %d", len(u.phases))
	}
	
	expected := []string{"plan", "implement", "test", "review", "fix", "verify", "commit"}
	for i, name := range expected {
		if u.phases[i].Name != name {
			t.Fatalf("phase %d: expected %s, got %s", i, name, u.phases[i].Name)
		}
	}
	t.Log("7 phases OK")
}

func TestPhaseLifecycle(t *testing.T) {
	u := NewUltracode("test-session")
	
	// Start plan
	idx, err := u.StartPhase("plan")
	if err != nil { t.Fatal(err) }
	if idx != 0 { t.Fatalf("expected idx 0, got %d", idx) }
	if u.phases[0].Status != PhaseRunning { t.Fatal("plan not running") }
	
	// End plan
	u.EndPhase("plan", true, nil)
	if u.phases[0].Status != PhasePassed { t.Fatal("plan not passed") }
	
	// Auto-advance to implement
	name, ok := u.AutoAdvance()
	if !ok { t.Fatal("auto-advance failed") }
	if name != "implement" { t.Fatalf("expected implement, got %s", name) }
	
	t.Log("Phase lifecycle OK: plan→passed, advance→implement")
}

func TestFailureRollback(t *testing.T) {
	u := NewUltracode("test-session")
	u.StartPhase("implement")
	u.EndPhase("implement", false, nil)
	
	if u.phases[1].Status != PhaseFailed { t.Fatal("implement not failed") }
	
	// Fix should still be pending
	if u.phases[4].Status != PhasePending { t.Fatal("fix should be pending") }
	
	t.Log("Failure rollback OK: implement failed, fix pending")
}

func TestSkipPhase(t *testing.T) {
	u := NewUltracode("test-session")
	u.SkipPhase("fix")
	if u.phases[4].Status != PhaseSkipped { t.Fatal("fix not skipped") }
	t.Log("Skip OK")
}

func TestPhaseSummary(t *testing.T) {
	u := NewUltracode("test-session")
	u.StartPhase("plan")
	summary := u.PhaseSummary()
	if summary == "" { t.Fatal("empty summary") }
	t.Logf("Summary: %s", summary)
}

func TestPOVIntegration(t *testing.T) {
	u := NewUltracode("test-session")
	ctx := u.Context()
	if ctx == "" { t.Fatal("empty context") }
	t.Logf("Context: %s", ctx)
}
