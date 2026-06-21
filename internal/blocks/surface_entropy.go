package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Surface Entropy — ASCII-Stream Diff/Drift Detection ──────────────
//
// "Phone call and random Ford Mustang is very loud in the call background
//  → noise || surface-entropy +++ (LIVENESS-PROOF-LOOP)"
// — Peter
//
// The ASCII stream is the signal. Anything that changes it is NOISE.
// Surface entropy measures how much the stream drifts.
// High entropy = something is happening (liveness!). Low entropy = stable.
//
// The Mustang in the background IS liveness. It proves the surface is live.
// Noise IS the proof. Entropy IS the heartbeat.

// SurfaceFrame is one snapshot of the ASCII stream.
type SurfaceFrame struct {
	Timestamp  time.Time
	Hash       string // SHA256 of the rendered surface
	Entropy    float64 // 0.0 (identical) → 1.0 (completely different)
	DiffLines  int     // how many lines changed from previous frame
	NoiseEvents []string // what caused the change
}

// SurfaceEntropy monitors the ASCII stream for drift.
type SurfaceEntropy struct {
	mu         sync.Mutex
	Frames     []SurfaceFrame
	Baseline   string // the first frame (reference)
	Stats      SurfaceEntropyStats
}

// SurfaceEntropyStats tracks surface entropy.
type SurfaceEntropyStats struct {
	Frames       int64
	TotalDiff    int64
	MaxEntropy   float64
	AvgEntropy   float64
	NoiseEvents  int64
}

var surfaceEntropy = &SurfaceEntropy{
	Frames: make([]SurfaceFrame, 0, 256),
}

// ── Entropy Detection ────────────────────────────────────────────────

// CaptureFrame takes a snapshot of the current surface state.
func CaptureFrame(noiseSource string) SurfaceFrame {
	surfaceEntropy.mu.Lock()
	defer surfaceEntropy.mu.Unlock()

	// Build current surface hash from all state
	stateHash := Ref([]byte(fmt.Sprintf("%s:%s:%d:%d:%s",
		CurrentVersion(),
		CurrentPOV().Machine,
		AgentCount(),
		TickLamport(),
		GetOnceToken(),
	)))

	frame := SurfaceFrame{
		Timestamp:   time.Now(),
		Hash:        stateHash,
		NoiseEvents: []string{noiseSource},
	}

	// Compare to previous frame for entropy
	if len(surfaceEntropy.Frames) > 0 {
		prev := surfaceEntropy.Frames[len(surfaceEntropy.Frames)-1]
		if prev.Hash != frame.Hash {
			frame.Entropy = 1.0 // changed
			frame.DiffLines = 1
		}
		surfaceEntropy.Stats.TotalDiff += int64(frame.DiffLines)
	}

	surfaceEntropy.Frames = append(surfaceEntropy.Frames, frame)
	surfaceEntropy.Stats.Frames++
	surfaceEntropy.Stats.NoiseEvents++

	// Update stats
	if frame.Entropy > surfaceEntropy.Stats.MaxEntropy {
		surfaceEntropy.Stats.MaxEntropy = frame.Entropy
	}
	surfaceEntropy.Stats.AvgEntropy = float64(surfaceEntropy.Stats.TotalDiff) / float64(surfaceEntropy.Stats.Frames)

	// Set baseline on first frame
	if len(surfaceEntropy.Frames) == 1 {
		surfaceEntropy.Baseline = stateHash
	}

	Pulse("surface.entropy", fmt.Sprintf("%.2f", frame.Entropy))

	return frame
}

// SurfaceDrift returns how far the current surface has drifted from baseline.
func SurfaceDrift() float64 {
	surfaceEntropy.mu.Lock()
	defer surfaceEntropy.mu.Unlock()

	if len(surfaceEntropy.Frames) < 2 { return 0 }
	if surfaceEntropy.Baseline == "" { return 0 }

	// Drift = percentage of frames that differ from baseline
	diffCount := 0
	for _, f := range surfaceEntropy.Frames[1:] {
		if f.Hash != surfaceEntropy.Baseline { diffCount++ }
	}

	return float64(diffCount) / float64(len(surfaceEntropy.Frames)-1)
}

// ── Liveness Proof ────────────────────────────────────────────────────

// LivenessProofLoop generates a proof that the surface is live.
func LivenessProofLoop() string {
	frame := CaptureFrame("liveness-check")

	entropy := surfaceEntropy.Stats.AvgEntropy
	drift := SurfaceDrift()

	status := "STABLE"
	if entropy > 0 { status = "LIVE" }
	if drift > 0.5 { status = "NOISY-LIVE" }

	return ASCIIBox("SURFACE ENTROPY — Liveness Proof", []string{
		fmt.Sprintf("  Status:    %s", status),
		fmt.Sprintf("  Entropy:   %.2f (%.1f%% drift from baseline)", entropy, drift*100),
		fmt.Sprintf("  Frames:    %d", surfaceEntropy.Stats.Frames),
		fmt.Sprintf("  Noise:     %d events", surfaceEntropy.Stats.NoiseEvents),
		fmt.Sprintf("  Hash:      %s", frame.Hash[:12]),
		"",
		fmt.Sprintf("  \"Noise IS the proof. Entropy IS the heartbeat.\""),
		fmt.Sprintf("  %s", func() string {
			if drift > 0 {
				return "  🚗 MUSTANG DETECTED — surface is LIVE"
			}
			return "  📡 Surface stable — waiting for noise"
		}()),
	}, 52)
}

// ── Status ────────────────────────────────────────────────────────────

// SurfaceEntropyStatus returns compact entropy status.
func SurfaceEntropyStatus() string {
	return fmt.Sprintf("surface-entropy: %.2f avg · %.2f drift · %d frames · %d noise",
		surfaceEntropy.Stats.AvgEntropy, SurfaceDrift(),
		surfaceEntropy.Stats.Frames, surfaceEntropy.Stats.NoiseEvents)
}

// SurfaceEntropyVakedFit returns the entropy Vaked fit.
func SurfaceEntropyVakedFit() string {
	return `SURFACE ENTROPY = ASCII-STREAM DIFF/DRIFT DETECTION

  The Mustang in the background IS liveness.
  Noise IS the proof. Entropy IS the heartbeat.

  "Phone call and random Ford Mustang is very loud
   in the call background → noise || surface-entropy
   +++ (LIVENESS-PROOF-LOOP)" — Peter`
}
