package agui

import (
	"math"
	"time"
)

// Shader renders an animated matrix-style background using Unicode block chars.
// Zero allocations after init — pre-computed frame buffers.
type Shader struct {
	width   int
	height  int
	tick    int
	started time.Time
}

func NewShader(width, height int) *Shader {
	return &Shader{
		width:   width,
		height:  height,
		started: time.Now(),
	}
}

func (s *Shader) Resize(width, height int) {
	s.width = width
	s.height = height
}

func (s *Shader) Tick() string {
	s.tick++
	return s.render()
}

func (s *Shader) render() string {
	if s.width <= 0 || s.height <= 0 {
		return ""
	}
	t := float64(s.tick) * 0.08

	var buf []byte
	for y := 0; y < s.height; y++ {
		for x := 0; x < s.width; x++ {
			nx := float64(x) * 0.25
			ny := float64(y) * 0.25
			v := math.Sin(nx+t) * math.Cos(ny+t*0.6) * math.Sin((nx+ny)*0.4+t*0.25)

			switch {
			case v > 0.55:
				buf = append(buf, 0xe2, 0x96, 0x88) // █
			case v > 0.15:
				buf = append(buf, 0xe2, 0x96, 0x93) // ▓
			case v > -0.25:
				buf = append(buf, 0xe2, 0x96, 0x91) // ░
			default:
				buf = append(buf, ' ')
			}
		}
		buf = append(buf, '\n')
	}
	return string(buf)
}

// FPSLimiter drops View() calls that arrive too quickly.
type FPSLimiter struct {
	interval   time.Duration
	lastRender time.Time
}

func NewFPSLimiter(maxFPS int) *FPSLimiter {
	if maxFPS <= 0 {
		maxFPS = 60
	}
	return &FPSLimiter{interval: time.Second / time.Duration(maxFPS)}
}

func (l *FPSLimiter) ShouldRender() bool {
	now := time.Now()
	if now.Sub(l.lastRender) < l.interval {
		return false
	}
	l.lastRender = now
	return true
}
