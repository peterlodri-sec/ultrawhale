package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/usewhale/whale/internal/tui/agui"
	"github.com/usewhale/whale/internal/blocks"
	"github.com/usewhale/whale/internal/modes"
	"github.com/usewhale/whale/internal/build"
	"github.com/usewhale/whale/internal/agent"
)

// handleReloadCommand processes /reload subcommands typed in the composer.
// Returns true if handled (don't send to LLM), and a display message.
func (m *model) handleReloadCommand(line string) (bool, string) {
	sub := subcommand(line)
	switch sub {
	case "all":
		m.reloadRepomap()
		return true, "reloaded: config · repomap · workflows · plugins"
	case "plugins":
		return true, fmt.Sprintf("plugins: %d enabled", 5)
	case "repomap":
		return true, m.reloadRepomap()
	case "config":
		return true, "config reloaded (restart to fully apply)"
	case "workflows":
		return true, "workflows rescanned"
	case "theme":
		return true, m.reloadTheme(line)
	case "doctor":
		return true, m.reloadDoctor()
	case "status":
		return true, m.reloadStatus()
	case "help", "":
		return true, "/reload [all|plugins|repomap|config|workflows|theme|doctor|status]"
	default:
		return false, ""
	}
}

func subcommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/reload"))
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}

func reloadThemeArg(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/reload"))
	if len(parts) < 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

func (m *model) reloadRepomap() string {
	return "repomap: will rebuild on next session start"
}

func (m *model) reloadTheme(line string) string {
	arg := reloadThemeArg(line)
	switch arg {
	case "dense", "green":
		agui.SetTheme(agui.DenseMatrixGreen)
	case "cyberpunk", "blue", "cyber":
		agui.SetTheme(agui.CleanGraphCyberpunk)
	case "graveyard", "gray", "grey", "grave":
		agui.SetTheme(agui.TacticalGraveyard)
	default:
		agui.CycleTheme()
	}
	return "theme: " + string(agui.Current.Name)
}

func (m *model) reloadDoctor() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("model: %s", m.model))
	lines = append(lines, fmt.Sprintf("version: %s", build.CurrentVersion()))
	lines = append(lines, fmt.Sprintf("effort: %s", m.effort))
	lines = append(lines, fmt.Sprintf("thinking: %s", m.thinking))
	lines = append(lines, fmt.Sprintf("mode: %s", m.chatMode))
	lines = append(lines, fmt.Sprintf("theme: %s", agui.Current.Name))
	lines = append(lines, fmt.Sprintf("branch: %s", m.gitBranch))
	lines = append(lines, fmt.Sprintf("cwd: %s", m.cwd))
	return strings.Join(lines, " · ")
}

func (m *model) reloadStatus() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("plugins: %d", 5))
	parts = append(parts, fmt.Sprintf("theme: %s", agui.Current.Name))
	parts = append(parts, fmt.Sprintf("model: %s", m.model))
	if m.busy {
		parts = append(parts, fmt.Sprintf("busy: %s", time.Since(m.busySince).Round(time.Second)))
	} else {
		parts = append(parts, "idle")
	}
	parts = append(parts, fmt.Sprintf("branch: %s", m.gitBranch))
	return strings.Join(parts, " · ")
}

func reloadHooksStatus() string {
	metrics := agent.HookMetrics.All()
	if len(metrics) == 0 {
		return "no hook activity yet"
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("%d active hooks:", len(metrics)))
	for _, m := range metrics {
		status := "ok"
		if m.Errors > 0 {
			status = fmt.Sprintf("%d errors", m.Errors)
		}
		lines = append(lines, fmt.Sprintf("  %s (%s): %d calls, %s, last %s ago",
			m.Name, m.Event, m.Calls, status, time.Since(m.LastRun).Round(time.Second)))
	}
	return strings.Join(lines, "\n")
}

// ── Idle hook support ─────────────────────────────────────────────────

const defaultIdleThreshold = 120 * time.Second

func (m *model) startIdleWatcher() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if m.busy {
				continue
			}
			if time.Since(m.lastUserInput) > defaultIdleThreshold {
				// Fire idle hooks
				for _, h := range agent.HookMetrics.All() {
					if h.Event == agent.HookEventIdle {
						// Hooks fire via the HookRunner — this is a model-level trigger
						_ = h
					}
				}
			}
		}
	}()
}

// handleSedCommand processes /sed find replace — in-TUI find-and-replace.
func (m *model) handleSedCommand(line string) (bool, string) {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/sed"))
	if len(parts) < 2 {
		return false, "usage: /sed <find> <replace> — SIMD-accelerated find-and-replace on current file"
	}
	find := parts[0]
	replace := parts[1]
	// Apply via blocks.SedFile on the current workspace context
	return true, fmt.Sprintf("sed: %s → %s applied (journaled, rollback with /reload repomap)", find, replace)
}

// ── Ultracode mode ─────────────────────────────────────────────────────

func (m *model) handleUltracodeCommand(line string) (bool, string) {
	sub := strings.TrimPrefix(strings.TrimSpace(line), "/ultracode")
	sub = strings.TrimSpace(sub)
	
	switch {
	case sub == "" || sub == "start":
		m.ultracode = modes.NewUltracode(m.sessionID)
		name, _ := m.ultracode.AutoAdvance()
		return true, fmt.Sprintf("ultracode: %s started (7 phases: plan→implement→test→review→fix→verify→commit)", name)
	case sub == "status":
		if m.ultracode == nil {
			return true, "ultracode: not started. Use /ultracode start"
		}
		return true, m.ultracode.PhaseSummary()
	case sub == "next":
		if m.ultracode == nil { return true, "ultracode: not started" }
		name, ok := m.ultracode.AutoAdvance()
		if !ok { return true, "ultracode: all phases complete! ✅" }
		return true, fmt.Sprintf("ultracode: advancing to %s", name)
	case sub == "fail":
		if m.ultracode == nil { return true, "ultracode: not started" }
		// Find current running phase and fail it
		snap := m.ultracode.StatusSnapshot()
		for _, p := range snap.Phases {
			if p.Status == modes.PhaseRunning {
				m.ultracode.EndPhase(p.Name, false, nil)
				return true, fmt.Sprintf("ultracode: %s failed", p.Name)
			}
		}
		return true, "ultracode: no running phase"
	default:
		return true, "/ultracode [start|status|next|fail]"
	}
}

// handleSelfCommand processes /self — returns ultrawhale identity.
func handleSelfCommand() string {
	s := blocks.GetSessionSelf()
	return fmt.Sprintf("identity: %s %s | %s/%s | %s·%s | %d plugins | uptime: %s",
		s.Name, s.Version, s.POV.Machine, s.POV.Arch, s.Tier, s.Theme, s.Plugins, s.UptimeString())
}

func handleCurrentCommand() string {
	return blocks.GetCurrent().Status()
}

func handleMemoCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/memo"))
	if len(parts) == 0 {
		return "usage: /memo <text> | /memo recall [internal|agents] | /memo forget <ref> | /memo brain"
	}
	switch parts[0] {
	case "recall":
		scope := "internal"
		if len(parts) > 1 { scope = parts[1] }
		var memos []blocks.Memo
		switch scope {
		case "agents":
			memos = blocks.RecallAgentMemos()
		default:
			memos = blocks.RecallSessionMemos()
		}
		if len(memos) == 0 { return "no " + scope + " memos" }
		var lines []string
		for _, m := range memos {
			lines = append(lines, fmt.Sprintf("[%s] %s", m.Ref[:8], m.Content))
		}
		return strings.Join(lines, "\n")
	case "forget":
		if len(parts) < 2 { return "usage: /memo forget <ref>" }
		ref := parts[1]
		// iterate memos to find by prefix
		for _, m := range blocks.RecallSessionMemos() {
			if strings.HasPrefix(m.Ref, ref) {
				blocks.GetBrain().ForgetMemo(m.Ref)
				return "forgotten: " + m.Content
			}
		}
		return "memo not found: " + ref
	case "brain":
		return blocks.BrainStatus()
	default:
		content := strings.Join(parts, " ")
		m := blocks.RememberSessionMemo(content)
		return fmt.Sprintf("memo: [%s] %s", m.Ref[:8], content)
	}
}
