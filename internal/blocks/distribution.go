package blocks

import "fmt"

// ── Distribution — Homebrew + Docker + Go Install ────────────────────
// v26: Package ultrawhale for all distribution channels.

// DistributionStatus returns distribution status.
func DistributionStatus() string {
	return fmt.Sprintf(`dist:
  brew:    brew install peterlodri-sec/ultrawhale/ultrawhale
  docker:  docker pull ghcr.io/peterlodri-sec/ultrawhale:%s
  go:      go install github.com/peterlodri-sec/ultrawhale/cmd/whale@%s
  binary:  https://github.com/peterlodri-sec/ultrawhale/releases/tag/%s`, CurrentVersion(), CurrentVersion(), CurrentVersion())
}
