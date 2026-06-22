package blocks

import (
	"fmt"
	"net/http"
	"time"
)

// ── LOCAL MODEL — Self-Hosted via Ollama ─────────────────────────────
//
// Default: qwen3.5:35b on Tailnet (M3 Max)
// Fallback: OpenRouter free pool (8 models)
// The local model is ALWAYS tried first. If unreachable, fallback.

type LocalModel struct {
	Name      string // "qwen3.5:35b"
	Provider  string // "ollama"
	Endpoint  string // "http://m3-max:11434"
	Available bool
	Latency   time.Duration
}

var localModel = &LocalModel{
	Name:     "qwen3.5:35b",
	Provider: "ollama",
	Endpoint: "http://m3-max.tailnet:11434" // M3-ONLY — never download here,
}

// PingLocalModel checks if the local model is reachable.
func PingLocalModel() bool {
	start := time.Now()
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/api/tags", localModel.Endpoint))
	if err != nil {
		localModel.Available = false
		return false
	}
	defer resp.Body.Close()
	localModel.Available = resp.StatusCode == 200
	localModel.Latency = time.Since(start)
	return localModel.Available
}

// LocalModelStatus returns the local model status line.
func LocalModelStatus() string {
	ping := PingLocalModel()
	icon := "🔴"
	latency := "—"
	if ping {
		icon = "🟢"
		latency = localModel.Latency.Round(time.Millisecond).String()
	}

	return fmt.Sprintf("%s %s · %s · %s", icon, localModel.Name, localModel.Endpoint, latency)
}
