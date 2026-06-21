package blocks

import (
	"fmt"
	"math"
	
	"sync"
	"time"
)

// ── [RADIO] — Live Lo-Fi Coding Music ────────────────────────────────
//
// RADIO = Testifies layer made audible.
// Every probe, every pulse, every ring becomes a sound.
// USER↔ultrawhale: by-definition-unique music generation.
//
// Components:
//   Beat:    code activity → BPM (commits per minute, agent spawns)
//   Melody:  Vaked layer health → notes (Declares→Reveals mapped to scale)
//   Bass:    recursion depth → sub frequencies (Fold depth → Hz)
//   Ambience: space topology → pad layers (node count → density)

// RadioStation generates live music from Vaked state.
type RadioStation struct {
	mu       sync.Mutex
	Active   bool
	Name     string    // "ultrawhale-radio"
	Genre    string    // "lo-fi-coding"
	Stats    RadioStats
	Listeners int
}

// RadioStats tracks music generation.
type RadioStats struct {
	BeatsGenerated   int64
	MelodiesGenerated int64
	NotesPlayed      int64
	CurrentBPM       int
	CurrentKey       string
	SessionID        string
}

var radio = &RadioStation{
	Name:     "ultrawhale-radio",
	Genre:    "lo-fi-coding",
	Listeners: 1, // the human
}

// ── Music Generation ─────────────────────────────────────────────────

// RadioStart begins live music generation.
func RadioStart() string {
	radio.mu.Lock()
	defer radio.mu.Unlock()

	if radio.Active { return "📻 already broadcasting" }
	radio.Active = true
	radio.Stats.SessionID = fmt.Sprintf("radio-%s", time.Now().Format("20060102-150405"))

	Log(LogInfo, "radio.start", radio.Stats.SessionID, "", "", 0, nil)
	Pulse("radio.start", radio.Stats.SessionID)

	return fmt.Sprintf("📻 %s · %s · live", radio.Name, radio.Genre)
}

// RadioStop stops music generation.
func RadioStop() string {
	radio.mu.Lock()
	defer radio.mu.Unlock()
	radio.Active = false
	return fmt.Sprintf("📻 off · %d beats · %d notes", radio.Stats.BeatsGenerated, radio.Stats.NotesPlayed)
}

// RadioNow generates the current musical state as description.
func RadioNow() string {
	radio.mu.Lock()
	defer radio.mu.Unlock()

	// Derive parameters from Vaked state
	bpm := deriveBPM()
	key := deriveKey()
	bassNote := deriveBass()
	ambience := deriveAmbience()
	melody := deriveMelody()

	radio.Stats.BeatsGenerated++
	radio.Stats.CurrentBPM = bpm
	radio.Stats.CurrentKey = key

	return fmt.Sprintf("🎵 [%s] %s · %s
├─ 🥁 %d BPM (%d agents active)
├─ 🎹 %s (%s layer health)
├─ 🎸 %s (fold depth %d)
├─ 🌫️  %s (%d space nodes)
└─ 🎧 %s · lo-fi · unique",
		radio.Name, radio.Genre,
		bpm, AgentCount(),
		key, layerHealthSummary(),
		bassNote, currentFoldDepth(),
		ambience, spaceNodeCount(),
		melody)
}

func deriveBPM() int {
	agents := AgentCount()
	if agents == 0 { return 60 } // default lo-fi BPM
	return 60 + agents*10
}

func deriveKey() string {
	keys := []string{"Cmaj7", "Dm9", "Em7", "Fmaj7", "G7", "Am7", "Bm7♭5"}
	idx := AgentCount() % len(keys)
	return keys[idx]
}

func deriveBass() string {
	depth := FoldDepth("")
	if depth < 0 { depth = 0 }
	freq := 55.0 * math.Pow(2, float64(depth)/12.0) // low A, rising with depth
	return fmt.Sprintf("%.0fHz", freq)
}

func deriveAmbience() string {
	nodes := spaceNodeCount()
	switch {
	case nodes == 0: return "silence"
	case nodes < 5: return "light pad"
	case nodes < 10: return "warm pad"
	default: return "full ambience"
	}
}

func deriveMelody() string {
	health := layerHealthSummary()
	return fmt.Sprintf("♪ %s", health)
}

func layerHealthSummary() string {
	if IsSacredHealthy() { return "healthy" }
	return "degraded"
}

func currentFoldDepth() int {
	return FoldDepth("")
}

func spaceNodeCount() int {
	spaceTopology.mu.Lock()
	defer spaceTopology.mu.Unlock()
	return len(spaceTopology.Nodes)
}

// ── Radio Status ──────────────────────────────────────────────────────

// RadioStatus returns compact radio status.
func RadioStatus() string {
	radio.mu.Lock()
	defer radio.mu.Unlock()
	if !radio.Active { return "📻 radio: off" }
	return fmt.Sprintf("📻 %s · %d BPM · %s · %d listeners",
		radio.Name, radio.Stats.CurrentBPM, radio.Stats.CurrentKey, radio.Listeners)
}

// RadioVakedFit returns RADIO's Vaked fit.
func RadioVakedFit() string {
	return `RADIO = TESTIFIES LAYER MADE AUDIBLE

  Every probe → beat
  Every layer → note
  Every recursion → bass
  Every node → ambience

  The system IS the music. You code. It plays.
  USER↔ultrawhale: by-definition-unique.

  "Co-wise unify space. The radio never stops."`
}
