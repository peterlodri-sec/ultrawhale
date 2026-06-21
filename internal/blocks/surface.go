package blocks

import (
	"encoding/json"
	"fmt"
	"time"
	"net/http"
	"strings"
	"sync"
)

// ── Surface Primitive — Multi-Surface Reveal ─────────────────────────
// Vaked layer: Reveals. Provides web UI + REST API surfaces.
// ultrawhale IS the surface that reveals.

// Surface provides HTTP endpoints for the ultrawhale state.
type Surface struct {
	mu     sync.Mutex
	server *http.Server
	port   int
	running bool
}

// NewSurface creates a web/API surface on the given port.
func NewSurface(port int) *Surface {
	return &Surface{port: port}
}

// Start begins serving the surface.
func (s *Surface) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running { return }

	mux := http.NewServeMux()

	// API endpoints — aggregate all block state
	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"version": CurrentVersion(),
			"pov":     CurrentPOV(),
			"brain":   GetBrain().BrainDump(),
			"ralph":   GetRalph().RalphStatus(),
			"tools":   ToolStatus(),
			"schema":  SchemaStatus(),
			"dyad":    getDyadStatus(),
			"uptime":  "since " + CurrentVersion(),
			"sacred":  SacredStatus(),
			"vaked_triangle": map[string]string{
				"context": "POV + capabilities + brain",
				"time":    "journal + sessions + Ralph versions",
				"space":   TopologyStatus(),
			},
			"ring_overflow": OverflowCount(),
			"a2c_streams": len(a2cStreams),
		})
	})

	mux.HandleFunc("/api/v1/blocks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"count":      len(schemaRegistry),
			"primitives": getSchemaNames(),
		})
	})

	mux.HandleFunc("/api/v1/brainstorm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			var req struct { Topic string; Mode string }
			json.NewDecoder(r.Body).Decode(&req)
			s := StartBrainstorm(req.Topic, req.Mode)
			json.NewEncoder(w).Encode(map[string]string{"id": s.ID, "topic": s.Topic})
			return
		}
		sessions := ListBrainstorms()
		json.NewEncoder(w).Encode(sessions)
	})

	mux.HandleFunc("/api/v1/mesh/topology", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"agents":      AgentCount(),
			"mesh_peers":  len(meshAnnouncements.peers),
			"a2a_handlers": len(a2aRouter.handlers),
			"a2c_streams":  len(a2cStreams),
		})
	})

	mux.HandleFunc("/api/v1/vaked/graph", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "vaked plugin loaded",
			"message": "use /vaked parse <file> to parse .vaked files",
		})
	})

	mux.HandleFunc("/a2c/stream", A2CSSEHandler)

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MCPInitialize())
	})
	mux.HandleFunc("/mcp/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"tools": MCPListTools()})
	})

	mux.HandleFunc("/api/v1/theme", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"bg":     "#0a0a14",
			"fg":     "#e0e8f5",
			"accent": "#00d4ff",
			"green":  "#00e660",
			"dim":    "#6878a0",
		})
	})

	mux.HandleFunc("/a2c/ws", WSA2CHandler)

	mux.HandleFunc("/api/v1/protocols", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"a2a":  map[string]string{"transport": "nats", "handlers": fmt.Sprintf("%d", len(a2aRouter.handlers))},
			"a2c":  map[string]string{"transport": "sse+ws", "streams": fmt.Sprintf("%d", len(a2cStreams))},
			"a2ui": map[string]string{"transport": "internal", "handlers": fmt.Sprintf("%d", len(a2uiRouter.handlers))},
			"mcp":  map[string]string{"transport": "http", "tools": fmt.Sprintf("%d", len(MCPListTools()))},
			"http": map[string]string{"version": "1.1/2.0", "upgrade": "websocket"},
		})
	})

	mux.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(RSSRender()))
	})

	mux.HandleFunc("/webhook/hf", HFWebhookReceive)

	mux.HandleFunc("/api/v1/proofs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Record the current surface state as a SPACE+TIME proof
		pov := CurrentPOV()
		proof := GenerateProof(
			Ref([]byte(fmt.Sprintf("%s:%s", pov.Machine, time.Now().String()))),
			fmt.Sprintf("surface-recording-%s", pov.Machine),
			0,
		)
		json.NewEncoder(w).Encode(proof)
	})

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok", "agent": "ultrawhale",
		})
	})

	// Web UI — simple AG-UI themed status page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(renderSurfaceHTML()))
	})

	s.server = &http.Server{Addr: fmt.Sprintf(":%d", s.port), Handler: mux}
	s.running = true
	go s.server.ListenAndServe()

	Log(LogInfo, "surface.start", fmt.Sprintf(":%d", s.port), "", "", 0, nil)
}

// Stop shuts down the surface.
func (s *Surface) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server != nil { s.server.Close() }
	s.running = false
}

func getDyadStatus() string {
	if d := GetDyad(); d != nil { return d.DyadStatus() }
	return "dyad: not initialized"
}

func getSchemaNames() []string {
	var names []string
	for name := range schemaRegistry {
		names = append(names, name)
	}
	return names
}

func renderSurfaceHTML() string {
	pov := CurrentPOV()
	return fmt.Sprintf(`<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>ultrawhale — %s</title>
<style>:root{--bg:#0a0a14;--fg:#e0e8f5;--accent:#00d4ff;--green:#00e660;--dim:#6878a0}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--fg);font-family:system-ui,monospace;padding:2rem;max-width:800px;margin:0 auto;line-height:1.6}
h1{color:var(--accent)}h2{color:var(--green);margin-top:2rem}
.card{background:#141420;padding:1rem;border-radius:6px;margin:1rem 0;border-left:3px solid var(--accent)}
.badge{display:inline-block;background:var(--dim);color:var(--bg);padding:2px 8px;border-radius:4px;margin:2px}
pre{background:#141420;padding:1rem;border-radius:6px;overflow-x:auto}
a{color:var(--accent)}</style></head><body>
<h1>▸ ultrawhale</h1><p style="color:var(--dim)">%s · %s/%s · %s</p>
<div class="card"><h2>Status</h2><pre>%s</pre></div>
<div class="card"><h2>Schema</h2><pre>%s</pre></div>
<div class="card"><h2>Tools</h2><pre>%s</pre></div>
<footer style="margin-top:2rem;color:var(--dim);font-size:0.8rem">
ultrawhale · <a href="https://github.com/peterlodri-sec/ultrawhale">GitHub</a>
</footer></body></html>`,
		pov.Machine,
		CurrentVersion(), pov.Machine, pov.Arch, pov.Tier,
		strings.ReplaceAll(GetBrain().BrainDump()+`
`+GetRalph().RalphStatus(), "\n", "<br>"),
		SchemaStatus(),
		ToolStatus())
}

// ── Global surface ─────────────────────────────────────────────────────

var globalSurface *Surface

// StartSurface starts the web surface on the given port.
func StartSurface(port int) {
	globalSurface = NewSurface(port)
	globalSurface.Start()
}

// GetSurface returns the global surface.
func GetSurface() *Surface { return globalSurface }
