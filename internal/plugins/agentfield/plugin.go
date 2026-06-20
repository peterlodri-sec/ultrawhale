// Package agentfield provides Supabase-backed AgentField control plane.

// Minimal: DID identity, REST API via PostgREST, workflow CRUD.
// Runs local Supabase (Postgres+PostgREST+GoTrue) on localhost:8585.
package agentfield


import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"context"
	"net/http"
	"strings"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/plugintypes"
	"github.com/usewhale/whale/internal/blocks"
)

const PluginID = "agentfield"

type Plugin struct {
	mu       sync.Mutex
	config   Config
	identity Identity
	server   *http.Server
	running  bool
	supabase *SupabaseConfig
}

type Config struct {
	Enabled       bool
	Port          int
	DataDir       string
	SupabaseURL   string // http://localhost:8586 for PostgREST
	AnonKey       string // public anon key
	ServiceKey    string // service_role key for admin
}

type SupabaseConfig struct {
	URL        string
	AnonKey    string
	ServiceKey string
}

type Identity struct {
	DID       string `json:"did"`
	PublicKey string `json:"public_key"`
	Agent     string `json:"agent"`
}

func NewPlugin() *Plugin {
	return &Plugin{config: Config{
		Enabled:     true,
		Port:        8585,
		SupabaseURL: "http://localhost:8586",
	}}
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "AgentField" }
func (p *Plugin) Version() string { return "0.2.0" }
func (p *Plugin) Description() string {
	return "Supabase-backed AgentField — DID identity, PostgREST API, workflow CRUD."
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{
		{Event: agent.HookEventSessionStart, Name: "af.start", Source: "plugin:agentfield",
			Priority: 90,
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				go p.start()
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
		{Event: agent.HookEventStop, Name: "af.stop", Source: "plugin:agentfield",
			Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
				p.stop()
				return agent.HookResult{Decision: agent.HookDecisionPass}
			}},
	}
}

// ── Start ──────────────────────────────────────────────────────────────

func (p *Plugin) start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running { return }

	home, _ := os.UserHomeDir()
	if p.config.DataDir == "" {
		p.config.DataDir = filepath.Join(home, ".whale", "agentfield")
	}
	os.MkdirAll(p.config.DataDir, 0o700)

	// Generate or load DID
	p.identity = loadOrCreateIdentity(p.config.DataDir)

	// Configure Supabase
	p.supabase = &SupabaseConfig{
		URL:     p.config.SupabaseURL,
		AnonKey: p.config.AnonKey,
	}

	// Start HTTP control plane
	mux := http.NewServeMux()
	p.registerRoutes(mux)

	p.server = &http.Server{Addr: fmt.Sprintf(":%d", p.config.Port), Handler: mux}
	p.running = true
	go p.server.ListenAndServe()
	time.Sleep(50 * time.Millisecond)
}

func (p *Plugin) registerRoutes(mux *http.ServeMux) {
	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"agent":     "ultrawhale",
			"supabase":  p.config.SupabaseURL,
			"did":       p.identity.DID,
			"brain":     blocks.BrainStatus(),
		})
	})

	// Agent status
	mux.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		pov := blocks.CurrentPOV()
			cur := blocks.GetCurrent()
		json.NewEncoder(w).Encode([]map[string]any{{
			"agent":   pov.Agent,
			"version": pov.Version,
			"machine": pov.Machine,
			"arch":    pov.Arch,
			"tier":    pov.Tier,
				"turn_count": cur.TurnCount,
				"total_tokens": cur.TotalTokens,
				"memory_mb": cur.MemoryMB,
				"cost_usd": cur.CostUSD,
			"did":     p.identity.DID,
		}})
	})

	// Execute agent command
	mux.HandleFunc("/api/v1/edge/status", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": blocks.EdgeStatus()})
	})
	mux.HandleFunc("/api/v1/edge/deploy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" { w.WriteHeader(405); return }
		var req struct { Type string; ID string }
		json.NewDecoder(r.Body).Decode(&req)
		var result string
		if req.Type == "agent" {
			// Only pure subagents (read_only/write) are edge-deployable.
			// Swarms have their own AgentField + HTTP ports — not deployable to CF Worker.
			agents := blocks.ListAgents()
			for _, a := range agents {
				if a.ID == req.ID || req.ID == "" {
					if _, err := a.DeployToEdge(); err != nil {
						result = err.Error()
					} else {
						result = fmt.Sprintf("deployed %s to edge", a.ID[:12])
					}
				}
			}
			if result == "" { result = "no matching agents found" }
		}
		json.NewEncoder(w).Encode(map[string]string{"result": result})
	})
	mux.HandleFunc("/api/v1/memos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			memos := blocks.RecallSessionMemos()
			json.NewEncoder(w).Encode(memos)
		case "POST":
			var req struct { Content string }
			json.NewDecoder(r.Body).Decode(&req)
			m := blocks.RememberSessionMemo(req.Content)
			json.NewEncoder(w).Encode(m)
		default:
			w.WriteHeader(405)
		}
	})
	mux.HandleFunc("/api/v1/execute/ultrawhale.", func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Path[strings.LastIndex(r.URL.Path, ".")+1:]
		if !allowedCommands[cmd] {
			w.WriteHeader(403)
			return
		}
		pov := blocks.CurrentPOV()
			cur := blocks.GetCurrent()
		json.NewEncoder(w).Encode(map[string]any{
			"agent":   pov.Agent,
			"command": cmd,
			"status":  "executed",
			"tier":    pov.Tier,
				"turn_count": cur.TurnCount,
				"total_tokens": cur.TotalTokens,
				"memory_mb": cur.MemoryMB,
				"cost_usd": cur.CostUSD,
		})
	})

	// Workflow CRUD — proxied to Supabase PostgREST
	mux.HandleFunc("/api/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// List workflows via PostgREST
			resp, _ := http.Get(p.config.SupabaseURL + "/workflows")
			if resp != nil {
				defer resp.Body.Close()
				var data any
				json.NewDecoder(resp.Body).Decode(&data)
				json.NewEncoder(w).Encode(data)
			}
		case "POST":
			if err := validateWorkflowInput(r); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			// Create workflow via PostgREST
			http.Post(p.config.SupabaseURL+"/workflows", "application/json", r.Body)
			w.WriteHeader(201)
		default:
			w.WriteHeader(405)
		}
	})

	mux.HandleFunc("/api/v1/workflows/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		resp, _ := http.Get(p.config.SupabaseURL + "/workflows?id=eq." + id)
		if resp != nil {
			defer resp.Body.Close()
			var data any
			json.NewDecoder(resp.Body).Decode(&data)
			json.NewEncoder(w).Encode(data)
		}
	})
}

func (p *Plugin) stopOld() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.server != nil { p.server.Close(); p.server = nil }
	p.running = false
}

// ── DID Identity ───────────────────────────────────────────────────────

func loadOrCreateIdentity(dir string) Identity {
	path := filepath.Join(dir, "did.json")
	data, err := os.ReadFile(path)
	if err == nil {
		var id Identity
		if json.Unmarshal(data, &id) == nil { return id }
	}
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	pubHex := pemFormat(pub)
	id := Identity{
		DID:       fmt.Sprintf("did:key:z%s", pubHex[:40]),
		PublicKey: pubHex,
		Agent:     "ultrawhale",
	}
	raw, _ := json.MarshalIndent(id, "", "  ")
	os.WriteFile(path, raw, 0o600)
	privPath := filepath.Join(dir, "private.key")
	os.WriteFile(privPath, []byte(hex.EncodeToString(priv)), 0o600)
	return id
}

func (p *Plugin) Doctor() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running { return "agentfield: not running" }
	return fmt.Sprintf("agentfield: localhost:%d, supabase=%s, DID=%s",
		p.config.Port, p.config.SupabaseURL, p.identity.DID[:45])
}

var allowedCommands = map[string]bool{
	"status": true, "doctor": true, "health": true, "version": true,
}

func validateWorkflowInput(r *http.Request) error {
	if r.ContentLength > 1_000_000 { // 1MB cap
		return fmt.Errorf("script too large (max 1MB)")
	}
	return nil
}

// gracefulShutdown drains in-flight requests before closing.
func (p *Plugin) gracefulShutdown() {
	p.mu.Lock()
	srv := p.server
	p.mu.Unlock()
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}
}

func (p *Plugin) stop() {
	p.gracefulShutdown()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.running = false
}

func pemFormat(pub ed25519.PublicKey) string {
	return hex.EncodeToString(pub) // stored as hex for now; PEM encode on export
}

func (p *Plugin) Manifest() plugintypes.Manifest {
	return plugintypes.Manifest{ID: p.ID(), Name: p.Name(), Version: p.Version(), Description: p.Description()}
}


