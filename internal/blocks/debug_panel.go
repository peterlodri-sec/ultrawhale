package blocks

import (
	"fmt"
	"sync/atomic"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── Debug Panel — FREEZE VIEW + PROBLEM REPORT ──────────────────────
//
// When a BIG_PROBLEM is detected, the debug panel activates:
//   1. FREEZE VIEW — capture current state snapshot
//   2. SHOW PROBLEM — display problem details
//   3. HOW TO REPORT — generate issue template
//   4. DEBUG LOG — store everything to ~/.ultrawhale/debug/

// DebugSnapshot captures the full system state.
type DebugSnapshot struct {
	Timestamp   string
	Version     string
	POV         string
	Problem     string
	Agents      int
	Memory      string
	Guarantees  []string
	LogTail     []string
}

// FreezeView captures a point-in-time snapshot.
func FreezeView(problem string) DebugSnapshot {
	pov := CurrentPOV()
	hardenReport := HardenAll()

	snapshot := DebugSnapshot{
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   CurrentVersion(),
		POV:       fmt.Sprintf("%s/%s/%s", pov.Machine, pov.Arch, pov.Tier),
		Problem:   problem,
		Agents:    AgentCount(),
		Memory:    fmt.Sprintf("%d blocks, %d plugins", len(schemaRegistry), 6),
		Guarantees: strings.Split(hardenReport, "\n"),
		LogTail:   getLogTail(20),
	}

	// Store to disk
	storeDebugSnapshot(snapshot)

	Log(LogWarn, "debug.freeze", problem, "", "", 0, nil)
	return snapshot
}

// ShowProblem displays the problem debug panel.
func ShowProblem(problem string) string {
	snapshot := FreezeView(problem)

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════╗\n")
	sb.WriteString("║  🔍 PROBLEM DEBUG PANEL                           ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║  Problem: %s\n", snapshot.Problem[:min(40, len(snapshot.Problem))]))
	sb.WriteString(fmt.Sprintf("║  Time:    %s\n", snapshot.Timestamp))
	sb.WriteString(fmt.Sprintf("║  Version: %s\n", snapshot.Version))
	sb.WriteString(fmt.Sprintf("║  POV:     %s\n", snapshot.POV))
	sb.WriteString(fmt.Sprintf("║  Agents:  %d\n", snapshot.Agents))
	sb.WriteString(fmt.Sprintf("║  Memory:  %s\n", snapshot.Memory))
	sb.WriteString("╠══════════════════════════════════════════════════╣\n")
	sb.WriteString("║  HOW TO REPORT:                                   ║\n")
	sb.WriteString("║  github.com/peterlodri-sec/ultrawhale/issues/new   ║\n")
	sb.WriteString("║  Include the debug log from:                       ║\n")
	sb.WriteString("║  ~/.ultrawhale/debug/                              ║\n")
	sb.WriteString("╠══════════════════════════════════════════════════╣\n")
	sb.WriteString("║  RECENT LOGS:                                      ║\n")
	for _, log := range snapshot.LogTail {
		sb.WriteString(fmt.Sprintf("║  %s\n", log[:min(46, len(log))]))
	}
	sb.WriteString("╚══════════════════════════════════════════════════╝")

	return sb.String()
}

func getLogTail(n int) []string {
	count := int(atomic.LoadInt64(&globalLogger.count))
	head := int(atomic.LoadInt64(&globalLogger.head))
	start := (head - n + len(globalLogger.buffer)) % len(globalLogger.buffer)
	
	var logs []string
	for i := 0; i < n && i < count; i++ {
		idx := (start + i) % len(globalLogger.buffer)
		logs = append(logs, globalLogger.buffer[idx].Operation[:min(48, len(globalLogger.buffer[idx].Operation))])
	}
	return logs
}

func storeDebugSnapshot(s DebugSnapshot) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".ultrawhale", "debug")
	os.MkdirAll(dir, 0o700)

	path := filepath.Join(dir, fmt.Sprintf("debug-%s.log", time.Now().Format("20060102-150405")))
	f, _ := os.Create(path)
	defer f.Close()

	fmt.Fprintf(f, "Problem: %s\n", s.Problem)
	fmt.Fprintf(f, "Time: %s\n", s.Timestamp)
	fmt.Fprintf(f, "Version: %s\n", s.Version)
	fmt.Fprintf(f, "POV: %s\n", s.POV)
	fmt.Fprintf(f, "Agents: %d\n", s.Agents)
	fmt.Fprintf(f, "Memory: %s\n", s.Memory)
	fmt.Fprintf(f, "Guarantees:\n")
	for _, g := range s.Guarantees {
		fmt.Fprintf(f, "  %s\n", g)
	}
	fmt.Fprintf(f, "Recent Logs:\n")
	for _, l := range s.LogTail {
		fmt.Fprintf(f, "  %s\n", l)
	}
}

// DebugStatus returns compact debug panel status.
func DebugPanelStatus() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".ultrawhale", "debug")
	entries, _ := os.ReadDir(dir)
	return fmt.Sprintf("debug: %d snapshots stored in %s", len(entries), dir)
}

// DebugPanelVakedFit returns debug panel Vaked fit.
func DebugPanelVakedFit() string {
	return `DEBUG PANEL = FREEZE + SHOW + REPORT + STORE

  /problem detect → FreezeView → ShowProblem → Store to disk
  Everything captured. Nothing lost.
  Report with the debug log. The Genesis block verifies.`
}
