package blocks

import (
	"fmt"
	"os/exec"
)

// ── CrabCC Primitive — Symbol-Level Code Indexing ─────────────────────
// Vaked layer: Indexes. Integrates with CrabCC for symbol-level indexing.

// CrabCCIndex represents a CrabCC code index.
type CrabCCIndex struct {
	Path      string // path to the index
	Symbols   int    // total symbols indexed
	Files     int    // total files indexed
	BuildTime string // last build time
}

// CrabCCBuild builds a CrabCC index for the workspace.
func CrabCCBuild(workspace string) (*CrabCCIndex, error) {
	if _, err := exec.LookPath("crabcc"); err != nil {
		return &CrabCCIndex{Path: "crabcc not installed"}, nil
	}

	cmd := exec.Command("crabcc", "index", "build", workspace)
	if out, err := cmd.CombinedOutput(); err != nil {
		Log(LogWarn, "crabcc.build", workspace, "", "", 0, err)
		return nil, fmt.Errorf("crabcc: %s", string(out))
	}

	idx := &CrabCCIndex{Path: ".crabcc/index.db"}
	Log(LogInfo, "crabcc.build", workspace, "", "", 0, nil)
	return idx, nil
}

// CrabCCLookup looks up a symbol in the CrabCC index.
func CrabCCLookup(symbol string) (string, error) {
	if _, err := exec.LookPath("crabcc"); err != nil {
		return "", fmt.Errorf("crabcc not installed")
	}

	cmd := exec.Command("crabcc", "lookup", "sym", symbol)
	out, err := cmd.CombinedOutput()
	if err != nil { return "", err }
	return string(out), nil
}

// CrabCCStatus returns compact CrabCC status.
func CrabCCStatus() string {
	if _, err := exec.LookPath("crabcc"); err != nil {
		return "crabcc: not installed (https://crabcc.com)"
	}
	return "crabcc: available"
}
