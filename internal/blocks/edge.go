// Package blocks — Edge is the Cloudflare edge-agent primitive.
// Fiber journal (resumable state) + mesh peer (WebSocket P2P).
// Activated when CF_API_TOKEN is set or via 'ultrawhale setup'.
package blocks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// EdgeAgent is a Cloudflare Durable Object deployed at the edge.
// Each edge agent has a fiber journal for resumability and mesh peers.
type EdgeAgent struct {
	mu sync.Mutex

	// Identity
	ID       string    // Durable Object ID
	Name     string    // "ultrawhale-indexer", "ultrawhale-planner"
	Role     string    // "indexer", "planner", "validator"
	Zone     string    // CF zone/domain

	// Fiber journal (resumable state)
	Fiber   *Fiber

	// Mesh peers
	Peers   []string  // other edge agent IDs

	// State
	Status   string    // "deploying", "active", "idle"
	DeployedAt time.Time
	POV       POV
}

// Fiber is a resumable agent state journal.
// Every prompt, tool call, and response is appended to the ledger.
// If client disconnects, agent continues. Reconnect resumes from last token.
type Fiber struct {
	ID       string
	Ledger   []FiberEntry     // append-only transaction log
	Buffer   []string         // buffered output tokens since last client ack
	Resumed  bool
	LastAck  time.Time
}

// FiberEntry is one entry in the fiber transaction ledger.
type FiberEntry struct {
	Seq       int       `json:"seq"`
	Timestamp time.Time `json:"ts"`
	Type      string    `json:"type"` // "prompt", "tool_call", "tool_result", "response_token"
	Content   string    `json:"content"`
	Ref       string    `json:"ref"` // sha256 of content
}

// NewEdgeAgent creates a new CF edge agent.
func NewEdgeAgent(name, role string) *EdgeAgent {
	_ = CurrentPOV()
	return &EdgeAgent{
		ID:     fmt.Sprintf("ultrawhale-%s-%d", role, time.Now().Unix()),
		Name:   name,
		Role:   role,
		Fiber:  NewFiber(),
		Peers:  []string{},
		Status: "idle",
	}
}

// NewFiber creates a new resumable fiber journal.
func NewFiber() *Fiber {
	return &Fiber{
		ID:      fmt.Sprintf("fiber-%d", time.Now().UnixNano()),
		Ledger:  make([]FiberEntry, 0, 256),
		Buffer:  make([]string, 0, 64),
	}
}

// AppendLedger writes an entry to the fiber journal.
func (f *Fiber) AppendLedger(typ, content string) {
	entry := FiberEntry{
		Seq:       len(f.Ledger),
		Timestamp: time.Now(),
		Type:      typ,
		Content:   content,
		Ref:       Ref([]byte(content + time.Now().String()))[:12],
	}
	f.Ledger = append(f.Ledger, entry)
	if len(f.Ledger) > 4096 {
		f.Ledger = f.Ledger[len(f.Ledger)-4096:] // keep last 4096 entries
	}
}

// BufferToken adds a token to the output buffer.
func (f *Fiber) BufferToken(token string) {
	f.Buffer = append(f.Buffer, token)
}

// FlushBuffer drains and returns buffered tokens.
func (f *Fiber) FlushBuffer() []string {
	out := f.Buffer
	f.Buffer = make([]string, 0, 64)
	return out
}

// Resume replays the ledger from the last acknowledged sequence.
func (f *Fiber) Resume(lastAckSeq int) []FiberEntry {
	f.Resumed = true
	if lastAckSeq < 0 { lastAckSeq = 0 }
	if lastAckSeq >= len(f.Ledger) { return nil }
	return f.Ledger[lastAckSeq:]
}

// ── CF Wrangler integration ───────────────────────────────────────────

// CFConfig holds Cloudflare credentials and project settings.
type CFConfig struct {
	APIToken   string // CF_API_TOKEN
	AccountID  string // CF_ACCOUNT_ID
	ZoneID     string // CF zone for the domain
	Project    string // wrangler project name
	DeployURL  string // e.g. https://ultrawhale-edge.crabcc.workers.dev
}

// DetectCF checks if Cloudflare credentials are available.
func DetectCF() *CFConfig {
	token := os.Getenv("CF_API_TOKEN")
	if token == "" {
		token = os.Getenv("CLOUDFLARE_API_TOKEN")
	}
	if token == "" {
		return nil
	}
	return &CFConfig{
		APIToken:  token,
		AccountID: os.Getenv("CF_ACCOUNT_ID"),
		Project:   "ultrawhale-edge",
	}
}

// DeployEdgeAgent deploys an edge agent to Cloudflare via wrangler.
func (e *EdgeAgent) DeployEdgeAgent(cfg *CFConfig) error {
	if cfg == nil {
		return fmt.Errorf("no CF config — run 'ultrawhale setup' or set CF_API_TOKEN")
	}

	// Check wrangler is available
	if _, err := exec.LookPath("wrangler"); err != nil {
		return fmt.Errorf("wrangler not found — install: npm i -g wrangler")
	}

	e.Status = "deploying"

	// Build the worker script
	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("ultrawhale-edge-%s.js", e.ID))
	workerJS := edgeWorkerTemplate(e)
	if err := os.WriteFile(scriptPath, []byte(workerJS), 0o644); err != nil {
		return err
	}

	// Deploy via wrangler
	cmd := exec.Command("wrangler", "deploy", scriptPath,
		"--name", e.ID,
		"--compatibility-date", time.Now().Format("2006-01-02"),
	)
	cmd.Env = append(os.Environ(),
		"CLOUDFLARE_API_TOKEN="+cfg.APIToken,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wrangler deploy: %v\n%s", err, string(out))
	}

	e.DeployedAt = time.Now()
	e.Status = "active"
	return nil
}

// edgeWorkerTemplate generates a Cloudflare Worker JS script.
func edgeWorkerTemplate(e *EdgeAgent) string {
	return fmt.Sprintf(`// ultrawhale edge agent: %s (%s)
// Deployed: %s
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    
    // Health check
    if (url.pathname === "/health") {
      return new Response(JSON.stringify({
        agent: "%s",
        role: "%s",
        status: "active",
        fiber_id: "%s"
      }), { headers: { "Content-Type": "application/json" } });
    }
    
    // Fiber resume — replay ledger from last sequence
    if (url.pathname === "/fiber/resume") {
      const lastSeq = parseInt(url.searchParams.get("seq") || "0");
      // In production: read from Durable Object storage (SQLite)
      return new Response(JSON.stringify({ resumed: true, from: lastSeq }));
    }
    
    // Default: proxy to ultrawhale orchestrator
    return new Response(JSON.stringify({
      agent: "%s",
      message: "edge agent active — use /health or /fiber/resume"
    }), { headers: { "Content-Type": "application/json" } });
  }
};
`, e.Name, e.Role, time.Now().Format(time.RFC3339),
		e.ID, e.Role, e.Fiber.ID,
		e.ID)
}

// EdgeStatus returns compact status for HUD/sidepanel.
func EdgeStatus() string {
	cfg := DetectCF()
	if cfg == nil {
		return "cf: not configured (run 'ultrawhale setup')"
	}
	return fmt.Sprintf("cf: %s (project: %s)", cfg.APIToken[:8]+"...", cfg.Project)
}

// SetupConfig holds all infra configuration options.
type SetupConfig struct {
	CF      *CFConfig
	Langfuse struct {
		PublicKey string
		SecretKey string
		Host      string
	}
	NATS struct {
		URL string
	}
	Bao struct {
		URL   string
		Token string
	}
	Supabase struct {
		URL      string
		AnonKey  string
	}
}
