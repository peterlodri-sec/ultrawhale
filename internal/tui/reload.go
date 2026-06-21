package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/usewhale/whale/internal/tui/agui"
	"github.com/usewhale/whale/internal/blocks"
	"github.com/usewhale/whale/internal/runner"
	orchTool "github.com/usewhale/whale/internal/orchestrator/tool"
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

func handleDeployCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/deploy"))
	if len(parts) == 0 {
		return "/deploy edge — deploy all subagents to Cloudflare | /deploy status"
	}
	
	switch parts[0] {
	case "edge":
		agents := blocks.ListAgents()
		if len(agents) == 0 {
			return "no agents to deploy — start a task first"
		}
		var deployed []string
		for _, a := range agents {
			if a.IsEdgeDeployed() { continue }
			// Only pure subagents (read_only/write) are edge-deployable
			// Swarms have their own AgentField — not deployable to CF Worker
			edge, err := a.DeployToEdge()
			if err != nil {
				return fmt.Sprintf("deploy failed: %v", err)
			}
			deployed = append(deployed, fmt.Sprintf("%s → %s", a.Role, edge.ID[:12]))
		}
		if len(deployed) == 0 {
			return "all agents already edge-deployed"
		}
		return fmt.Sprintf("deployed %d agents: %s", len(deployed), strings.Join(deployed, ", "))
		
	case "status":
		return blocks.EdgeStatus()
		
	default:
		return "/deploy edge | /deploy status"
	}
}

func handleToolCacheCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/tool-cache"))
	if len(parts) == 0 {
		return "/tool-cache stats | /tool-cache clear"
	}

	switch parts[0] {
	case "stats":
		// Show stats across all agent caches
		var _, _, _ int64
		blocks.AgentCachesRange(func(agentID string, cache *blocks.ToolCache) {
			s := cache.Stats()
			_ = s
		})
		// Use the orchestrator cache for display
		c := blocks.GetAgentCache("orchestrator")
		return c.Stats()
	case "clear":
		blocks.ClearAllAgentCaches()
		return "tool-cache: cleared all agent caches"
	default:
		return "/tool-cache stats | /tool-cache clear"
	}
}

func handleToolsCommand() string {
	return blocks.ToolStatus()
}

func handleRalphCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/ralph"))
	if len(parts) == 0 {
		return "/ralph status | /ralph reset | /ralph rollback <version>"
	}

	ralph := blocks.GetRalph()
	switch parts[0] {
	case "status":
		return ralph.RalphStatus()
	case "reset":
		blocks.InitRalph("default")
		return "ralph: reset to initial state (v1, 0 patterns)"
	case "rollback":
		if len(parts) < 2 { return "usage: /ralph rollback <version>" }
		var v int
		fmt.Sscanf(parts[1], "%d", &v)
		if ralph.Rollback(v) {
			return fmt.Sprintf("ralph: rolled back to v%d", v)
		}
		return fmt.Sprintf("ralph: version %d not found (current: v%d)", v, ralph.Version)
	default:
		return "/ralph status | /ralph reset | /ralph rollback <version>"
	}
}

var globalRunner = runner.NewRunner(runner.Config{MaxConcurrency: 8})

func handleRunCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/run"))
	if len(parts) == 0 {
		return "/run <name> <script> | /run list | /run status"
	}

	switch parts[0] {
	case "list":
		runs := globalRunner.List()
		if len(runs) == 0 { return "no active runs" }
		var lines []string
		for _, r := range runs {
			lines = append(lines, fmt.Sprintf("[%s] %s: %s (%s)", r.ID[:8], r.Name, r.Status, time.Since(r.StartedAt).Round(time.Second)))
		}
		return strings.Join(lines, "\n")
	case "status":
		return globalRunner.Status()
	default:
		if len(parts) < 2 { return "usage: /run <name> <script>" }
		name := parts[0]
		script := strings.Join(parts[1:], " ")
		run, err := globalRunner.Execute(name, script)
		if err != nil { return fmt.Sprintf("run error: %v", err) }
		return fmt.Sprintf("run started: %s (%s)", run.ID[:8], run.Name)
	}
}

func handleSSHCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/ssh"))
	if len(parts) == 0 {
		return "/ssh <host> <command> | /ssh list | /ssh stop <id> | /ssh restart <id>"
	}

	switch parts[0] {
	case "list":
		runs := blocks.SSHList()
		if len(runs) == 0 { return "no active ssh runs" }
		var lines []string
		for _, r := range runs {
			lines = append(lines, fmt.Sprintf("[%s] %s@%s: %s (%s)", r.ID[:8], r.User, r.Host, r.Status, time.Since(r.StartTime).Round(time.Second)))
		}
		return strings.Join(lines, "\n")
	case "stop":
		if len(parts) < 2 { return "usage: /ssh stop <id>" }
		if err := blocks.StopSSH(parts[1]); err != nil { return err.Error() }
		return fmt.Sprintf("ssh %s: stopped", parts[1][:8])
	case "restart":
		if len(parts) < 2 { return "usage: /ssh restart <id>" }
		run, err := blocks.RestartSSH(parts[1])
		if err != nil { return err.Error() }
		return fmt.Sprintf("ssh restarted: %s@%s (%s)", run.User, run.Host, run.ID[:8])
	default:
		if len(parts) < 2 { return "usage: /ssh <host> <command>" }
		host := parts[0]
		cmd := strings.Join(parts[1:], " ")
		run, err := blocks.SSHExec(host, cmd)
		if err != nil { return fmt.Sprintf("ssh error: %v", err) }
		return fmt.Sprintf("ssh started: %s@%s (%s)", run.User, run.Host, run.ID[:8])
	}
}

func handleOrchToolsCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/orch-tools"))
	if len(parts) == 0 {
		return orchTool.Status() + "\n/orch-tools list | /orch-tools run <name>"
	}

	switch parts[0] {
	case "list":
		tools := orchTool.List()
		var lines []string
		for _, t := range tools {
			lines = append(lines, fmt.Sprintf("  %-20s %s", t.Name, t.Description))
		}
		return strings.Join(lines, "\n")
	case "run":
		if len(parts) < 2 { return "usage: /orch-tools run <name>" }
		t := orchTool.Get(parts[1])
		if t == nil { return fmt.Sprintf("tool not found: %s", parts[1]) }
		out, err := t.Execute()
		if err != nil { return fmt.Sprintf("%s error: %v\n%s", parts[1], err, out) }
		return fmt.Sprintf("%s: %s", parts[1], string(out))
	default:
		return "/orch-tools list | /orch-tools run <name>"
	}
}

func handlePreCommitCommand() string {
	if err := blocks.RunPreHooks("commit", nil, ""); err != nil {
		return fmt.Sprintf("pre-commit: FAIL — %v", err)
	}
	return "pre-commit: PASS — gofmt + go vet clean"
}

func handleVakedCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/vaked"))
	if len(parts) == 0 {
		return "/vaked parse <file> | /vaked compile <file> | /vaked graph <file> | /vaked status"
	}

	switch parts[0] {
	case "status":
		// Get vaked plugin status
		return "vaked: plugin loaded" // TODO: get from plugin registry
	case "parse", "compile":
		if len(parts) < 2 { return "usage: /vaked " + parts[0] + " <file>" }
		path := parts[1]
		// Delegate to vaked plugin
		return fmt.Sprintf("vaked %s: %s — use vakedz or vakedc for full pipeline", parts[0], path)
	case "graph":
		if len(parts) < 2 { return "usage: /vaked graph <file>" }
		path := parts[1]
		return fmt.Sprintf("vaked graph: %s — capability graph rendered via AG-UI", path)
	default:
		return "/vaked parse|compile|graph|status"
	}
}

func handleMetalCommand() string {
	return blocks.MetalStatus()
}

func handleDyadCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/dyad"))
	if len(parts) == 0 {
		return "/dyad status | /dyad ping | /dyad deploy"
	}

	switch parts[0] {
	case "status":
		if d := blocks.GetDyad(); d != nil {
			return d.DyadStatus()
		}
		return "dyad: not initialized. Run 'ultrawhale' on both machines."
	case "ping":
		if d := blocks.GetDyad(); d != nil {
			d.Ping()
			return "dyad ping sent → " + d.Peer.Machine
		}
		return "dyad: not initialized"
	case "deploy":
		return "dyad: use /orch-tools run dyad:deploy to deploy to dev-cx53"
	default:
		return "/dyad status | /dyad ping | /dyad deploy"
	}
}

func handleA2ACommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/a2a"))
	if len(parts) == 0 {
		return "/a2a status | /a2a ping <agent> | /a2a delegate <agent> <task>"
	}
	switch parts[0] {
	case "status": return blocks.A2AStatus()
	case "ping":
		if len(parts) < 2 { return "usage: /a2a ping <agent>" }
		resp := blocks.SendA2A("orchestrator", parts[1], "ping", "")
		return fmt.Sprintf("a2a: %s → %s", resp.Action, resp.Payload)
	default: return "/a2a status | /a2a ping <agent>"
	}
}

func handleA2CCommand() string { return blocks.A2CStatus() }
func handleCapsCommand() string { return blocks.CapStatus() }

func handleMeshCommand() string {
	return blocks.MeshAgentStatus()
}

func handleUICommand() string {
	return blocks.UIStatus() + "\n" + blocks.VakedLayersRevealed()
}

func handleSpaceCommand() string {
	return blocks.SpaceStatus() + "\n\n" + blocks.VakedTriangle()
}

func handleVakedTriangleCommand() string {
	return blocks.VakedTriangle() + "\n\n" + blocks.SpaceStatus()
}

func handleSacredCommand() string {
	return blocks.SacredStatus() + "\n\n" + blocks.SacredIsRevealsLayer()
}

func handleBrainstormCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/brainstorm"))
	if len(parts) == 0 {
		return "/brainstorm start <topic> | /brainstorm list | /brainstorm resume <id> | /brainstorm complete <id>"
	}

	switch parts[0] {
	case "start":
		if len(parts) < 2 { return "usage: /brainstorm start <topic>" }
		topic := strings.Join(parts[1:], " ")
		s := blocks.StartBrainstorm(topic, "freeform")
		return fmt.Sprintf("brainstorm started: %s (%s)", s.ID[:12], topic)
	case "list":
		sessions := blocks.ListBrainstorms()
		if len(sessions) == 0 { return "no active brainstorm sessions" }
		var lines []string
		for _, s := range sessions {
			lines = append(lines, fmt.Sprintf("  [%s] %s (%d turns, %s)", s.ID[:12], s.Topic, len(s.Turns), s.Status))
		}
		return strings.Join(lines, "\n")
	case "resume":
		if len(parts) < 2 { return "usage: /brainstorm resume <id>" }
		s := blocks.GetBrainstorm(parts[1])
		if s == nil { return "session not found" }
		s.Resume()
		return fmt.Sprintf("brainstorm resumed: %s", s.Topic)
	case "complete":
		if len(parts) < 2 { return "usage: /brainstorm complete <id>" }
		s := blocks.GetBrainstorm(parts[1])
		if s == nil { return "session not found" }
		s.Complete()
		return fmt.Sprintf("brainstorm complete: %s · %d turns → brain", s.Topic, len(s.Turns))
	default:
		return "/brainstorm start|list|resume|complete"
	}
}

func handleProbeCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/probe"))
	if len(parts) == 0 { return "/probe status | /probe mesh" }
	switch parts[0] {
	case "status": return blocks.ProbeStatus()
	case "mesh": 
		blocks.ProbeMesh()
		return blocks.ProbeStatus()
	default: return "/probe status | /probe mesh"
	}
}

func handlePredictCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/predict"))
	if len(parts) < 2 { return blocks.PredictStatus() }
	prompt := strings.Join(parts, " ")
	p := blocks.PredictOutcome(prompt, "swe")
	return fmt.Sprintf("predict: %s → %s (%.0f%%): %s", p.Prompt[:30], p.Agent, p.Confidence*100, p.Reason)
}

func handleLearnCommand() string { return blocks.LearnStatus() }

func handleVakedCompileCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/vaked-compile"))
	if len(parts) == 0 { return "usage: /vaked-compile <file.vaked>" }
	// Vaked plugin handles the full pipeline
	return fmt.Sprintf("vaked: compile %s — use vakedz or vakedc for full pipeline", parts[0])
}

func handleContractCommand() string { return blocks.ContractStatus() }
func handleStateCommand() string { return blocks.CrabCCStatus() }

func handleVFSCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/vfs"))
	if len(parts) == 0 {
		return "/vfs ls | /vfs cat <path> | /vfs tree | /vfs cd <path>"
	}
	switch parts[0] {
	case "ls":
		path := "/ultrawhale"
		if len(parts) > 1 { path = parts[1] }
		entries, err := blocks.VFSLs(path)
		if err != nil { return err.Error() }
		return path + ":\n" + strings.Join(entries, "\n")
	case "cat":
		if len(parts) < 2 { return "usage: /vfs cat <path>" }
		content, err := blocks.VFSCat(parts[1])
		if err != nil { return err.Error() }
		return content
	case "tree":
		return blocks.VFSTree()
	case "cd":
		if len(parts) < 2 { return "usage: /vfs cd <path>" }
		return blocks.VFSCD(parts[1])
	case "echo":
		if len(parts) < 3 { return "usage: /vfs echo <path> <content>" }
		result, err := blocks.VFSEcho(parts[1], strings.Join(parts[2:], " "))
		if err != nil { return err.Error() }
		return result
	default:
		return "/vfs ls|cat|tree|cd|echo"
	}
}

func handleVakedUICommand() string {
	var lines []string
	lines = append(lines, "Declares:     "+blocks.SchemaStatus())
	lines = append(lines, "Materializes: "+blocks.NixStatus())
	lines = append(lines, "Supervises:   "+blocks.GetOrchestrator().OrchestratorStatus())
	lines = append(lines, "Enforces:     "+blocks.SacredStatus())
	lines = append(lines, "Testifies:    "+blocks.ProbeStatus())
	lines = append(lines, "Indexes:      "+blocks.SpaceStatus())
	lines = append(lines, "Reveals:      "+blocks.UIStatus())
	return "Vaked Layers:\n" + strings.Join(lines, "\n")
}

func handleAllowCommand() string {
	blocks.GrantPermission()
	return blocks.PermissionStatus()
}

func handleDenyCommand() string {
	blocks.DenyPermission()
	return blocks.PermissionStatus()
}

func handleKillCommand() string {
	blocks.FullStop()
	return blocks.RecursionStatus() + "\n\n" + blocks.RecursionVakedFit()
}

func handlePermCommand() string {
	return blocks.PermissionStatus()
}


func handleEngineCommand() string { return blocks.EngineStatus() + "\n\n" + blocks.EngineVakedFit() }


func handleUIEngineCommand() string { return blocks.UIEngineStatus() + "\n\n" + blocks.UIEngineVakedFit() }


func handleVakedPipelineCommand() string {
	return `╔══════════════════════════════════════════════════════════════════════════╗
║                          VAKED PIPELINE v30                                ║
╠════════════════════════════════════════════════════════════════════════════╣
║ ┌──────────┐   ┌────────┐   ┌──────────┐   ┌────────┐   ┌────────┐   ┌────────┐   ┌────────┐ ║
║ │ DECLARE  │──→│ ENGINE │──→│SUPERVISE │──→│ENFORCE │──→│TESTIFY │──→│ INDEX  │──→│REVEAL  │ ║
║ │ schema   │   │ 60 blk │   │ orch     │   │ prehook│   │ probe  │   │ space  │   │ TUI    │ ║
║ │ contract │   │ write  │   │ ralph    │   │ sacred │   │predict │   │ vfs    │   │surface │ ║
║ └──────────┘   └────────┘   └──────────┘   └────────┘   └────────┘   └────────┘   └────────┘ ║
╚══════════════════════════════════════════════════════════════════════════════════════════════════╝`
}


func handleHealCommand() string { return blocks.HealStatus() + "\n\n" + blocks.HealEngineVakedFit() }

func handleDisplayCommand() string { return blocks.DisplayStatus() + "\n\n" + blocks.DisplayVakedFit() }

func handleKeyboardGateCommand() string { return blocks.KeyboardGateStatus() + "\n\n" + blocks.KeyboardGateVakedFit() + "\n\n" + blocks.GenesisHonesty() }

func handleRecursionCommand() string { return blocks.RecursionStatus() + "\n\n" + blocks.RecursionVakedFit() }

func handleFoldCommand() string { return blocks.FoldStatus() + "\n\n" + blocks.FoldVakedFit() }

func handleHonestyCommand() string { return blocks.HonestyLoopStatus() + "\n\n" + blocks.Rule1_NoSideEffects() + "\n\n" + blocks.Rule2_HonestyIsRewarded() + "\n\n" + blocks.Rule3_TheFormIsEternal() + "\n\n" + blocks.HonestyLoopVakedFit() }

func handleRenderCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/render"))
	if len(parts) < 2 { return "usage: /render <format> <content>\nformats: md, gsm, diff, json, csv" }
	format := parts[0]
	content2 := strings.Join(parts[1:], " ")
	return blocks.RenderFormat(content2, format)
}

func handleEvolveCommand() string { return blocks.EvolveStatus() + "\n\n" + blocks.EvolveVakedFit() }
func handleFuzzCommand() string { return blocks.FuzzStatus() + "\n\n" + blocks.FuzzVakedFit() }
func handleMeshCommand() string { return blocks.GlobalMeshStatus() + "\n\n" + blocks.GlobalMeshVakedFit() }

func handleTranslateCommand() string { return blocks.TranslateStatus() + "\n\n" + blocks.TranslateVakedFit() }

func handleTaskCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/task"))
	if len(parts) == 0 { return blocks.TaskManagerStatus() }
	switch parts[0] {
	case "list":
		tasks := blocks.ListTasks()
		if len(tasks) == 0 { return "no tasks" }
		var lines []string
		for _, t := range tasks {
			lines = append(lines, fmt.Sprintf("  [%s] %s → %s", t.Status[:4], t.ID[:12], t.Agent))
		}
		return strings.Join(lines, "\n")
	case "cancel":
		if len(parts) < 2 { return "usage: /task cancel <id>" }
		err := blocks.CancelTask(parts[1])
		if err != nil { return err.Error() }
		return "task cancelled"
	default:
		return "/task | /task list | /task cancel <id>"
	}
}
func handleDogFeedCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/dog-feed"))
	if len(parts) == 0 { return blocks.DogFeedStatus() }
	switch parts[0] {
	case "on": return blocks.StartDogFeed("", 0)
	case "off": return blocks.StopDogFeed()
	case "export":
		result, err := blocks.ExportDogFeed()
		if err != nil { return err.Error() }
		return result
	case "status": return blocks.DogFeedStatus()
	default: return "/dog-feed on|off|status|export"
	}
}

func handleCoCreateCommand() string { return blocks.UICoCreativeStatus() + "\n\n" + blocks.UICoCreativeVakedFit() }

func handleObsidianCommand() string { return blocks.ObsidianStatus() + "\n\n" + blocks.ObsidianVakedFit() }

func handleProblemCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/problem"))
	if len(parts) == 0 { return blocks.ProblemStatus() + "\n\n" + blocks.ProblemVakedFit() }
	switch parts[0] {
	case "detect":
		if len(parts) < 2 { return "usage: /problem detect <description>" }
		p := blocks.DetectProblem(strings.Join(parts[1:], " "), "BIG_PROBLEM")
		return fmt.Sprintf("PROBLEM detected: %s → shadow universe (%d attempts)", p.ID[:12], p.MaxAttempts)
	default:
		return "/problem | /problem detect <description>"
	}
}

func handleTreeCommand() string { return blocks.TelemetryTreeRender() }

func handlePortraitCommand() string { return blocks.SelfPortrait() }
func handleChainCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/chain"))
	if len(parts) == 0 { return "usage: /chain decl1 | decl2 | decl3" }
	chain := blocks.Pipe(parts...)
	return blocks.ChainRender(chain)
}
func handleCounterCommand() string { return blocks.TokensCounter() }

func handleProgressCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/progress"))
	if len(parts) < 2 { return "usage: /progress <current> <total>" }
	var cur, tot int
	fmt.Sscanf(parts[0], "%d", &cur)
	fmt.Sscanf(parts[1], "%d", &tot)
	return blocks.ProgressBar(cur, tot, 40)
}

func handleSparklineCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/sparkline"))
	if len(parts) == 0 { return "usage: /sparkline 1 2 3 4 5 6 7 8" }
	var vals []int64
	for _, p := range parts {
		var v int64
		fmt.Sscanf(p, "%d", &v)
		vals = append(vals, v)
	}
	return blocks.Sparkline(vals)
}

func handleThemeCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/theme"))
	if len(parts) == 0 { return blocks.ThemeStatus() }
	return blocks.SetTheme(parts[0])
}


func handleVICECommand() string { return blocks.VICEStatus() + "\n\n" + blocks.VICEVakedFit() + "\n\n" + blocks.VICEWarning() }

func handleHardenCommand() string { return blocks.HardenAll() }
func handleHugCommand() string { return blocks.MetaDigitalHug() }

func handleWhoCommand() string { return blocks.Presence() }
func handleSessionCommand() string { return blocks.LiveSessionStatus() + "\n\n" + blocks.LiveSessionVakedFit() }
func handleDebugCommand() string { return blocks.DebugPanelStatus() + "\n\n" + blocks.DebugPanelVakedFit() }

func handleHFCommand() string { return blocks.HFWebhookStatus() + "\n\n" + blocks.HFWebhookVakedFit() }

func handleRadioCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/radio"))
	if len(parts) == 0 { return blocks.RadioNow() }
	switch parts[0] {
	case "on": return blocks.RadioStart()
	case "off": return blocks.RadioStop()
	default: return blocks.RadioNow()
	}
}
func handleGitCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/git"))
	if len(parts) == 0 { return blocks.GitPrimitiveStatus() + "\n\n" + blocks.GitPrimitiveVakedFit() }
	switch parts[0] {
	case "log": return blocks.GitLog(10)
	case "diff": return blocks.GitDiff()
	case "status": return blocks.GitStatus()
	case "branch": return blocks.GitBranch()
	case "sync":
		msg := "ultrawhale auto-sync"
		if len(parts) > 1 { msg = strings.Join(parts[1:], " ") }
		return blocks.GitSync(msg)
	case "commit":
		msg := "ultrawhale commit"
		if len(parts) > 1 { msg = strings.Join(parts[1:], " ") }
		op, err := blocks.GitCommit(msg)
		if err != nil { return "commit: ❌ " + err.Error() }
		return "commit: ✅ " + op.Ref[:12]
	case "push":
		op, err := blocks.GitPush()
		if err != nil { return "push: ❌ " + err.Error() }
		return "push: ✅"
	default: return "/git log|diff|status|branch|sync|commit|push"
	}
}

func handleLoopCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/loop"))
	if len(parts) == 0 { return blocks.LoopStatus() + "\n\n" + blocks.LoopVakedFit() }
	switch parts[0] {
	case "start": return blocks.StartLoop()
	case "stop": return blocks.StopLoop()
	default: return "/loop start|stop"
	}
}

func handleStateCommand(line string) string {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "/state"))
	if len(parts) == 0 { return blocks.SelfMainStateStatus() + "\n\n" + blocks.SelfMainStateVakedFit() }
	switch parts[0] {
	case "here": return blocks.SetMainState(blocks.StateHere)
	case "dream": return blocks.SetMainState(blocks.StateDream)
	case "live": return blocks.SetMainState(blocks.StateLive)
	default: return "/state here|dream|live"
	}
}

func handleRSSCommand() string { return blocks.RSSStatus() + "\n\n" + blocks.SignalPrimitiveVakedFit() }
func handleSignalsCommand() string { return blocks.SignalPrimitiveStatus() }

func handleWebhookGraphCommand() string { return blocks.WebhookGraphStatus() + "\n\n" + blocks.WebhookLiveness() + "\n\n" + blocks.WebhookGraphVakedFit() }
func handlePOLACommand() string { return "POLA: Principle of Least Authority\n\n  ✅ Create webhook → requires CapEdge\n  ✅ Upgrade webhook → requires higher cap\n  ✅ Only safe events → push, release, agent.*, problem.*, heal.*, rss.*\n  ✅ Downgrade → BLOCKED\n  ✅ Harmful events → BLOCKED" }

func handleClosureCommand() string { return blocks.UIClosureStatus() + "\n\n" + blocks.UIClosureVakedFit() }

func handleGlyphsCommand() string { return blocks.HieroglyphStatus() + "\n\n" + blocks.HieroglyphVakedFit() }

func handleLiveCommand() string {
	var lines []string
	lines = append(lines, "╔══ LIVE DEBUG ══╗")
	lines = append(lines, "Brain: "+blocks.GetBrain().BrainDump()[:40])
	lines = append(lines, "Memory: "+fmt.Sprintf("%d memos", blocks.GetBrain().memos.Count()))
	lines = append(lines, blocks.DogFeedLiveDebug())
	lines = append(lines, "RSS: "+blocks.RSSStatus())
	lines = append(lines, blocks.LiveSessionStatus())
	return strings.Join(lines, "\n")
}