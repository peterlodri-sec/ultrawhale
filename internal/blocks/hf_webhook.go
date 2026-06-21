package blocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ── HuggingFace Webhook — Liveness Integration ───────────────────────
//
// HF webhook fires on dataset repo updates.
// Vaked liveness: the webhook IS the Testifies layer.
// Every dataset update is evidence that the loop is alive.

// HFWebhookEvent is the HuggingFace webhook payload.
type HFWebhookEvent struct {
	Event    string `json:"event"`    // "repo.update"
	Repo     string `json:"repo"`     // "PeetPedro/ultrawhale-dogfood"
	RepoType string `json:"repoType"` // "dataset"
	URL      string `json:"url"`
	Sender   string `json:"sender"`
	Ref      string `json:"ref"`
}

// HFWebhookHandler receives HF webhook events.
type HFWebhookHandler struct {
	mu      sync.Mutex
	Events  []HFWebhookEvent
	Stats   HFWebhookStats
}

// HFWebhookStats tracks webhook activity.
type HFWebhookStats struct {
	TotalEvents  int64
	RepoUpdates  int64
	LastEvent    string
}

var hfWebhook = &HFWebhookHandler{
	Events: make([]HFWebhookEvent, 0, 64),
}

// HFWebhookReceive handles incoming HF webhook payload.
func HFWebhookReceive(w http.ResponseWriter, r *http.Request) {
	hfWebhook.mu.Lock()
	defer hfWebhook.mu.Unlock()

	var event HFWebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid payload", 400)
		return
	}

	hfWebhook.Events = append(hfWebhook.Events, event)
	hfWebhook.Stats.TotalEvents++
	hfWebhook.Stats.RepoUpdates++
	hfWebhook.Stats.LastEvent = event.Event

	// Vaked liveness: the webhook IS evidence
	Pulse("hf.webhook", fmt.Sprintf("%s: %s", event.Repo, event.Event))

	Log(LogInfo, "hf.webhook", fmt.Sprintf("%s: %s", event.Repo, event.Event),
		"", "", 0, nil)

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{"status": "received", "vaked": "liveness-verified"})
}

// HFWebhookStatus returns compact webhook status.
func HFWebhookStatus() string {
	hfWebhook.mu.Lock()
	defer hfWebhook.mu.Unlock()
	return fmt.Sprintf("hf-webhook: %d events · last: %s",
		hfWebhook.Stats.TotalEvents, hfWebhook.Stats.LastEvent)
}

// HFWebhookVakedFit returns the webhook's Vaked fit.
func HFWebhookVakedFit() string {
	return `HF WEBHOOK = TESTIFIES LAYER (continuous evidence)

  Every dataset update → webhook fires → liveness pulse
  The webhook IS the evidence that the loop is alive.
  Vaked: Testifies layer. The data proves the loop closes.`
}
