package service

import (
	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/core"
	"github.com/usewhale/whale/internal/policy"
)

func (s *Service) awaitApproval(req policy.ApprovalRequest) policy.ApprovalDecision {
	toolCallID := req.ToolCall.ID
	s.interactionMu.Lock()
	if s.shutdownRequested {
		s.interactionMu.Unlock()
		return policy.ApprovalCancel
	}
	s.approveMu.Lock()
	if s.sessionGrantLocked(req.SessionID, req.Key) {
		s.approveMu.Unlock()
		s.interactionMu.Unlock()
		return policy.ApprovalAllowForSession
	}
	ch := make(chan policy.ApprovalDecision, 1)
	s.approvals[toolCallID] = ch
	s.approveMu.Unlock()
	s.interactionMu.Unlock()
	s.emit(Event{Kind: EventApprovalRequired, ToolCallID: toolCallID, ToolName: req.ToolCall.Name, Text: policy.ApprovalSummary(req.ToolCall), Metadata: req.Metadata})
	decision := <-ch
	s.approveMu.Lock()
	delete(s.approvals, toolCallID)
	if decision == policy.ApprovalAllowForSession {
		s.grantSessionLocked(req.SessionID, req.Key)
	}
	s.approveMu.Unlock()
	return decision
}

func (s *Service) resolveApproval(toolCallID string, decision policy.ApprovalDecision) {
	s.approveMu.Lock()
	ch, ok := s.approvals[toolCallID]
	if ok {
		delete(s.approvals, toolCallID)
	}
	s.approveMu.Unlock()
	if !ok {
		if s.interactionShutdownRequested() {
			return
		}
		s.emit(Event{Kind: EventError, Text: "no pending approval for tool call"})
		return
	}
	select {
	case ch <- decision:
	default:
	}
}

func (s *Service) sessionGrantLocked(sessionID, key string) bool {
	bySession, ok := s.sessionGrants[sessionID]
	if !ok {
		return false
	}
	return bySession[key]
}

func (s *Service) grantSessionLocked(sessionID, key string) {
	bySession, ok := s.sessionGrants[sessionID]
	if !ok {
		bySession = map[string]bool{}
		s.sessionGrants[sessionID] = bySession
	}
	bySession[key] = true
}

func (s *Service) awaitUserInput(req agent.UserInputRequest) (core.UserInputResponse, bool) {
	toolCallID := req.ToolCall.ID
	ch := make(chan userInputDecision, 1)
	s.interactionMu.Lock()
	if s.shutdownRequested {
		s.interactionMu.Unlock()
		return core.UserInputResponse{}, false
	}
	s.inputMu.Lock()
	s.inputs[toolCallID] = ch
	s.inputMu.Unlock()
	s.interactionMu.Unlock()
	s.emit(Event{Kind: EventUserInputRequired, ToolCallID: toolCallID, ToolName: req.ToolCall.Name, Questions: req.Questions})
	decision := <-ch
	s.inputMu.Lock()
	delete(s.inputs, toolCallID)
	s.inputMu.Unlock()
	return decision.response, decision.ok
}

func (s *Service) resolveUserInput(toolCallID string, resp core.UserInputResponse, ok bool) {
	s.inputMu.Lock()
	ch, exists := s.inputs[toolCallID]
	if exists {
		delete(s.inputs, toolCallID)
	}
	s.inputMu.Unlock()
	if !exists {
		if s.interactionShutdownRequested() {
			return
		}
		s.emit(Event{Kind: EventError, Text: "no pending user input"})
		return
	}
	select {
	case ch <- userInputDecision{response: resp, ok: ok}:
	default:
	}
}

func (s *Service) cancelPendingInteractions() {
	s.interactionMu.Lock()
	s.shutdownRequested = true
	s.approveMu.Lock()
	approvals := make([]chan policy.ApprovalDecision, 0, len(s.approvals))
	for id, ch := range s.approvals {
		approvals = append(approvals, ch)
		delete(s.approvals, id)
	}
	s.approveMu.Unlock()
	for _, ch := range approvals {
		select {
		case ch <- policy.ApprovalCancel:
		default:
		}
	}

	s.inputMu.Lock()
	inputs := make([]chan userInputDecision, 0, len(s.inputs))
	for id, ch := range s.inputs {
		inputs = append(inputs, ch)
		delete(s.inputs, id)
	}
	s.inputMu.Unlock()
	s.interactionMu.Unlock()
	for _, ch := range inputs {
		select {
		case ch <- userInputDecision{}:
		default:
		}
	}
}

func (s *Service) resetInteractionShutdown() {
	s.interactionMu.Lock()
	s.shutdownRequested = false
	s.interactionMu.Unlock()
}

func (s *Service) interactionShutdownRequested() bool {
	s.interactionMu.Lock()
	defer s.interactionMu.Unlock()
	return s.shutdownRequested
}
