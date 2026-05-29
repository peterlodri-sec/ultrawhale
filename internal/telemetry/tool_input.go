package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ToolInputEventsSuffix = ".tool_input_events.jsonl"

type ToolInputEvent struct {
	TS                 int64  `json:"ts"`
	Session            string `json:"session"`
	Model              string `json:"model,omitempty"`
	AssistantMessageID string `json:"assistant_message_id,omitempty"`
	ToolCallID         string `json:"tool_call_id,omitempty"`
	Tool               string `json:"tool,omitempty"`
	Event              string `json:"event"`
	RepairKind         string `json:"repair_kind,omitempty"`
	Path               string `json:"path,omitempty"`
	BeforeType         string `json:"before_type,omitempty"`
	AfterType          string `json:"after_type,omitempty"`
	ErrorCode          string `json:"error_code,omitempty"`
}

func ToolInputEventsPath(sessionsDir, sessionID string) string {
	return filepath.Join(strings.TrimSpace(sessionsDir), sanitizeSessionID(sessionID)+ToolInputEventsSuffix)
}

func AppendToolInputEvent(sessionsDir string, rec ToolInputEvent, now time.Time) error {
	sessionsDir = strings.TrimSpace(sessionsDir)
	if sessionsDir == "" || strings.TrimSpace(rec.Session) == "" || strings.TrimSpace(rec.Event) == "" {
		return nil
	}
	if rec.TS == 0 {
		rec.TS = now.UnixMilli()
	}
	path := ToolInputEventsPath(sessionsDir, rec.Session)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}

func sanitizeSessionID(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "default"
	}
	v = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		default:
			return '_'
		}
	}, v)
	return v
}
