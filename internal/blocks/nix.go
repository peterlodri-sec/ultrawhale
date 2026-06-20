package blocks

import (
	"fmt"
	"os/exec"
	"strings"
)

// ── Nix Primitive — Materialization from Blocks ──────────────────────
// Vaked layer: Materializes. Generates flake.nix and shell.nix from blocks.

func detectToolchains() []string {
	var pkgs []string
	pov := CurrentPOV()
	if _, err := exec.LookPath("go"); err == nil { pkgs = append(pkgs, "go") }
	if _, err := exec.LookPath("zig"); err == nil { pkgs = append(pkgs, "zig") }
	if _, err := exec.LookPath("python3"); err == nil { pkgs = append(pkgs, "python3") }
	if _, err := exec.LookPath("docker"); err == nil { pkgs = append(pkgs, "docker") }
	_ = pov
	return pkgs
}

func NixStatus() string {
	pkgs := detectToolchains()
	return fmt.Sprintf("nix: %d toolchains (%s) [%s]", len(pkgs), strings.Join(pkgs, ", "), CurrentPOV().Machine)
}
