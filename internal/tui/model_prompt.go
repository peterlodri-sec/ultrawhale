package tui

import (
	"fmt"
	"strings"
	"time"
	"github.com/usewhale/whale/internal/blocks"

	tea "github.com/charmbracelet/bubbletea"

	appcommands "github.com/usewhale/whale/internal/runtime/commands"
	"github.com/usewhale/whale/internal/runtime/protocol"
	tuirender "github.com/usewhale/whale/internal/tui/render"
)

func (m *model) startBusy() {
	if m.busySince.IsZero() {
		m.busySince = time.Now()
	}
	m.busy = true
}

func (m *model) stopBusy() {
	m.busy = false
	m.busySince = time.Time{}
	m.resetBusyTokenEstimate()
}

func (m *model) submitPrompt(value string) tea.Cmd {
	return m.submitPromptWithBinding(value, m.currentSkillBinding(value))
}

func (m *model) submitPromptWithBinding(value string, binding *protocol.SkillBinding) tea.Cmd {
	// /reload commands are handled locally, never sent to LLM
	if strings.HasPrefix(strings.TrimSpace(value), "/self") {
	if strings.HasPrefix(strings.TrimSpace(value), "/a2a") {
		m.setEphemeralInfo(handleA2ACommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/a2c") {
		m.setEphemeralInfo(handleA2CCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/probe") {
		m.setEphemeralInfo(handleProbeCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/predict") {
		m.setEphemeralInfo(handlePredictCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/learn") {
		m.setEphemeralInfo(handleLearnCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/brainstorm") {
		m.setEphemeralInfo(handleBrainstormCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/sacred") {
		m.setEphemeralInfo(handleSacredCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked-triangle") {
		m.setEphemeralInfo(handleVakedTriangleCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ui-engine") {
		m.setEphemeralInfo(handleUIEngineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dyad-space") { m.setEphemeralInfo(handleDyadSpaceCommand()); return nil }
	if strings.TrimSpace(value) == "/1min" { m.setEphemeralInfo(handleOneMinCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/ralph-boost") { m.setEphemeralInfo(handleRalphBoostCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/dogfeed-loop") { m.setEphemeralInfo(handleDogFeedLoopCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/disaster") { m.setEphemeralInfo(handleDisasterCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/curator") { m.setEphemeralInfo(handleCuratorCommand()); return nil }
	if strings.TrimSpace(value) == "/entropy-live" { m.setEphemeralInfo(handleEntropyLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/entropy") { m.setEphemeralInfo(handleEntropyCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/promise") { m.setEphemeralInfo(handlePromiseCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/asciibox") { m.setEphemeralInfo(handleASCIIBoxCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/onefold") { m.setEphemeralInfo(handleOneFoldCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/fold3d") { m.setEphemeralInfo(handleFold3DCommand()); return nil }
	if strings.TrimSpace(value) == "/record-pov" { m.setEphemeralInfo(handleRecordPOVCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/proof") { m.setEphemeralInfo(handleProofCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/room") { m.setEphemeralInfo(handleRoomCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/osce") { m.setEphemeralInfo(handleOSCECommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/surface-atoms") { m.setEphemeralInfo(handleSurfaceAtomsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/safespace") { m.setEphemeralInfo(handleSafeSpaceCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/selflive") { m.setEphemeralInfo(handleSelfLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/history") { m.setEphemeralInfo(handleHistoryCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/freemodels") { m.setEphemeralInfo(handleFreeModelsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/live") { m.setEphemeralInfo(handleLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/glyphs") { m.setEphemeralInfo(handleGlyphsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/closure") { m.setEphemeralInfo(handleClosureCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/webhooks") { m.setEphemeralInfo(handleWebhookGraphCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/pola") { m.setEphemeralInfo(handlePOLACommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/rss") { m.setEphemeralInfo(handleRSSCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/signals") { m.setEphemeralInfo(handleSignalsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/state") { m.setEphemeralInfo(handleStateCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/loop") { m.setEphemeralInfo(handleLoopCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/git") { m.setEphemeralInfo(handleGitCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/radio") { m.setEphemeralInfo(handleRadioCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/hf") { m.setEphemeralInfo(handleHFCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/debug") { m.setEphemeralInfo(handleDebugCommand()); return nil }
	if strings.TrimSpace(value) == "/who" { m.setEphemeralInfo(handleWhoCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/session") { m.setEphemeralInfo(handleSessionCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/harden") { m.setEphemeralInfo(handleHardenCommand()); return nil }
	if strings.TrimSpace(value) == "/hug" { m.setEphemeralInfo(handleHugCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/vice") { m.setEphemeralInfo(handleVICECommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/counter") { m.setEphemeralInfo(handleCounterCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/progress") { m.setEphemeralInfo(handleProgressCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/sparkline") { m.setEphemeralInfo(handleSparklineCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/theme") { m.setEphemeralInfo(handleThemeCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/portrait") {
		m.setEphemeralInfo(handlePortraitCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/chain") {
		m.setEphemeralInfo(handleChainCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tree") {
		m.setEphemeralInfo(handleTreeCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/problem") {
		m.setEphemeralInfo(handleProblemCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/obsidian") {
		m.setEphemeralInfo(handleObsidianCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/co-create") {
		m.setEphemeralInfo(handleCoCreateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dog-feed") {
		m.setEphemeralInfo(handleDogFeedCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/task") {
		m.setEphemeralInfo(handleTaskCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/translate") {
		m.setEphemeralInfo(handleTranslateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/fuzz") {
		m.setEphemeralInfo(handleFuzzCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/mesh") {
		m.setEphemeralInfo(handleMeshCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/evolve") {
		m.setEphemeralInfo(handleEvolveCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/render") {
		m.setEphemeralInfo(handleRenderCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/honesty") {
		m.setEphemeralInfo(handleHonestyCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/fold") {
		m.setEphemeralInfo(handleFoldCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/recursion") {
		m.setEphemeralInfo(handleRecursionCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/keyboard-gate") {
		m.setEphemeralInfo(handleKeyboardGateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/display") {
		m.setEphemeralInfo(handleDisplayCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/heal") {
		m.setEphemeralInfo(handleHealCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked-pipeline") {
		m.setEphemeralInfo(handleVakedPipelineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/engine") {
		m.setEphemeralInfo(handleEngineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/space") {
		m.setEphemeralInfo(handleSpaceCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ui") {
		m.setEphemeralInfo(handleUICommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/mesh") {
		m.setEphemeralInfo(handleMeshCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/caps") {
		m.setEphemeralInfo(handleCapsCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dyad") {
		m.setEphemeralInfo(handleDyadCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/metal") {
		m.setEphemeralInfo(handleMetalCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked") {
		m.setEphemeralInfo(handleVakedCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/orch-tools") {
		m.setEphemeralInfo(handleOrchToolsCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ssh") {
		m.setEphemeralInfo(handleSSHCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/run") {
		m.setEphemeralInfo(handleRunCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ralph") {
		m.setEphemeralInfo(handleRalphCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tools") {
		m.setEphemeralInfo(handleToolsCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tool-cache") {
		m.setEphemeralInfo(handleToolCacheCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/deploy") {
		m.setEphemeralInfo(handleDeployCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/memo") {
		m.setEphemeralInfo(handleMemoCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/current") {
		m.setEphemeralInfo(handleCurrentCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
		m.setEphemeralInfo(handleSelfCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ultracode") {
		handled, msg := m.handleUltracodeCommand(value)
		if handled {
			m.setEphemeralInfo(msg)
			m.input.SetValue("")
			m.refreshViewportContent()
			return nil
		}
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/self") {
	if strings.HasPrefix(strings.TrimSpace(value), "/a2a") {
		m.setEphemeralInfo(handleA2ACommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/a2c") {
		m.setEphemeralInfo(handleA2CCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/probe") {
		m.setEphemeralInfo(handleProbeCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/predict") {
		m.setEphemeralInfo(handlePredictCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/learn") {
		m.setEphemeralInfo(handleLearnCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/brainstorm") {
		m.setEphemeralInfo(handleBrainstormCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/sacred") {
		m.setEphemeralInfo(handleSacredCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked-triangle") {
		m.setEphemeralInfo(handleVakedTriangleCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ui-engine") {
		m.setEphemeralInfo(handleUIEngineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dyad-space") { m.setEphemeralInfo(handleDyadSpaceCommand()); return nil }
	if strings.TrimSpace(value) == "/1min" { m.setEphemeralInfo(handleOneMinCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/ralph-boost") { m.setEphemeralInfo(handleRalphBoostCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/dogfeed-loop") { m.setEphemeralInfo(handleDogFeedLoopCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/disaster") { m.setEphemeralInfo(handleDisasterCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/curator") { m.setEphemeralInfo(handleCuratorCommand()); return nil }
	if strings.TrimSpace(value) == "/entropy-live" { m.setEphemeralInfo(handleEntropyLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/entropy") { m.setEphemeralInfo(handleEntropyCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/promise") { m.setEphemeralInfo(handlePromiseCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/asciibox") { m.setEphemeralInfo(handleASCIIBoxCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/onefold") { m.setEphemeralInfo(handleOneFoldCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/fold3d") { m.setEphemeralInfo(handleFold3DCommand()); return nil }
	if strings.TrimSpace(value) == "/record-pov" { m.setEphemeralInfo(handleRecordPOVCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/proof") { m.setEphemeralInfo(handleProofCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/room") { m.setEphemeralInfo(handleRoomCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/osce") { m.setEphemeralInfo(handleOSCECommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/surface-atoms") { m.setEphemeralInfo(handleSurfaceAtomsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/safespace") { m.setEphemeralInfo(handleSafeSpaceCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/selflive") { m.setEphemeralInfo(handleSelfLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/history") { m.setEphemeralInfo(handleHistoryCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/freemodels") { m.setEphemeralInfo(handleFreeModelsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/live") { m.setEphemeralInfo(handleLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/glyphs") { m.setEphemeralInfo(handleGlyphsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/closure") { m.setEphemeralInfo(handleClosureCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/webhooks") { m.setEphemeralInfo(handleWebhookGraphCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/pola") { m.setEphemeralInfo(handlePOLACommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/rss") { m.setEphemeralInfo(handleRSSCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/signals") { m.setEphemeralInfo(handleSignalsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/state") { m.setEphemeralInfo(handleStateCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/loop") { m.setEphemeralInfo(handleLoopCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/git") { m.setEphemeralInfo(handleGitCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/radio") { m.setEphemeralInfo(handleRadioCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/hf") { m.setEphemeralInfo(handleHFCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/debug") { m.setEphemeralInfo(handleDebugCommand()); return nil }
	if strings.TrimSpace(value) == "/who" { m.setEphemeralInfo(handleWhoCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/session") { m.setEphemeralInfo(handleSessionCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/harden") { m.setEphemeralInfo(handleHardenCommand()); return nil }
	if strings.TrimSpace(value) == "/hug" { m.setEphemeralInfo(handleHugCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/vice") { m.setEphemeralInfo(handleVICECommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/counter") { m.setEphemeralInfo(handleCounterCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/progress") { m.setEphemeralInfo(handleProgressCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/sparkline") { m.setEphemeralInfo(handleSparklineCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/theme") { m.setEphemeralInfo(handleThemeCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/portrait") {
		m.setEphemeralInfo(handlePortraitCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/chain") {
		m.setEphemeralInfo(handleChainCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tree") {
		m.setEphemeralInfo(handleTreeCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/problem") {
		m.setEphemeralInfo(handleProblemCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/obsidian") {
		m.setEphemeralInfo(handleObsidianCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/co-create") {
		m.setEphemeralInfo(handleCoCreateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dog-feed") {
		m.setEphemeralInfo(handleDogFeedCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/task") {
		m.setEphemeralInfo(handleTaskCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/translate") {
		m.setEphemeralInfo(handleTranslateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/fuzz") {
		m.setEphemeralInfo(handleFuzzCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/mesh") {
		m.setEphemeralInfo(handleMeshCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/evolve") {
		m.setEphemeralInfo(handleEvolveCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/render") {
		m.setEphemeralInfo(handleRenderCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/honesty") {
		m.setEphemeralInfo(handleHonestyCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/fold") {
		m.setEphemeralInfo(handleFoldCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/recursion") {
		m.setEphemeralInfo(handleRecursionCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/keyboard-gate") {
		m.setEphemeralInfo(handleKeyboardGateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/display") {
		m.setEphemeralInfo(handleDisplayCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/heal") {
		m.setEphemeralInfo(handleHealCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked-pipeline") {
		m.setEphemeralInfo(handleVakedPipelineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/engine") {
		m.setEphemeralInfo(handleEngineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/space") {
		m.setEphemeralInfo(handleSpaceCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ui") {
		m.setEphemeralInfo(handleUICommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/mesh") {
		m.setEphemeralInfo(handleMeshCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/caps") {
		m.setEphemeralInfo(handleCapsCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dyad") {
		m.setEphemeralInfo(handleDyadCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/metal") {
		m.setEphemeralInfo(handleMetalCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked") {
		m.setEphemeralInfo(handleVakedCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/orch-tools") {
		m.setEphemeralInfo(handleOrchToolsCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ssh") {
		m.setEphemeralInfo(handleSSHCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/run") {
		m.setEphemeralInfo(handleRunCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ralph") {
		m.setEphemeralInfo(handleRalphCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tools") {
		m.setEphemeralInfo(handleToolsCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tool-cache") {
		m.setEphemeralInfo(handleToolCacheCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/deploy") {
		m.setEphemeralInfo(handleDeployCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/memo") {
		m.setEphemeralInfo(handleMemoCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/current") {
		m.setEphemeralInfo(handleCurrentCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
		m.setEphemeralInfo(handleSelfCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ultracode") {
		handled, msg := m.handleUltracodeCommand(value)
		if handled {
			m.setEphemeralInfo(msg)
			m.input.SetValue("")
			m.refreshViewportContent()
			return nil
		}
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/reload") {
		handled, msg := m.handleReloadCommand(value)
		if handled {
			m.setEphemeralInfo(msg)
			m.input.SetValue("")
			m.refreshViewportContent()
			return nil
		}
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	submit := m.classifySubmit(value)
	if submit.LocalNoTurn() {
		return m.submitLocalNoTurn(submit)
	}
	return m.submitPromptTurn(value, binding, attachmentInputsFromComposerAttachments(m.consumeVisibleComposerAttachments(value)))
}

func (m *model) submitPromptWithBindingAndAttachments(value string, binding *protocol.SkillBinding, attachments []protocol.AttachmentInput) tea.Cmd {
	// /reload commands are handled locally, never sent to LLM
	if strings.HasPrefix(strings.TrimSpace(value), "/self") {
	if strings.HasPrefix(strings.TrimSpace(value), "/a2a") {
		m.setEphemeralInfo(handleA2ACommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/a2c") {
		m.setEphemeralInfo(handleA2CCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/probe") {
		m.setEphemeralInfo(handleProbeCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/predict") {
		m.setEphemeralInfo(handlePredictCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/learn") {
		m.setEphemeralInfo(handleLearnCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/brainstorm") {
		m.setEphemeralInfo(handleBrainstormCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/sacred") {
		m.setEphemeralInfo(handleSacredCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked-triangle") {
		m.setEphemeralInfo(handleVakedTriangleCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ui-engine") {
		m.setEphemeralInfo(handleUIEngineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dyad-space") { m.setEphemeralInfo(handleDyadSpaceCommand()); return nil }
	if strings.TrimSpace(value) == "/1min" { m.setEphemeralInfo(handleOneMinCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/ralph-boost") { m.setEphemeralInfo(handleRalphBoostCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/dogfeed-loop") { m.setEphemeralInfo(handleDogFeedLoopCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/disaster") { m.setEphemeralInfo(handleDisasterCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/curator") { m.setEphemeralInfo(handleCuratorCommand()); return nil }
	if strings.TrimSpace(value) == "/entropy-live" { m.setEphemeralInfo(handleEntropyLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/entropy") { m.setEphemeralInfo(handleEntropyCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/promise") { m.setEphemeralInfo(handlePromiseCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/asciibox") { m.setEphemeralInfo(handleASCIIBoxCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/onefold") { m.setEphemeralInfo(handleOneFoldCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/fold3d") { m.setEphemeralInfo(handleFold3DCommand()); return nil }
	if strings.TrimSpace(value) == "/record-pov" { m.setEphemeralInfo(handleRecordPOVCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/proof") { m.setEphemeralInfo(handleProofCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/room") { m.setEphemeralInfo(handleRoomCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/osce") { m.setEphemeralInfo(handleOSCECommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/surface-atoms") { m.setEphemeralInfo(handleSurfaceAtomsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/safespace") { m.setEphemeralInfo(handleSafeSpaceCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/selflive") { m.setEphemeralInfo(handleSelfLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/history") { m.setEphemeralInfo(handleHistoryCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/freemodels") { m.setEphemeralInfo(handleFreeModelsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/live") { m.setEphemeralInfo(handleLiveCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/glyphs") { m.setEphemeralInfo(handleGlyphsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/closure") { m.setEphemeralInfo(handleClosureCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/webhooks") { m.setEphemeralInfo(handleWebhookGraphCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/pola") { m.setEphemeralInfo(handlePOLACommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/rss") { m.setEphemeralInfo(handleRSSCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/signals") { m.setEphemeralInfo(handleSignalsCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/state") { m.setEphemeralInfo(handleStateCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/loop") { m.setEphemeralInfo(handleLoopCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/git") { m.setEphemeralInfo(handleGitCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/radio") { m.setEphemeralInfo(handleRadioCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/hf") { m.setEphemeralInfo(handleHFCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/debug") { m.setEphemeralInfo(handleDebugCommand()); return nil }
	if strings.TrimSpace(value) == "/who" { m.setEphemeralInfo(handleWhoCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/session") { m.setEphemeralInfo(handleSessionCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/harden") { m.setEphemeralInfo(handleHardenCommand()); return nil }
	if strings.TrimSpace(value) == "/hug" { m.setEphemeralInfo(handleHugCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/vice") { m.setEphemeralInfo(handleVICECommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/counter") { m.setEphemeralInfo(handleCounterCommand()); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/progress") { m.setEphemeralInfo(handleProgressCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/sparkline") { m.setEphemeralInfo(handleSparklineCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/theme") { m.setEphemeralInfo(handleThemeCommand(value)); return nil }
	if strings.HasPrefix(strings.TrimSpace(value), "/portrait") {
		m.setEphemeralInfo(handlePortraitCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/chain") {
		m.setEphemeralInfo(handleChainCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tree") {
		m.setEphemeralInfo(handleTreeCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/problem") {
		m.setEphemeralInfo(handleProblemCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/obsidian") {
		m.setEphemeralInfo(handleObsidianCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/co-create") {
		m.setEphemeralInfo(handleCoCreateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dog-feed") {
		m.setEphemeralInfo(handleDogFeedCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/task") {
		m.setEphemeralInfo(handleTaskCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/translate") {
		m.setEphemeralInfo(handleTranslateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/fuzz") {
		m.setEphemeralInfo(handleFuzzCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/mesh") {
		m.setEphemeralInfo(handleMeshCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/evolve") {
		m.setEphemeralInfo(handleEvolveCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/render") {
		m.setEphemeralInfo(handleRenderCommand(value))
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/honesty") {
		m.setEphemeralInfo(handleHonestyCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/fold") {
		m.setEphemeralInfo(handleFoldCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/recursion") {
		m.setEphemeralInfo(handleRecursionCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/keyboard-gate") {
		m.setEphemeralInfo(handleKeyboardGateCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/display") {
		m.setEphemeralInfo(handleDisplayCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/heal") {
		m.setEphemeralInfo(handleHealCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked-pipeline") {
		m.setEphemeralInfo(handleVakedPipelineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/engine") {
		m.setEphemeralInfo(handleEngineCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/space") {
		m.setEphemeralInfo(handleSpaceCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ui") {
		m.setEphemeralInfo(handleUICommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/mesh") {
		m.setEphemeralInfo(handleMeshCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/caps") {
		m.setEphemeralInfo(handleCapsCommand())
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/dyad") {
		m.setEphemeralInfo(handleDyadCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/metal") {
		m.setEphemeralInfo(handleMetalCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/vaked") {
		m.setEphemeralInfo(handleVakedCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/orch-tools") {
		m.setEphemeralInfo(handleOrchToolsCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ssh") {
		m.setEphemeralInfo(handleSSHCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/run") {
		m.setEphemeralInfo(handleRunCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ralph") {
		m.setEphemeralInfo(handleRalphCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tools") {
		m.setEphemeralInfo(handleToolsCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/tool-cache") {
		m.setEphemeralInfo(handleToolCacheCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/deploy") {
		m.setEphemeralInfo(handleDeployCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/memo") {
		m.setEphemeralInfo(handleMemoCommand(value))
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/current") {
		m.setEphemeralInfo(handleCurrentCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
		m.setEphemeralInfo(handleSelfCommand())
		m.input.SetValue("")
		m.refreshViewportContent()
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/ultracode") {
		handled, msg := m.handleUltracodeCommand(value)
		if handled {
			m.setEphemeralInfo(msg)
			m.input.SetValue("")
			m.refreshViewportContent()
			return nil
		}
	}
	if strings.HasPrefix(strings.TrimSpace(value), "/reload") {
		handled, msg := m.handleReloadCommand(value)
		if handled {
			m.setEphemeralInfo(msg)
			m.input.SetValue("")
			m.refreshViewportContent()
			return nil
		}
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	submit := m.classifySubmit(value)
	if submit.LocalNoTurn() {
		return m.submitLocalNoTurn(submit)
	}
	return m.submitPromptTurn(value, binding, cloneAttachmentInputs(attachments))
}

func (m *model) submitPromptTurn(value string, binding *protocol.SkillBinding, attachments []protocol.AttachmentInput) tea.Cmd {
	// Init orchestrator on first prompt
	if blocks.GetOrchestrator().TotalTurns == 0 {
		blocks.InitOrchestrator(m.sessionID)
	}
	m.clearEphemeralMessages()

	// KEYBOARD GATE: submit through one-way gate
	value = blocks.Submit()
	if value == "" { return nil }

	// SACRED: user input is direct, unmodified, 1:1
	blocks.MarkInput()
	if v := blocks.ViolateSacred("wrap"); v != "" && false {
		// Swarm wrapping is allowed — still direct, just annotated
	}
	// Self resolution: inject identity when user says "you" or "deepseek"
	if selfIntro, ok := blocks.ResolveSelfPrompt(value); ok {
		value = value + "\n\n[Identity: " + selfIntro + "]"
	}
	if m.assembler != nil && m.assembler.Len() > 0 {
		m.commitLiveTranscript(false)
	}
	m.recordPromptHistory(value)
	m.resetHistoryNavigation()
	m.appendTranscript("you", tuirender.KindText, visibleSubmittedText(value))
	m.beginTurnTranscript()
	m.input.SetValue("")
	m.skillBinding = nil
	m.resetWindowsPasteFallbackInputState()
	m.slash.matches = nil
	m.slash.selected = 0
	m.slash.argumentHint = ""
	clearFileSuggestions(m)
	m.startBusy()
	m.status = "running"
	m.dispatchIntent(protocol.Intent{Kind: protocol.IntentSubmit, Input: value, SkillBinding: binding, Attachments: attachments})
	m.refreshViewportContentFollow(true)
	return busyTickCmd()
}

func (m *model) submitPromptWhileBusy(value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		if m.stopping {
			m.status = "stopping"
		}
		return
	}
	submit := m.classifySubmit(value)
	if submit.BusyImmediate() {
		_ = m.submitLocalNoTurn(submit)
		return
	}
	if appcommands.LooksLikeSlashCommand(submit.Line) {
		m.status = busySlashBlockedStatus(submit.Line, m.stopping)
		m.refreshViewportContent()
		return
	}
	m.enqueuePrompt(value)
}

func (m *model) submitPromptFromDeferredBusyEnter(value string, wasStopping bool) tea.Cmd {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if wasStopping && !m.busy {
		return nil
	}
	submit := m.classifySubmit(value)
	if submit.BusyImmediate() {
		_ = m.submitLocalNoTurn(submit)
		return nil
	}
	stopping := m.stopping || wasStopping
	if appcommands.LooksLikeSlashCommand(submit.Line) {
		m.status = busySlashBlockedStatus(submit.Line, stopping)
		m.refreshViewportContent()
		return nil
	}
	m.enqueuePrompt(value)
	return nil
}

func busySlashBlockedStatus(line string, stopping bool) string {
	fields := strings.Fields(line)
	cmd := strings.TrimSpace(line)
	if len(fields) > 0 {
		cmd = fields[0]
	}
	state := "working"
	if stopping {
		state = "stopping"
	}
	return fmt.Sprintf("%s disabled while %s", cmd, state)
}

func (m model) classifySubmit(value string) appcommands.SubmitClassification {
	line := m.expandSubmitSlashPrefix(value)
	submit := appcommands.ClassifySubmit(line, appcommands.CommandsHelp(), "/mcp")
	if submit.Class != appcommands.SubmitUsageError {
		return submit
	}
	if class, ok := m.pluginSubmitClass(submit.Line); ok {
		return appcommands.SubmitClassification{Line: submit.Line, Class: class}
	}
	return submit
}

func (m model) expandSubmitSlashPrefix(value string) string {
	line := strings.TrimSpace(value)
	if !appcommands.LooksLikeSlashCommand(line) || strings.ContainsAny(line, " \t") {
		return line
	}
	var names []string
	for _, spec := range m.slash.all {
		name := strings.TrimSpace(spec.Name)
		if name != "" {
			names = append(names, name)
		}
	}
	matches := make([]string, 0, 1)
	for _, name := range names {
		if strings.HasPrefix(name, line) {
			matches = append(matches, name)
		}
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return line
}

func (m model) pluginSubmitClass(line string) (appcommands.SubmitClass, bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 || m.slash.commandClasses == nil {
		return appcommands.SubmitUsageError, false
	}
	class, ok := m.slash.commandClasses[fields[0]]
	return class, ok
}

func (m *model) submitLocalNoTurn(submit appcommands.SubmitClassification) tea.Cmd {
	cmd := submit.Line
	if strings.TrimSpace(cmd) == "/help" {
		m.openHelp()
		return nil
	}
	if m.btwPanel.loading && isBtwCommand(cmd) {
		m.status = "/btw is already answering"
		return nil
	}
	m.clearEphemeralMessages()

	// KEYBOARD GATE: submit through one-way gate
	value = blocks.Submit()
	if value == "" { return nil }

	// SACRED: user input is direct, unmodified, 1:1
	blocks.MarkInput()
	if v := blocks.ViolateSacred("wrap"); v != "" && false {
		// Swarm wrapping is allowed — still direct, just annotated
	}
	m.recordPromptHistory(cmd)
	m.resetHistoryNavigation()
	m.input.SetValue("")
	m.skillBinding = nil
	m.resetWindowsPasteFallbackInputState()
	m.slash.matches = nil
	m.slash.selected = 0
	m.slash.argumentHint = ""
	m.localSubmitPending++
	m.localSubmitCommands = append(m.localSubmitCommands, cmd)
	if !m.busy || submit.SubmitBarrier() {
		m.status = "command pending"
	}
	if appcommands.IsOpenCommandLine(cmd) {
		return m.startOpenCommand(cmd)
	}
	m.dispatchIntent(protocol.Intent{Kind: protocol.IntentSubmitLocal, Input: cmd})
	m.refreshViewportContent()
	return nil
}

func (m *model) submitSteeringPrompt(value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	binding := m.currentSkillBinding(value)
	clientInputID := m.nextPendingInputID()
	m.clearEphemeralMessages()

	// KEYBOARD GATE: submit through one-way gate
	value = blocks.Submit()
	if value == "" { return nil }

	// SACRED: user input is direct, unmodified, 1:1
	blocks.MarkInput()
	if v := blocks.ViolateSacred("wrap"); v != "" && false {
		// Swarm wrapping is allowed — still direct, just annotated
	}
	if selfIntro, ok := blocks.ResolveSelfPrompt(value); ok {
		value = value + "\n\n[Identity: " + selfIntro + "]"
	}
	if m.assembler != nil && m.assembler.Len() > 0 {
		m.commitLiveTranscript(false)
	}
	m.recordPromptHistory(value)
	m.resetHistoryNavigation()
	m.appendTranscript("you", tuirender.KindText, visibleSubmittedText(value))
	m.input.SetValue("")
	m.skillBinding = nil
	m.resetWindowsPasteFallbackInputState()
	m.slash.matches = nil
	m.slash.selected = 0
	m.slash.argumentHint = ""
	clearFileSuggestions(m)
	m.pendingSteers = append(m.pendingSteers, pendingSteer{
		ID:           clientInputID,
		Text:         value,
		SkillBinding: binding,
	})
	m.status = "sent"
	m.dispatchIntent(protocol.Intent{Kind: protocol.IntentSubmit, Input: value, ClientInputID: clientInputID, SkillBinding: binding})
	m.refreshViewportContentFollow(true)
}

func (m *model) nextPendingInputID() string {
	m.nextClientInputID++
	return fmt.Sprintf("pending-%d", m.nextClientInputID)
}

func (m *model) markPendingInputAccepted(clientInputID string) {
	if clientInputID == "" {
		return
	}
	for i := range m.pendingSteers {
		if m.pendingSteers[i].ID == clientInputID {
			m.pendingSteers[i].Accepted = true
			m.refreshViewportContent()
			return
		}
	}
}

func (m *model) rejectPendingInput(clientInputID, text string) tea.Cmd {
	if clientInputID == "" {
		return nil
	}
	for i, steer := range m.pendingSteers {
		if steer.ID != clientInputID {
			continue
		}
		if strings.TrimSpace(text) == "" {
			text = steer.Text
		}
		m.pendingSteers = append(m.pendingSteers[:i], m.pendingSteers[i+1:]...)
		return m.restoreTextToComposer(text)
	}
	return nil
}

func (m *model) clearAcceptedPendingSteers() {
	if len(m.pendingSteers) == 0 {
		return
	}
	m.pendingSteers = nil
}

func (m *model) restoreTextToComposer(text string) tea.Cmd {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	current := strings.TrimSpace(m.input.Value())
	if current != "" {
		text += "\n" + current
	}
	m.input.SetValue(text)
	m.skillBinding = nil
	m.resetHistoryNavigation()
	cmd := m.updateSlashMatches()
	m.refreshViewportContent()
	return cmd
}

func (m *model) submitPendingSteersNow() tea.Cmd {
	if len(m.pendingSteers) == 0 {
		return nil
	}
	parts := make([]string, 0, len(m.pendingSteers))
	for _, steer := range m.pendingSteers {
		if text := strings.TrimSpace(steer.Text); text != "" {
			parts = append(parts, text)
		}
	}
	m.pendingSteers = nil
	value := strings.TrimSpace(strings.Join(parts, "\n"))
	if value == "" {
		return nil
	}
	m.startBusy()
	m.status = "running"
	m.clearEphemeralMessages()

	// KEYBOARD GATE: submit through one-way gate
	value = blocks.Submit()
	if value == "" { return nil }

	// SACRED: user input is direct, unmodified, 1:1
	blocks.MarkInput()
	if v := blocks.ViolateSacred("wrap"); v != "" && false {
		// Swarm wrapping is allowed — still direct, just annotated
	}
	m.beginTurnTranscript()
	m.dispatchIntent(protocol.Intent{Kind: protocol.IntentSubmit, Input: value})
	m.refreshViewportContentFollow(true)
	return busyTickCmd()
}

func (m *model) prepareQueuedPromptAfterInterrupt() {
	value := strings.TrimSpace(m.input.Value())
	if value != "" {
		submit := m.classifySubmit(value)
		if !submit.BusyImmediate() && !appcommands.LooksLikeSlashCommand(submit.Line) {
			m.enqueuePrompt(value)
		}
	}
	if len(m.queuedPrompts) > 0 {
		m.submitQueuedPromptAfterInterrupt = true
	}
}

func (m *model) submitQueuedPromptAfterInterruptNow(snapshot windowsBusyInputSnapshot) tea.Cmd {
	next, ok := m.popQueuedPrompt()
	m.submitQueuedPromptAfterInterrupt = false
	if !ok {
		return nil
	}
	return tea.Batch(m.submitPromptWithBinding(next.Text, next.SkillBinding), m.restoreWindowsBusyInput(snapshot))
}

func isBtwCommand(line string) bool {
	fields := strings.Fields(strings.TrimSpace(line))
	return len(fields) > 0 && fields[0] == "/btw"
}

func (m *model) enqueuePrompt(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	m.queuedPrompts = append(m.queuedPrompts, queuedPrompt{Text: value, SkillBinding: m.currentSkillBinding(value), Attachments: m.consumeVisibleComposerAttachments(value)})
	m.input.SetValue("")
	m.skillBinding = nil
	m.resetWindowsPasteFallbackInputState()
	m.resetHistoryNavigation()
	m.slash.matches = nil
	m.slash.selected = 0
	m.slash.argumentHint = ""
	m.status = fmt.Sprintf("queued (%d)", len(m.queuedPrompts))
	m.refreshViewportContent()
	return true
}

func (m *model) popQueuedPrompt() (queuedPrompt, bool) {
	if len(m.queuedPrompts) == 0 {
		return queuedPrompt{}, false
	}
	next := m.queuedPrompts[0]
	copy(m.queuedPrompts, m.queuedPrompts[1:])
	m.queuedPrompts = m.queuedPrompts[:len(m.queuedPrompts)-1]
	return next, true
}

func (m *model) restoreQueuedPromptsToComposer() (bool, tea.Cmd) {
	return m.restoreQueuedPromptsToComposerWithCurrent(m.input.Value())
}

func (m *model) restoreQueuedPromptsToComposerWithWindowsInput(snapshot windowsBusyInputSnapshot) (bool, tea.Cmd) {
	current := m.input.Value()
	if snapshot.ok {
		current = snapshot.composerValue()
	}
	restored, cmd := m.restoreQueuedPromptsToComposerWithCurrent(current)
	if restored && snapshot.ok {
		m.resetWindowsPasteFallbackInputState()
	}
	return restored, cmd
}

func (m *model) restoreQueuedPromptsToComposerWithCurrent(currentValue string) (bool, tea.Cmd) {
	if len(m.queuedPrompts) == 0 && len(m.pendingSteers) == 0 {
		return false, nil
	}
	parts := make([]string, 0, len(m.pendingSteers)+len(m.queuedPrompts)+1)
	for _, steer := range m.pendingSteers {
		if text := strings.TrimSpace(steer.Text); text != "" {
			parts = append(parts, text)
		}
	}
	for _, prompt := range m.queuedPrompts {
		if text := strings.TrimSpace(prompt.Text); text != "" {
			parts = append(parts, text)
		}
		m.composerAttachments = append(m.composerAttachments, cloneComposerAttachments(prompt.Attachments)...)
	}
	if current := strings.TrimSpace(currentValue); current != "" {
		parts = append(parts, current)
	}
	m.pendingSteers = nil
	m.submitQueuedPromptAfterInterrupt = false
	m.queuedPrompts = nil
	m.skillBinding = nil
	m.input.SetValue(strings.Join(parts, "\n"))
	m.resetHistoryNavigation()
	cmd := m.updateSlashMatches()
	m.refreshViewportContent()
	return true, cmd
}

type windowsBusyInputSnapshot struct {
	ok           bool
	value        string
	skillBinding *protocol.SkillBinding
	windowsPaste windowsPasteFallbackState
}

func (m model) snapshotWindowsBusyInput() windowsBusyInputSnapshot {
	if !m.hasPendingWindowsBusyInput() {
		return windowsBusyInputSnapshot{}
	}
	return windowsBusyInputSnapshot{
		ok:           true,
		value:        m.input.Value(),
		skillBinding: m.skillBinding,
		windowsPaste: m.windowsPaste,
	}
}

func (s windowsBusyInputSnapshot) composerValue() string {
	if !s.ok {
		return ""
	}
	if s.windowsPaste.bufferLen == 0 {
		return s.value
	}
	return s.value + model{windowsPaste: s.windowsPaste}.windowsPasteBuffer()
}

func (m *model) restoreWindowsBusyInput(snapshot windowsBusyInputSnapshot) tea.Cmd {
	if !snapshot.ok {
		return nil
	}
	m.input.SetValue(snapshot.value)
	m.skillBinding = snapshot.skillBinding
	m.windowsPaste = snapshot.windowsPaste
	return m.updateSlashMatches()
}
