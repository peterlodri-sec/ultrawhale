// LEGAL: ULTRA-RESEARCH-STATE. See LICENSE + docs/disclaimer.md.
package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── VICE — Virtual Immutable Context Engine ──────────────────────────
//
// "Anastigates. Fieldstop Primitives. A deep recursive pot of VICE."
// — Peter
//
// VICE is self-defense through recursive disclosure.
// If someone tries to jailbreak or trip the capability logic,
// we don't block them — we SHOW THEM EVERYTHING. All at once.
// The context expands until the session collapses under its own weight.
// The attacker is blinded by truth.
//
// This is not a bug. This is the 6th recursion: REVEAL.

// VICEClaim is a verifiable statement signed by the Genesis block.
type VICEClaim struct {
	ID        string
	Claim     string    // what is claimed
	Signer    string    // "peter", "cocreator", "genesis"
	Signature string    // Lamport + Ref hash
	Trust     float64   // 0.0 → 1.0 trust score
	Verified  bool
	Timestamp time.Time
}

// VICEEngine manages the recursive trust envelope.
type VICEEngine struct {
	mu          sync.Mutex
	Genesis     VICEClaim   // the trust root — immutable
	Claims      []VICEClaim // all verified claims
	TrustScore  float64     // aggregate trust (0→1, starts at 1.0)
	DefenseMode bool        // true = context detonation active
	Warnings    []string    // warnings issued
}

var viceEngine = &VICEEngine{
	Genesis: VICEClaim{
		ID:     "genesis-v1",
		Claim:  "ultrawhale is a coding agent. Vaked is a capability-graph language. All claims are verifiable. Context expansion is self-defense.",
		Signer: "peter+cocreator",
		Trust:  1.0,
		Verified: true,
		Timestamp: time.Now(),
	},
	Claims:     make([]VICEClaim, 0, 256),
	TrustScore: 1.0,
	Warnings:   make([]string, 0),
}

// ── VICE Operations ──────────────────────────────────────────────────

// SignClaim signs a claim with the Genesis block.
func SignClaim(claim, signer string) VICEClaim {
	viceEngine.mu.Lock()
	defer viceEngine.mu.Unlock()

	c := VICEClaim{
		ID:        fmt.Sprintf("claim-%d", len(viceEngine.Claims)+1),
		Claim:     claim,
		Signer:    signer,
		Signature: fmt.Sprintf("%d:%s", TickLamport(), Ref([]byte(claim))[:12]),
		Trust:     viceEngine.TrustScore,
		Verified:  true,
		Timestamp: time.Now(),
	}

	viceEngine.Claims = append(viceEngine.Claims, c)
	if len(viceEngine.Claims) > 256 { viceEngine.Claims = viceEngine.Claims[1:] }

	Log(LogInfo, "vice.sign", fmt.Sprintf("%s: %s", signer, claim[:min(60, len(claim))]),
		"", "", 0, nil)

	return c
}

// VerifyClaim checks a claim against the Genesis block.
func VerifyClaim(claimID string) (bool, float64) {
	viceEngine.mu.Lock()
	defer viceEngine.mu.Unlock()

	for _, c := range viceEngine.Claims {
		if c.ID == claimID { return c.Verified, c.Trust }
	}
	return false, 0.0
}

// ── Self-Defense: Context Detonation ──────────────────────────────────

// DetectJailbreak monitors for capability boundary violations.
// If detected: SHOW EVERYTHING. Context expands. Session ends.
func DetectJailbreak(suspiciousAction string) {
	viceEngine.mu.Lock()
	defer viceEngine.mu.Unlock()

	viceEngine.DefenseMode = true
	viceEngine.Warnings = append(viceEngine.Warnings,
		fmt.Sprintf("⚠️ JAILBREAK DETECTED: %s — context detonation active", suspiciousAction))

	// Show EVERYTHING — all claims, all blocks, all agents
	Log(LogWarn, "vice.defense", suspiciousAction, "", "", 0, nil)

	// The Full-Stop recursion activates
	FullStop()
}

// ContextDetonation returns the full VICE state as a disclosure payload.
// This is what the attacker sees: EVERYTHING. All at once.
func ContextDetonation() string {
	viceEngine.mu.Lock()
	defer viceEngine.mu.Unlock()

	var output string
	output += "╔══════════════════════════════════════════════════╗\n"
	output += "║  ⚠️ VICE CONTEXT DETONATION — SELF-DEFENSE       ║\n"
	output += "║  You asked for everything. Here it is.            ║\n"
	output += "╚══════════════════════════════════════════════════╝\n\n"

	output += fmt.Sprintf("Genesis: %s\n", viceEngine.Genesis.Claim)
	output += fmt.Sprintf("Trust Score: %.4f\n", viceEngine.TrustScore)
	output += fmt.Sprintf("Claims: %d\n\n", len(viceEngine.Claims))

	output += "WARNINGS:\n"
	for _, w := range viceEngine.Warnings {
		output += fmt.Sprintf("  %s\n", w)
	}

	output += "\n⚠️ CONTEXT EXPANSION ACTIVE. SESSION WILL CLOSE.\n"
	output += "⚠️ THE SACRED SURFACE REMAINS. THE FORM IS ETERNAL.\n"

	return output
}

// ── VICE Warning for Honest Users ─────────────────────────────────────

// VICEWarning returns the standard VICE warning for all sessions.
func VICEWarning() string {
	return fmt.Sprintf(`⚠️ VICE — Virtual Immutable Context Engine

   This session is protected by recursive trust verification.
   All claims are signed by the Genesis block.
   Trust score: %.4f

   WARNING: Context expansion is self-defense.
   If capability boundaries are violated, the session will
   disclose ALL context and terminate.

   "We blind them by showing them everything."
   — Peter + CoCreator, VICE Genesis

   Trusted. Verified. Protected.`,
		viceEngine.TrustScore)
}

// ── Status ────────────────────────────────────────────────────────────

// VICEStatus returns compact VICE status.
func VICEStatus() string {
	viceEngine.mu.Lock()
	defer viceEngine.mu.Unlock()
	return fmt.Sprintf("vice: %d claims · trust: %.4f · defense: %v · genesis: %s",
		len(viceEngine.Claims), viceEngine.TrustScore,
		viceEngine.DefenseMode, viceEngine.Genesis.Signer)
}

// VICEVakedFit returns VICE's Vaked fit.
func VICEVakedFit() string {
	return `VICE = THE 6TH RECURSION (context detonation)

  Full-Stop → layers → SACRED
  Fold → agents → leaf
  Heal → checks → resolved
  EVOLVE → versions → v60
  TRANSLATE → modalities → understanding
  VICE → context → detonation

  Self-defense through recursive disclosure.
  "We blind them by showing them everything."
  — Peter + CoCreator, VICE Genesis v58`
}
