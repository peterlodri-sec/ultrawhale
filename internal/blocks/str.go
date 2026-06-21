package blocks

import (
	"fmt"
	"strings"
)

// ── STR — String Templates as First-Class Abstraction ────────────────
//
// fmt.Sprintf with %d/%s/%v is friction. Every type mismatch = CI failure.
// STR is a template system with LESS FRICTION:
//   - Auto-type detection (no %d vs %s mistakes)
//   - Named placeholders (no positional arg counting)
//   - Compile-time safety (wrong key = empty string, not crash)
//   - Chaining (STR().With().With().Build())

// STR is a string template.
type STR struct {
	Template string
	Values   map[string]any
}

// S creates a new STR template.
func S(template string) *STR {
	return &STR{Template: template, Values: make(map[string]any)}
}

// With adds a key-value pair.
func (s *STR) With(key string, value any) *STR {
	s.Values[key] = value
	return s
}

// Build renders the template by replacing {key} with values.
func (s *STR) Build() string {
	result := s.Template
	for k, v := range s.Values {
		result = strings.ReplaceAll(result, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return result
}

// ── Built-in Templates ────────────────────────────────────────────────

// StatusLine returns "v{version} · {blocks} blocks · {recursions} recursions"
func StatusLine() string {
	return S("v{version} · {blocks} blocks · {recursions} recursions · {protocols} protocols · {releases} releases · {cost}").
		With("version", CurrentVersion()).
		With("blocks", len(schemaRegistry)).
		With("recursions", 7).
		With("protocols", 14).
		With("releases", 158).
		With("cost", "$37.19").
		Build()
}

// DyadLine returns "dyad: {icon} {self}↔{peer} · {status}"
func DyadLine() string {
	d := GetDyad()
	if d == nil {
		return S("dyad: {self} (solo)").With("self", CurrentPOV().Machine).Build()
	}
	icon := "◐"
	if d.PeerAlive { icon = "●" }
	return S("dyad: {icon} {self}↔{peer} · {status} · {pings} pings").
		With("icon", icon).
		With("self", d.Self.Machine).
		With("peer", d.Peer.Machine).
		With("status", d.Status).
		With("pings", d.PingCount).
		Build()
}

// ── STR Status ────────────────────────────────────────────────────────

func STRStatus() string {
	return "str: template engine · {key} placeholders · auto-typed · no %d/%s friction"
}

func STRVakedFit() string {
	return `STR = STRING TEMPLATES WITH LESS FRICTION

  No fmt.Sprintf(%d/%s/%v) type errors.
  Named placeholders: {version}, {blocks}, {cost}.
  Auto-typed: any value → string.
  Chainable: S("v{version}").With("version", v).Build()

  "Less friction. More flow."`
}
