package policy

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/usewhale/whale/internal/core"
)

type ApprovalRequest struct {
	SessionID string
	ToolCall  core.ToolCall
	Spec      core.ToolSpec
	Reason    string
	Code      string
	Key       string
	Metadata  map[string]any
}

type ApprovalDecision int

const (
	ApprovalDeny ApprovalDecision = iota
	ApprovalAllow
	ApprovalAllowForSession
	ApprovalCancel
)

func (d ApprovalDecision) Approved() bool {
	return d == ApprovalAllow || d == ApprovalAllowForSession
}

func (d ApprovalDecision) ForSession() bool {
	return d == ApprovalAllowForSession
}

func (d ApprovalDecision) Canceled() bool {
	return d == ApprovalCancel
}

type ApprovalFunc func(req ApprovalRequest) ApprovalDecision

type SessionApprovalCache struct {
	mu     sync.RWMutex
	data   map[string]map[string]bool
	loaded map[string]bool
}

func NewSessionApprovalCache() *SessionApprovalCache {
	return &SessionApprovalCache{
		data:   make(map[string]map[string]bool),
		loaded: make(map[string]bool),
	}
}

func (c *SessionApprovalCache) Has(sessionID, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	bySession, ok := c.data[sessionID]
	if !ok {
		return false
	}
	return bySession[key]
}

func (c *SessionApprovalCache) Grant(sessionID, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	bySession, ok := c.data[sessionID]
	if !ok {
		bySession = make(map[string]bool)
		c.data[sessionID] = bySession
	}
	bySession[key] = true
}

func (c *SessionApprovalCache) SetLoaded(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loaded[sessionID] = true
}

func (c *SessionApprovalCache) IsLoaded(sessionID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded[sessionID]
}

func (c *SessionApprovalCache) Merge(sessionID string, keys map[string]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	bySession, ok := c.data[sessionID]
	if !ok {
		bySession = make(map[string]bool)
		c.data[sessionID] = bySession
	}
	for k, v := range keys {
		if v {
			bySession[k] = true
		}
	}
}

func ApprovalKey(call core.ToolCall) string {
	base := call.Name
	var body map[string]any
	if err := json.Unmarshal([]byte(call.Input), &body); err != nil {
		return base + "|" + strings.TrimSpace(call.Input)
	}
	switch call.Name {
	case "edit", "write", "read_file":
		if v, _ := body["file_path"].(string); strings.TrimSpace(v) != "" {
			return base + "|file:" + strings.TrimSpace(v)
		}
	case "shell_run":
		if v, _ := body["command"].(string); strings.TrimSpace(v) != "" {
			return base + "|cmd:" + strings.TrimSpace(v)
		}
	}
	return base + "|" + strings.TrimSpace(call.Input)
}

func ApprovalSummary(call core.ToolCall) string {
	var body map[string]any
	if err := json.Unmarshal([]byte(call.Input), &body); err != nil {
		return call.Name
	}
	switch call.Name {
	case "shell_run":
		if v, _ := body["command"].(string); strings.TrimSpace(v) != "" {
			return fmt.Sprintf("shell_run: %s", strings.TrimSpace(v))
		}
	case "write":
		if v, _ := body["file_path"].(string); strings.TrimSpace(v) != "" {
			return fmt.Sprintf("write: %s", strings.TrimSpace(v))
		}
	case "edit":
		if v, _ := body["file_path"].(string); strings.TrimSpace(v) != "" {
			return fmt.Sprintf("edit: %s", strings.TrimSpace(v))
		}
	case "apply_patch":
		return "apply_patch: patch payload"
	}
	return call.Name
}

func ApprovalScope(call core.ToolCall) string {
	var body map[string]any
	if err := json.Unmarshal([]byte(call.Input), &body); err != nil {
		return "workspace"
	}
	if v, _ := body["file_path"].(string); strings.TrimSpace(v) != "" {
		return "file:" + strings.TrimSpace(v)
	}
	if v, _ := body["path"].(string); strings.TrimSpace(v) != "" {
		return "path:" + strings.TrimSpace(v)
	}
	if call.Name == "shell_run" {
		return "shell"
	}
	if call.Name == "apply_patch" {
		return "patch"
	}
	return "workspace"
}
