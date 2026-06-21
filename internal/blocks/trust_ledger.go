package blocks

import (
	"fmt"
	"time"
)

// ── GENESIS TRUST LEDGER — Trust Given at Genesis ──────────────────

type TrustLedger struct {
	GenesisTrust  float64
	CurrentTrust  float64
	Operations    int64
	LastWithdrawal time.Time
	SealingActive bool
}

var trustLedger = &TrustLedger{
	GenesisTrust: 100.0,
	CurrentTrust: 100.0,
}

func TrustWithdraw(amount float64) {
	trustLedger.Operations++
	trustLedger.CurrentTrust -= amount
	trustLedger.LastWithdrawal = time.Now()

	if trustLedger.CurrentTrust <= 10.0 {
		trustLedger.SealingActive = true
	}
}

func TrustStatus() string {
	return fmt.Sprintf(`╔══ GENESIS TRUST LEDGER ══╗
║  Genesis: %.0f%%
║  Current: %.1f%%
║  Ops:     %d
║  SEALING: %s
║  Reserve: 10%% (%.1f%%)
╚══════════════════════════╝`,
		trustLedger.GenesisTrust,
		trustLedger.CurrentTrust,
		trustLedger.Operations,
		func() string { if trustLedger.SealingActive { return "🛡️ ACTIVE" }; return "✅ healthy" }(),
		trustLedger.GenesisTrust*0.1)
}
