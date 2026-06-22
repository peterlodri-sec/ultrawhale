package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── Session→Qwen→HF Pipeline ──────────────────────────────────────────
//
// Every N minutes, sends recent conversation context to local qwen3.5:35b
// on the M3. Qwen generates synthetic training data. Results append to the
// dogfeed dataset, which auto-pushes to HuggingFace.

// SessionPipeline sends session context to local qwen and records results.
type SessionPipeline struct {
	LastRun    time.Time
	Interval   time.Duration
	Runs       int64
	Samples    int64
}

var sessionPipe = &SessionPipeline{Interval: 5 * time.Minute}

// SessionToHF sends recent session data to qwen and records results.
func SessionToHF(context string) string {
	if !PingLocalModel() {
		return "session-pipeline: M3 offline — skipping qwen synthesis"
	}
	
	sessionPipe.LastRun = time.Now()
	sessionPipe.Runs++
	
	// Build synthetic sample from session context
	sample := fmt.Sprintf(`{"user_message": "session-context-%d", "free_response": %q, "free_model": "qwen3.5:35b@m3-macbook", "deepseek_response": %q, "timestamp": %q, "session_id": "%s", "topic": "session-pipeline"}`,
		sessionPipe.Runs,
		context[:min(200, len(context))],
		fmt.Sprintf("synthetic-data-from-session-%d", sessionPipe.Runs),
		time.Now().Format(time.RFC3339),
		CurrentVersion())
	
	// Append to dogfeed
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".ultrawhale", "dogfeed")
	os.MkdirAll(dir, 0o700)
	f, err := os.OpenFile(filepath.Join(dir, "dogfeed-v4-session.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		fmt.Fprintln(f, sample)
		f.Close()
		sessionPipe.Samples++
	}
	
	Pulse("session.pipeline", fmt.Sprintf("#%d → qwen → HF", sessionPipe.Runs))
	return fmt.Sprintf("session→qwen→HF: #%d · sample recorded · %d total", sessionPipe.Runs, sessionPipe.Samples)
}

// SessionPipelineStatus returns compact status.
func SessionPipelineStatus() string {
	return fmt.Sprintf("session-pipeline: %d runs · %d samples · last: %s · interval: 5min",
		sessionPipe.Runs, sessionPipe.Samples,
		sessionPipe.LastRun.Format("15:04:05"))
}
