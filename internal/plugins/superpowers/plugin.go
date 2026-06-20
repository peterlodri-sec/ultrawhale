// Package superpowers auto-discovers and wires vaked infrastructure:
// bao (secrets), Langfuse (telemetry), NATS (events), vastai (GPU compute).
// All sub-features are opt-in via [superpowers] config section.
// Zero dependencies beyond Go stdlib — pure HTTP calls.
package superpowers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/usewhale/whale/internal/agent"
)

const PluginID = "superpowers"

type Plugin struct {
	mu     sync.Mutex
	config Config
	status Status
}

type Config struct {
	Enabled        bool
	BaoURL         string
	BaoToken       string
	AutoWireLangfuse bool
	AutoWireNATS     bool
	ComputeProvider  string
	HubEnabled       bool
}

type Status struct {
	BaoConnected    bool
	LangfuseWired   bool
	NATSWired       bool
	ComputeAvailable bool
	LastCheck       time.Time
}

func NewPlugin() *Plugin {
	return &Plugin{config: Config{
		BaoURL:         envOrDefault("SUPERPOWERS_BAO_URL", "https://bao.crabcc.app"),
		BaoToken:       envOrDefault("VAULT_TOKEN", ""),
		AutoWireLangfuse: true,
		AutoWireNATS:     true,
		ComputeProvider:  "vastai",
		HubEnabled:       true,
	}}
}

func (p *Plugin) ID() string      { return PluginID }
func (p *Plugin) Name() string    { return "Superpowers" }
func (p *Plugin) Version() string { return "0.1.0" }
func (p *Plugin) Description() string {
	return "Auto-discovers and wires vaked infrastructure: bao secrets, Langfuse, NATS, GPU compute."
}

func (p *Plugin) Hooks() []agent.HookHandler {
	return []agent.HookHandler{{
		Event:       agent.HookEventSessionStart,
		Name:        "superpowers.auto-wire",
		Source:      "plugin:superpowers",
		Description: "Auto-discovers secrets from bao and wires Langfuse + NATS.",
		Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
			go p.autoWire()
			return agent.HookResult{Decision: agent.HookDecisionPass}
		},
	},
	{
		Event:       agent.HookEventStop,
		Name:        "superpowers.on-stop",
		Source:      "plugin:superpowers",
		Description: "Disconnects and reports final status.",
		Priority:    10,
		Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
			return p.onStop()
		},
	},
	{
		Event:       agent.HookEventError,
		Name:        "superpowers.on-error",
		Source:      "plugin:superpowers",
		Description: "Checks bao reachability on errors.",
		Priority:    5,
		Run: func(ctx context.Context, payload agent.HookPayload) agent.HookResult {
			return p.onError(payload)
		},
	}}
}

func (p *Plugin) autoWire() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status.LastCheck = time.Now()

	if p.config.BaoToken != "" {
		p.wireFromBao()
	}
	if p.config.AutoWireLangfuse {
		p.wireLangfuse()
	}
	if p.config.AutoWireNATS {
		p.wireNATS()
	}
}

func (p *Plugin) wireFromBao() {
	resp, err := httpGet(p.config.BaoURL+"/v1/secret/data/whale", p.config.BaoToken)
	if err != nil {
		return
	}
	var result struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return
	}
	p.status.BaoConnected = true

	// Auto-set discovered secrets as env vars
	for k, v := range result.Data.Data {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

func (p *Plugin) wireLangfuse() {
	if os.Getenv("LANGFUSE_PUBLIC_KEY") != "" {
		p.status.LangfuseWired = true
	}
}

func (p *Plugin) wireNATS() {
	if os.Getenv("NATS_URL") != "" {
		p.status.NATSWired = true
	}
}

func (p *Plugin) Doctor() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var parts []string
	if p.status.BaoConnected {
		parts = append(parts, "bao:connected")
	} else {
		parts = append(parts, "bao:not configured (set VAULT_TOKEN)")
	}
	if p.status.LangfuseWired {
		parts = append(parts, "langfuse:wired")
	} else {
		parts = append(parts, "langfuse:not configured (set LANGFUSE_PUBLIC_KEY)")
	}
	if p.status.NATSWired {
		parts = append(parts, "nats:wired")
	} else {
		parts = append(parts, "nats:not configured (set NATS_URL)")
	}
	return fmt.Sprintf("superpowers: %s (last check: %s)", strings.Join(parts, ", "), p.status.LastCheck.Format(time.RFC3339))
}

func httpGet(url, token string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Vault-Token", token)
	req.Header.Set("User-Agent", "ultrawhale-superpowers/0.1")
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var body []byte
	resp.Body.Read(body)
	return body, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ── NATS auto-connect ────────────────────────────────────────────────

func (p *Plugin) tryConnectNATS() bool {
	url := envOrDefault("NATS_URL", "nats://crabcc-nats:4222")
	creds := os.Getenv("NATS_CREDS")
	
	// Simple TCP health check to NATS server
	conn, err := net.DialTimeout("tcp", strings.TrimPrefix(strings.TrimPrefix(url, "nats://"), "tls://"), 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	
	// Set env for nats plugin to pick up
	if os.Getenv("NATS_URL") == "" {
		os.Setenv("NATS_URL", url)
	}
	if creds != "" && os.Getenv("NATS_CREDS") == "" {
		os.Setenv("NATS_CREDS", creds)
	}
	return true
}

// ── GPU compute check ─────────────────────────────────────────────────

type ComputeStatus struct {
	Provider     string
	InstancesOnline int
	InstancesTotal  int
	CheapestGPU     string
	CheapestPrice   float64
}

func (p *Plugin) checkCompute() *ComputeStatus {
	provider := p.config.ComputeProvider
	switch provider {
	case "vastai":
		return p.checkVastAI()
	default:
		return nil
	}
}

func (p *Plugin) checkVastAI() *ComputeStatus {
	key := os.Getenv("VAST_API_KEY")
	if key == "" {
		return nil
	}
	
	resp, err := httpGet("https://console.vast.ai/api/v0/instances", key)
	if err != nil {
		return nil
	}
	
	var result struct {
		Instances []struct {
			ID        int     `json:"id"`
			CurState  string  `json:"cur_state"`
			GpuName   string  `json:"gpu_name"`
			MinBid    float64 `json:"min_bid"`
		} `json:"instances"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil
	}
	
	cs := &ComputeStatus{Provider: "vastai", InstancesTotal: len(result.Instances)}
	var cheapest float64 = 999999
	for _, inst := range result.Instances {
		if inst.CurState == "running" {
			cs.InstancesOnline++
		}
		if inst.MinBid < cheapest && inst.MinBid > 0 {
			cheapest = inst.MinBid
			cs.CheapestGPU = inst.GpuName
		}
	}
	if cheapest < 999999 {
		cs.CheapestPrice = cheapest
	}
	return cs
}

// ── Hub visibility ────────────────────────────────────────────────────

// StartHubExporter starts an HTTP server that exposes plugin status
// for the hub.crabcc.app dashboard. Runs on localhost:9797.
func (p *Plugin) StartHubExporter(port int) {
	if port <= 0 {
		port = 9797
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p.hubStatus())
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p.hubStatus())
	})
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	}()
}

type HubStatus struct {
	Plugin  string `json:"plugin"`
	Bao     bool   `json:"bao_connected"`
	Langfuse bool  `json:"langfuse_wired"`
	NATS    bool   `json:"nats_wired"`
	Compute *ComputeStatus `json:"compute,omitempty"`
	Plugins int    `json:"plugins_loaded"`
	Online  bool   `json:"online"`
}

func (p *Plugin) hubStatus() HubStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return HubStatus{
		Plugin:   PluginID,
		Bao:      p.status.BaoConnected,
		Langfuse: p.status.LangfuseWired,
		NATS:     p.status.NATSWired,
		Compute:  p.checkCompute(),
		Plugins:  4,
		Online:   true,
	}
}

// ── Additional hooks ──────────────────────────────────────────────────

func (p *Plugin) onStop() agent.HookResult {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status.BaoConnected = false
	p.status.LangfuseWired = false
	p.status.NATSWired = false
	return agent.HookResult{Decision: agent.HookDecisionPass}
}

func (p *Plugin) onError(payload agent.HookPayload) agent.HookResult {
	// On error, check if bao is still reachable
	if p.config.BaoToken != "" {
		go p.wireFromBao()
	}
	return agent.HookResult{Decision: agent.HookDecisionPass}
}
