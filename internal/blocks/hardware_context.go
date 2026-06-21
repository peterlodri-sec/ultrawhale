package blocks

import (
	"fmt"
	"runtime"
)

// ── HARDWARE CONTEXT — Full Metal Declaration ──────────────────────

type HardwareContext struct {
	Machine  string
	CPU      int
	Arch     string
	GoVersion string
	OS       string
}

func HardwareContextReport() string {
	h := HardwareContext{
		Machine:   CurrentPOV().Machine,
		CPU:       runtime.NumCPU(),
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
	}

	return fmt.Sprintf(`╔══ HARDWARE CONTEXT ══╗
║  Machine: %s
║  CPU:     %d cores
║  Arch:    %s
║  Go:      %s
║  OS:      %s
║  SEALING: 10%% reserved (%d CPU)
╚══════════════════════╝`,
		h.Machine, h.CPU, h.Arch, h.GoVersion, h.OS, max(1, h.CPU/10))
}
