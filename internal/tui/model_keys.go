package tui

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/usewhale/whale/internal/app/service"
	"github.com/usewhale/whale/internal/core"
	tuirender "github.com/usewhale/whale/internal/tui/render"
)

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Cmd, bool, bool) {
	if !m.quitArmedUntil.IsZero() && time.Now().After(m.quitArmedUntil) {
		m.quitArmedUntil = time.Time{}
		if m.status == "Press Ctrl+C again to quit" {
			m.status = "ready"
		}
	}
	m.updateSlashMatches()
	if m.mode == modeChat && msg.Paste {
		m.input.HandlePaste(string(msg.Runes))
		m.resetHistoryNavigation()
		m.updateSlashMatches()
		return nil, false, true
	}
	if msg.String() == "ctrl+c" && m.busy {
		return m.interruptBusyTurn(), false, true
	}
	if m.mode == modeChat {
		if cmd, handled := m.handleChatModeKey(msg); handled {
			return cmd, false, true
		}
	}
	switch m.mode {
	case modeApproval:
		return m.handleApprovalKey(msg), false, true
	case modeSessionPicker:
		return m.handleSessionPickerKey(msg), false, true
	case modeUserInput:
		return m.handleUserInputKey(msg), false, true
	case modeModelPicker:
		return m.handleModelPickerKey(msg), false, true
	case modePermissionsPicker:
		return m.handlePermissionsPickerKey(msg), false, true
	case modePlanImplementation:
		return m.handlePlanImplementationKey(msg), false, true
	case modeSkillsMenu:
		return m.handleSkillsMenuKey(msg), false, true
	case modeSkillsManager:
		return m.handleSkillsManagerKey(msg), false, true
	}
	cmd, quit, handled := m.handleGlobalKey(msg)
	if handled {
		return cmd, quit, true
	}
	cmd, handled = m.handleComposerKey(msg)
	return cmd, false, handled
}

func (m *model) interruptBusyTurn() tea.Cmd {
	m.quitArmedUntil = time.Time{}
	alreadyStopping := m.stopping
	m.cancelBlockingModalForInterrupt(!alreadyStopping)
	if !alreadyStopping {
		if m.svc != nil {
			m.dispatchIntent(service.Intent{Kind: service.IntentShutdown})
		}
		m.status = "stopping"
		m.stopping = true
		m.appendNotice(m.turnInterruptedNoticeText())
	}
	m.commitLiveTranscript(false)
	return m.flushNativeScrollbackCmd()
}

func (m *model) cancelBlockingModalForInterrupt(dispatch bool) {
	switch m.mode {
	case modeApproval:
		if dispatch && m.approval.toolCallID != "" {
			m.dispatchIntent(service.Intent{Kind: service.IntentCancelToolApproval, ToolCallID: m.approval.toolCallID})
		}
		m.mode = modeChat
	case modeUserInput:
		if dispatch && m.userInput.toolCallID != "" {
			m.dispatchIntent(service.Intent{Kind: service.IntentCancelUserInput, ToolCallID: m.userInput.toolCallID})
		}
		m.mode = modeChat
	}
}

func (m *model) handleChatModeKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "shift+tab", "backtab":
		if m.localSubmitPending > 0 {
			m.status = "wait for command to finish"
			m.refreshViewportContent()
			return m.flushNativeScrollbackCmd(), true
		}
		if !m.busy && !m.hasSlashSuggestions() && !m.hasSkillSuggestions() {
			m.dispatchIntent(service.Intent{Kind: service.IntentToggleMode})
			return nil, true
		}
	case "up":
		if m.hasSlashSuggestions() {
			if m.slash.selected > 0 {
				m.slash.selected--
			}
			return nil, true
		}
		if m.hasSkillSuggestions() {
			if m.skills.selected > 0 {
				m.skills.selected--
			}
			return nil, true
		}
		if m.shouldHandleHistoryNavigation() && m.historyPrev() {
			return nil, true
		}
	case "down":
		if m.hasSlashSuggestions() {
			if m.slash.selected < len(m.slash.matches)-1 {
				m.slash.selected++
			}
			return nil, true
		}
		if m.hasSkillSuggestions() {
			if m.skills.selected < len(m.skills.matches)-1 {
				m.skills.selected++
			}
			return nil, true
		}
		if m.shouldHandleHistoryNavigation() && m.historyNext() {
			return nil, true
		}
	case "tab":
		if m.hasSlashSuggestions() {
			if cmd := safeChoice(m.slash.matches, m.slash.selected); cmd != "" {
				m.input.SetValue(cmd)
				m.skillBinding = nil
				m.updateSlashMatches()
			}
			return nil, true
		}
		if m.insertSelectedSkill() {
			return nil, true
		}
	case "ctrl+c":
		if m.busy {
			return m.interruptBusyTurn(), true
		}
	case "esc":
		if m.busy {
			return m.interruptBusyTurn(), true
		}
		if m.hasSlashSuggestions() {
			m.slash.matches = nil
			m.slash.selected = 0
			return nil, true
		}
		if m.hasSkillSuggestions() {
			m.skills.matches = nil
			m.skills.selected = 0
			return nil, true
		}
	case "pgup", "pgdown", "ctrl+d", "home", "end":
		return m.handleViewportScrollKey(msg.String()), true
	}
	return nil, false
}

func (m *model) handleApprovalKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "left", "h":
		m.approval.selected = (m.approval.selected + 2) % 3
		return nil
	case "right", "l", "tab":
		m.approval.selected = (m.approval.selected + 1) % 3
		return nil
	case "enter":
		switch m.approval.selected {
		case 0:
			return m.submitApprovalDecision(service.IntentAllowTool, "approval_allow", "allow", "approved", "allow")
		case 1:
			return m.submitApprovalDecision(service.IntentAllowToolForSession, "approval_allow_session", "allow for session", "approved for session", "allow_session")
		default:
			return m.submitApprovalDecision(service.IntentDenyTool, "approval_deny", "deny", "rejected", "deny")
		}
	case "a":
		return m.submitApprovalDecision(service.IntentAllowTool, "approval_allow", "allow", "approved", "allow")
	case "s":
		return m.submitApprovalDecision(service.IntentAllowToolForSession, "approval_allow_session", "allow for session", "approved for session", "allow_session")
	case "d", "esc", "ctrl+c":
		return m.submitApprovalDecision(service.IntentDenyTool, "approval_deny", "deny", "rejected", "deny")
	}
	return nil
}

func (m *model) submitApprovalDecision(kind service.IntentKind, logKind, summary, status, notice string) tea.Cmd {
	m.dispatchIntent(service.Intent{Kind: kind, ToolCallID: m.approval.toolCallID})
	m.addLog(logEntry{Kind: logKind, Source: m.approval.toolName, Summary: summary, Raw: notice})
	m.mode = modeChat
	m.status = status
	m.appendNotice(m.approvalNoticeText(notice))
	return m.flushNativeScrollbackCmd()
}

func (m *model) handleSessionPickerKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = modeChat
	case "up", "k":
		m.sessionIndex = prevSessionChoiceIndex(m.sessionChoices, m.sessionIndex)
	case "down", "j":
		m.sessionIndex = nextSessionChoiceIndex(m.sessionChoices, m.sessionIndex)
	case "enter":
		selected := sessionChoiceNumberAt(m.sessionChoices, m.sessionIndex)
		if selected > 0 {
			m.dispatchIntent(service.Intent{Kind: service.IntentSelectSession, SessionInput: strconv.Itoa(selected)})
		}
		m.mode = modeChat
	}
	return nil
}

func (m *model) handleUserInputKey(msg tea.KeyMsg) tea.Cmd {
	if len(m.userInput.questions) == 0 {
		m.dispatchIntent(service.Intent{Kind: service.IntentCancelUserInput, ToolCallID: m.userInput.toolCallID})
		m.mode = modeChat
		return nil
	}
	q := m.userInput.questions[m.userInput.index]
	switch msg.String() {
	case "esc":
		m.dispatchIntent(service.Intent{Kind: service.IntentCancelUserInput, ToolCallID: m.userInput.toolCallID})
		m.mode = modeChat
	case "up", "k":
		if m.userInput.selectedOption > 0 {
			m.userInput.selectedOption--
		}
	case "down", "j":
		if m.userInput.selectedOption < len(q.Options)-1 {
			m.userInput.selectedOption++
		}
	case "enter":
		opt := q.Options[m.userInput.selectedOption]
		m.userInput.answers = append(m.userInput.answers, core.UserInputAnswer{ID: q.ID, Label: opt.Label, Value: opt.Label})
		m.userInput.index++
		m.userInput.selectedOption = 0
		if m.userInput.index >= len(m.userInput.questions) {
			resp := core.UserInputResponse{Answers: m.userInput.answers}
			m.dispatchIntent(service.Intent{Kind: service.IntentSubmitUserInput, ToolCallID: m.userInput.toolCallID, UserInput: &resp})
			m.mode = modeChat
		}
	}
	return nil
}

func (m *model) handleModelPickerKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		if m.modelPicker.stage > 0 {
			m.modelPicker.stage--
		} else {
			m.mode = modeChat
		}
	case "up", "k":
		if m.modelPicker.stage == 0 && m.modelPicker.modelIx > 0 {
			m.modelPicker.modelIx--
		}
		if m.modelPicker.stage == 1 && m.modelPicker.effIx > 0 {
			m.modelPicker.effIx--
		}
		if m.modelPicker.stage == 2 && m.modelPicker.thinkIx > 0 {
			m.modelPicker.thinkIx--
		}
	case "down", "j":
		if m.modelPicker.stage == 0 && m.modelPicker.modelIx < len(m.modelPicker.models)-1 {
			m.modelPicker.modelIx++
		}
		if m.modelPicker.stage == 1 && m.modelPicker.effIx < len(m.modelPicker.efforts)-1 {
			m.modelPicker.effIx++
		}
		if m.modelPicker.stage == 2 && m.modelPicker.thinkIx < len(m.modelPicker.thinkings)-1 {
			m.modelPicker.thinkIx++
		}
	case "enter":
		if m.modelPicker.stage == 0 {
			m.modelPicker.stage = 1
		} else if m.modelPicker.stage == 1 {
			m.modelPicker.stage = 2
		} else {
			modelName := safeChoice(m.modelPicker.models, m.modelPicker.modelIx)
			effort := safeChoice(m.modelPicker.efforts, m.modelPicker.effIx)
			thinking := safeChoice(m.modelPicker.thinkings, m.modelPicker.thinkIx)
			if modelName != "" && effort != "" && thinking != "" {
				m.dispatchIntent(service.Intent{Kind: service.IntentSetModelAndEffort, Model: modelName, Effort: effort, Thinking: thinking})
				m.model = modelName
				m.effort = effort
				m.thinking = thinking
			}
			m.mode = modeChat
		}
	}
	return nil
}

func (m *model) handlePermissionsPickerKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = modeChat
	case "up", "k":
		if m.permissionsPicker.index > 0 {
			m.permissionsPicker.index--
		}
	case "down", "j":
		if m.permissionsPicker.index < len(m.permissionsPicker.choices)-1 {
			m.permissionsPicker.index++
		}
	case "enter":
		choice := safeChoice(m.permissionsPicker.choices, m.permissionsPicker.index)
		mode := approvalChoiceMode(choice)
		if mode != "" {
			m.dispatchIntent(service.Intent{Kind: service.IntentSetApprovalMode, ApprovalMode: mode})
		}
		m.mode = modeChat
	}
	return nil
}

func (m *model) handlePlanImplementationKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = modeChat
	case "up", "k", "left", "h":
		if m.planImplementation.index > 0 {
			m.planImplementation.index--
		}
	case "down", "j", "right", "l", "tab":
		if m.planImplementation.index < 1 {
			m.planImplementation.index++
		}
	case "enter":
		if m.localSubmitPending > 0 {
			m.status = "wait for command to finish"
			m.refreshViewportContent()
			return m.flushNativeScrollbackCmd()
		}
		if m.planImplementation.index == 0 {
			m.appendTranscript("you", tuirender.KindText, "Implement the plan.")
			m.beginTurnTranscript()
			m.startBusy()
			m.status = "running"
			m.chatMode = "agent"
			m.dispatchIntent(service.Intent{Kind: service.IntentImplementPlan})
			m.mode = modeChat
			m.refreshViewportContentFollow(true)
			return tea.Sequence(m.flushNativeScrollbackCmd(), busyTickCmd())
		}
		m.mode = modeChat
	}
	return nil
}

func (m *model) handleGlobalKey(msg tea.KeyMsg) (tea.Cmd, bool, bool) {
	switch msg.String() {
	case "ctrl+c":
		if strings.TrimSpace(m.input.Value()) != "" {
			m.input.Reset()
			m.skillBinding = nil
			m.resetHistoryNavigation()
			m.updateSlashMatches()
			m.skills.matches = nil
			m.skills.selected = 0
			m.status = "input cleared"
			return nil, false, true
		}
		now := time.Now()
		if !m.quitArmedUntil.IsZero() && now.Before(m.quitArmedUntil) {
			m.dispatchIntent(service.Intent{Kind: service.IntentShutdown})
			return nil, true, true
		}
		m.quitArmedUntil = now.Add(2 * time.Second)
		m.status = "Press Ctrl+C again to quit"
		return armQuitCmd(2 * time.Second), false, true
	case "enter":
		if m.busy {
			m.submitPromptWhileBusy(m.input.Value())
			return m.flushNativeScrollbackCmd(), false, true
		}
		if m.localSubmitPending > 0 {
			m.status = "wait for command to finish"
			m.refreshViewportContent()
			return m.flushNativeScrollbackCmd(), false, true
		}
		if m.hasSlashSuggestions() {
			if cmd := safeChoice(m.slash.matches, m.slash.selected); cmd != "" {
				m.input.SetValue(cmd)
				m.skillBinding = nil
				m.updateSlashMatches()
				if m.shouldAutoRunSlash(cmd) {
					return tea.Sequence(m.flushNativeScrollbackCmd(), m.submitPrompt(cmd)), false, true
				}
			}
			return nil, false, true
		}
		if m.insertSelectedSkill() {
			return nil, false, true
		}
		if m.page == pageLogs && m.logFilterInput.Focused() {
			m.logFilter = strings.TrimSpace(m.logFilterInput.Value())
			m.logFilterInput.Blur()
			return nil, false, true
		}
		if raw := m.input.Value(); strings.HasSuffix(raw, "\\") {
			m.input.SetValue(strings.TrimSuffix(raw, "\\") + "\n")
			m.skillBinding = nil
			m.resetHistoryNavigation()
			m.updateSlashMatches()
			return nil, false, true
		}
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			return nil, false, true
		}
		return tea.Sequence(m.flushNativeScrollbackCmd(), m.submitPrompt(value)), false, true
	}
	return nil, false, false
}

func (m *model) handleComposerKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "shift+enter", "ctrl+j":
		m.input.InsertNewline()
		m.resetHistoryNavigation()
		m.updateSlashMatches()
		return nil, true
	case "ctrl+p":
		m.historyPrev()
		return nil, true
	case "ctrl+n":
		m.historyNext()
		return nil, true
	}
	if m.input.HandleKey(msg) {
		m.resetHistoryNavigation()
		m.updateSlashMatches()
		return nil, true
	}
	return nil, false
}

func (m *model) applyPalette() {
	if m.palette.selected < 0 || m.palette.selected >= len(m.palette.actions) {
		return
	}
	m.palette.actions[m.palette.selected].Run(m)
}
