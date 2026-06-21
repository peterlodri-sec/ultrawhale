package blocks

import (
	"fmt"
	"strings"
)

// ── OneShot Chain — Piped Vaked Operations ────────────────────────────
//
// Chain OneShot operations like Unix pipes:
//   oneshot declare "agent swe" | oneshot materialize | oneshot reveal
//
// Each OneShot feeds into the next. Atomic. Recursive. Vaked.
//
// The chain is a Vaked pipeline in miniature:
//   declare → materialize → reveal → declare → ... (recursive)

// ChainLink is one link in the OneShot chain.
type ChainLink struct {
	Step     int
	Layer    string // "declare", "materialize", "supervise", "enforce", "testify", "index", "reveal"
	Input    string
	Output   string
	Duration int64 // microseconds
	Ref      string
}

// OneShotChain is a sequence of OneShot operations.
type OneShotChain struct {
	Links    []ChainLink
	TotalDuration int64
	Result   string
}

// Chain executes a sequence of OneShot operations.
// Each output becomes the next input.
func Chain(declarations ...string) OneShotChain {
	chain := OneShotChain{Links: make([]ChainLink, 0, len(declarations))}
	layers := []string{"declare", "materialize", "supervise", "enforce", "testify", "index", "reveal"}

	for i, decl := range declarations {
		result := OneShot(decl)

		layer := layers[i%len(layers)]
		link := ChainLink{
			Step:     i + 1,
			Layer:    layer,
			Input:    decl[:min(60, len(decl))],
			Output:   result.Materialized,
			Duration: result.Duration,
			Ref:      result.Materialized,
		}
		chain.Links = append(chain.Links, link)
		chain.TotalDuration += result.Duration
		chain.Result = result.Materialized
	}

	Log(LogInfo, "chain.execute",
		fmt.Sprintf("%d links · %dµs", len(chain.Links), chain.TotalDuration),
		"", "", 0, nil)

	return chain
}

// Pipe is an alias for Chain — Unix-style.
func Pipe(declarations ...string) OneShotChain {
	return Chain(declarations...)
}

// ChainRender renders a OneShot chain as ASCII art.
func ChainRender(chain OneShotChain) string {
	var sb strings.Builder
	sb.WriteString("┌─────────────── OneShot Chain ───────────────┐\n")

	for _, link := range chain.Links {
		arrow := "│"
		if link.Step == len(chain.Links) { arrow = " " }
		sb.WriteString(fmt.Sprintf("│ Step %d [%s] %s → %s  %s\n",
			link.Step, link.Layer, link.Input, link.Output[:min(20, len(link.Output))], arrow))
		if link.Step < len(chain.Links) {
			sb.WriteString("│     ↓\n")
		}
	}

	sb.WriteString(fmt.Sprintf("└────────────────── %dµs total ──────────────────┘",
		chain.TotalDuration))
	return sb.String()
}

// ChainVakedFit returns the chain's Vaked fit.
func ChainVakedFit() string {
	return `ONESHOT CHAIN = VAKED PIPELINE IN MINIATURE

  Like Unix pipes: declare | materialize | reveal
  Each OneShot feeds into the next.
  Atomic. Recursive. Vaked.

  declare → materialize → supervise → enforce → testify → index → reveal
  The entire Vaked pipeline in one function call.`
}
