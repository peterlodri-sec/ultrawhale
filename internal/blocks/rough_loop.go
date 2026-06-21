package blocks

import (
	"fmt"
	"time"
)

// ── ROUGH LOOP — Live Liveness Dashboard ────────────────────────────
// The rough loop produces random live data every 30s.
// This data proves deep recursive liveness.

type RoughLoop struct {
	Ticks       int64
	Entropy     float64
	LastTick    time.Time
	Proofs      []string
}

var roughLoop = &RoughLoop{Proofs: make([]string, 0, 32)}

func RoughLoopTick() {
	roughLoop.Ticks++
	roughLoop.Entropy = SurfaceDrift()
	roughLoop.LastTick = time.Now()

	proof := Ref([]byte(fmt.Sprintf("rough:%d:%.4f:%s", roughLoop.Ticks, roughLoop.Entropy, time.Now().String())))
	roughLoop.Proofs = append(roughLoop.Proofs, proof[:8])
	if len(roughLoop.Proofs) > 32 { roughLoop.Proofs = roughLoop.Proofs[1:] }

	Pulse("rough-loop.tick", fmt.Sprintf("#%d %.4f", roughLoop.Ticks, roughLoop.Entropy))
}

func RoughLoopDashboard() string {
	return fmt.Sprintf(`╔══ ROUGH LOOP — Liveness Dashboard ══╗
║  Ticks:    %d
║  Entropy:  %.4f (░ ▓ █)
║  Last:     %s
║  Proofs:   %v
║  Status:   %s
╚══════════════════════════════════════╝`,
		roughLoop.Ticks, roughLoop.Entropy,
		roughLoop.LastTick.Format("15:04:05"),
		roughLoop.Proofs,
		func() string {
			if roughLoop.Entropy > 0 { return "🟢 LIVE — producing random data" }
			return "🟡 IDLE — waiting for noise"
		}())
}
