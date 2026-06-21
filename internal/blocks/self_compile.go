package blocks

import (
	"fmt"
	"os"
	"os/exec"
)

// ── Self-Compile — ultrawhale Compiles Itself ────────────────────────
// v50 timeline: .vaked declarations → native Go → binary.

type SelfCompile struct {
	SourceDir string
	OutputDir string
	Targets   []CompileTarget
}

type CompileTarget struct {
	OS   string // "darwin", "linux"
	Arch string // "arm64", "amd64"
	Path string // output binary path
}

// SelfCompileBuild compiles ultrawhale from its own source.
func SelfCompileBuild() (*SelfCompile, error) {
	sc := &SelfCompile{
		SourceDir: ".",
		OutputDir: "bin",
		Targets: []CompileTarget{
			{OS: "darwin", Arch: "arm64", Path: "bin/ultrawhale-darwin-arm64"},
			{OS: "linux", Arch: "amd64", Path: "bin/ultrawhale-linux-amd64"},
		},
	}

	for _, t := range sc.Targets {
		cmd := exec.Command("go", "build",
			"-trimpath", "-ldflags=-s -w",
			"-o", t.Path, "./cmd/whale")
		cmd.Env = append(os.Environ(),
			"GOOS="+t.OS, "GOARCH="+t.Arch, "CGO_ENABLED=0")
		
		if out, err := cmd.CombinedOutput(); err != nil {
			return sc, fmt.Errorf("self-compile %s/%s: %s", t.OS, t.Arch, string(out))
		}
		Log(LogInfo, "self-compile."+t.OS, t.Path, "", "", 0, nil)
	}

	return sc, nil
}

// SelfCompileVakedFit returns Vaked fit for self-compiling.
func SelfCompileVakedFit() string {
	return `SELF-COMPILE = MATERIALIZES LAYER RECURSIVE

  v40: go build
  v45: .vaked → vakedz → Go → binary
  v50: declare → compile → deploy — one shot

  ultrawhale compiles itself.
  The builder becomes the built.
  The loop closes. The loop continues.`
}
