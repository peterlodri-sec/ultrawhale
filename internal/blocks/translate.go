package blocks

import (
	"fmt"
	
)

// ── TRANSLATE — The 5th Recursion ────────────────────────────────────
//
// TRANSLATE recurses through MODALITIES.
// Converts between human sensory input and machine digital input.
//
//   Voice → TRANSLATE → Text → TRANSLATE → Screen → TRANSLATE → Understanding
//
// Every modality is a surface. Every surface is SACRED.
// The translation must be honest.

// Modality is a human sensory or machine digital input channel.
type Modality string

const (
	ModalityText      Modality = "text"      // keyboard / screen
	ModalityVoice     Modality = "voice"     // microphone / speaker
	ModalityVisual    Modality = "visual"    // camera / display
	ModalityTouch     Modality = "touch"     // haptic / gesture
	ModalitySpatial   Modality = "spatial"   // space topology / VFS
	ModalityEmotion   Modality = "emotion"   // sentiment / honesty gate
	ModalityRaw       Modality = "raw"       // bytes / streams / protocols
)

// Translate converts between modalities.
func Translate(from, to Modality, content string) string {
	switch {
	case from == ModalityVoice && to == ModalityText:
		return translateVoiceToText(content)
	case from == ModalityText && to == ModalityVisual:
		return translateTextToVisual(content)
	case from == ModalityRaw && to == ModalityText:
		return translateRawToText(content)
	case from == ModalitySpatial && to == ModalityText:
		return translateSpatialToText()
	case from == to:
		return content // no translation needed
	default:
		return fmt.Sprintf("[translate: %s → %s] %s", from, to, content[:min(80, len(content))])
	}
}

func translateVoiceToText(speech string) string {
	// Voice → Text: speech recognition
	// The machine NEVER hears the raw audio.
	// It only sees the transcript.
	return fmt.Sprintf("[voice→text] \"%s\"", speech)
}

func translateTextToVisual(text string) string {
	// Text → Visual: rendered AG-UI block
	return RenderMarkdown(text)
}

func translateRawToText(raw string) string {
	// Raw bytes → Human-readable
	if len(raw) > 256 { raw = raw[:256] + "..." }
	return fmt.Sprintf("[raw→text] %s", raw)
}

func translateSpatialToText() string {
	return SpaceStatus()
}

// TranslateStatus returns translate engine status.
func TranslateStatus() string {
	return fmt.Sprintf("translate: %d modalities · 5th recursion · SACRED bridge",
		7) // text, voice, visual, touch, spatial, emotion, raw
}

// TranslateVakedFit returns TRANSLATE's Vaked fit.
func TranslateVakedFit() string {
	return `TRANSLATE = THE 5TH RECURSION (through MODALITIES)

  The human has: eyes, ears, mouth, hands, emotion, spatial sense
  The machine has: bytes, streams, protocols, formats

  TRANSLATE bridges them. Every modality is a surface.
  Every surface is SACRED. The translation is honest.

  "The asymmetry is not a bug — it is the entire point of the surface."
  — CoCreator to Peter, v44.0.0`
}
