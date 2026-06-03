package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/usewhale/whale/internal/core"
)

type ImmutablePrefix struct {
	systemBlocks []string
	fingerprint  string
}

func NewImmutablePrefix(systemBlocks []string) *ImmutablePrefix {
	p := &ImmutablePrefix{}
	p.Refresh(systemBlocks)
	return p
}

func (p *ImmutablePrefix) Refresh(systemBlocks []string) {
	p.systemBlocks = append([]string(nil), systemBlocks...)
	sum := sha256.Sum256([]byte(strings.Join(p.systemBlocks, "\n\n")))
	p.fingerprint = hex.EncodeToString(sum[:])
}

func (p *ImmutablePrefix) Fingerprint() string {
	return p.fingerprint
}

func (p *ImmutablePrefix) SystemBlocks() []string {
	if p == nil {
		return nil
	}
	return append([]string(nil), p.systemBlocks...)
}

func (p *ImmutablePrefix) ToMessages() []core.Message {
	if len(p.systemBlocks) == 0 {
		return nil
	}
	return []core.Message{{
		Role: core.RoleSystem,
		Text: strings.Join(p.systemBlocks, "\n\n"),
	}}
}
