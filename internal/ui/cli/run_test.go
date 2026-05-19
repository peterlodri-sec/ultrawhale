package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usewhale/whale/internal/app"
	"github.com/usewhale/whale/internal/core"
	"github.com/usewhale/whale/internal/session"
	"github.com/usewhale/whale/internal/store"
)

func TestRunPreservesHiddenSyntheticLocalCommandPrompt(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "test-key")
	workspace := t.TempDir()
	t.Chdir(workspace)
	writeCLITestSkill(t, filepath.Join(workspace, ".whale", "skills", "test-skill"), "test-skill")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"done\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	stdin := strings.NewReader("/skills-improver propose test-skill\n")
	restoreStdin := replaceStdin(t, stdin)
	defer restoreStdin()

	cfg := app.DefaultConfig()
	cfg.DataDir = t.TempDir()
	cfg.APIBaseURL = srv.URL
	cfg.ApprovalMode = "never"
	if err := Run(cfg, app.StartOptions{NewSession: true}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	summaries, err := session.ListSessions(store.DefaultSessionsDir(cfg.DataDir), 1)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("sessions = %d, want 1", len(summaries))
	}
	if strings.Contains(summaries[0].Meta.Title, "save_skill_proposal") {
		t.Fatalf("synthetic prompt leaked into session title: %q", summaries[0].Meta.Title)
	}
	if summaries[0].Conversation != "(no message yet)" {
		t.Fatalf("synthetic prompt should not be visible conversation title: %q", summaries[0].Conversation)
	}

	msg := firstSessionMessage(t, filepath.Join(store.DefaultSessionsDir(cfg.DataDir), summaries[0].ID+".jsonl"))
	if msg.Role != core.RoleUser || !msg.Hidden || !strings.Contains(msg.Text, "save_skill_proposal") {
		t.Fatalf("first message role=%s hidden=%v text=%q", msg.Role, msg.Hidden, msg.Text)
	}
}

func replaceStdin(t *testing.T, input *strings.Reader) func() {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdin: %v", err)
	}
	if _, err := input.WriteTo(w); err != nil {
		t.Fatalf("write stdin pipe: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	orig := os.Stdin
	os.Stdin = r
	return func() {
		os.Stdin = orig
		_ = r.Close()
	}
}

func writeCLITestSkill(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	content := "---\nname: " + name + "\ndescription: Test skill.\n---\n\n# Test Skill\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
}

func firstSessionMessage(t *testing.T, path string) core.Message {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatalf("empty session file: %v", scanner.Err())
	}
	var msg core.Message
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		t.Fatalf("decode session message: %v", err)
	}
	return msg
}
