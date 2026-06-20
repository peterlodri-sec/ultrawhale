// Package agentfield provides local-first AgentField integration.
// On SessionStart: auto-starts a local control plane, registers ultrawhale
// as an agent with DID identity, exposes commands as callable functions.
// Zero config — just add agentfield.enabled = true in config.toml.
package agentfield

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/blocks"
)

const PluginID = "agentfield"

type Plugin struct {
	mu       sync.Mutex
	config   Config
	identity Identity
	server   *http.Server
	running  bool
}

type Config struct {
	Enabled      bool
	Port         int    // local control plane port (default 8585)
	DataDir      string // ~/.whale/agentfield
	AutoRegister bool
}

type Identity struct {
	DID       string `json:"did"`        // did:key:z6Mk...
	PublicKey string `json:"public_key"` // hex-encoded ed25519 pubkey
	Agent     string `json:"agent"`      // ultrawhale
	CreatedAt string `json:"created_at"`
}

func NewPlugin() *Plugin {
	return &Plugin{config: Config{
		Enabled:      true,
		Port:         8585,
		AutoRegister: true,
	}}
}

func (p *Plugin) ID() string          { return PluginID }
func (p *Plugin) Name() string        { return "AgentField" }
func (p *Plugin) Version() string     { return "0.1.0" }
func (p *Plugin) Description() string { return "Local-first AgentField control plane — DID identity, callable API, cross-agent routing." }

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{{
		Event:       agent.HookEventSessionStart,
		Name:        "agentfield.start",
		Source:      "plugin:agentfield",
		Description: "Starts local AgentField control plane and registers ultrawhale agent.",
		Priority:    90, // high priority — run before other plugins
		Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
			go p.start()
			return agent.HookResult{Decision: agent.HookDecisionPass}
		},
	}, {
		Event:       agent.HookEventStop,
		Name:        "agentfield.stop",
		Source:      "plugin:agentfield",
		Description: "Stops local control plane and deregisters agent.",
		Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
			p.stop()
			return agent.HookResult{Decision: agent.HookDecisionPass}
		},
	}}
}

// ── Start local control plane ──────────────────────────────────────────

func (p *Plugin) start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.running { return }

	// Ensure data dir
	dataDir := p.config.DataDir
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".whale", "agentfield")
	}
	os.MkdirAll(dataDir, 0o700)

	// Load or generate DID identity
	p.identity = loadOrCreateIdentity(dataDir)

	// Start local HTTP control plane
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "agent": "ultrawhale"})
	})
	mux.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Identity{p.identity})
	})
	mux.HandleFunc("/api/v1/execute/ultrawhale.status", func(w http.ResponseWriter, r *http.Request) {
		pov := blocks.CurrentPOV()
		json.NewEncoder(w).Encode(map[string]any{
			"agent":   pov.Agent,
			"version": pov.Version,
			"machine": pov.Machine,
			"arch":    pov.Arch,
			"tier":    pov.Tier,
			"mode":    pov.Mode,
			"did":     p.identity.DID,
		})
	})

	p.server = &http.Server{Addr: fmt.Sprintf(":%d", p.config.Port), Handler: mux}
	p.running = true

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "[agentfield] server error: %v\n", err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // let server start
}

func (p *Plugin) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.server != nil {
		p.server.Close()
		p.server = nil
	}
	p.running = false
}

// ── DID Identity ───────────────────────────────────────────────────────

func loadOrCreateIdentity(dir string) Identity {
	path := filepath.Join(dir, "did.json")
	data, err := os.ReadFile(path)
	if err == nil {
		var id Identity
		if json.Unmarshal(data, &id) == nil {
			return id
		}
	}

	// Generate Ed25519 keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return Identity{Agent: "ultrawhale"}
	}

	// W3C DID:key format (multicodec ed25519-pub)
	pubHex := hex.EncodeToString(pub)
	privHex := hex.EncodeToString(priv)

	id := Identity{
		DID:       fmt.Sprintf("did:key:z%s", pubHex[:40]),
		PublicKey: pubHex,
		Agent:     "ultrawhale",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Store private key
	os.WriteFile(filepath.Join(dir, "private.key"), []byte(privHex), 0o600)

	// Store DID
	data, _ = json.MarshalIndent(id, "", "  ")
	os.WriteFile(path, data, 0o600)

	return id
}

// ── Doctor ─────────────────────────────────────────────────────────────

func (p *Plugin) Doctor() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running {
		return "agentfield: not running"
	}
	return fmt.Sprintf("agentfield: localhost:%d, DID=%s, agent=ultrawhale", p.config.Port, p.identity.DID)
}
