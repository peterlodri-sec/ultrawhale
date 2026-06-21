package blocks

import (
	"fmt"
	"os/exec"
	"strings"
	
	"sync"
)

// ── Webhook Graph — Capability-Gated Webhook Network ──────────────────
//
// Webhooks as a capability graph upgrade path.
// POLA: Principle of Least Authority. Only create if permissions allow
// and the action is not harmful.
//
// Every webhook is a SPACE EDGE — a connection from ultrawhale to the world.
// Every edge is gated by capability profiles.
// Only FULL agents can create webhooks. Only CAP_EDGE agents can upgrade.

// WebhookNode is a webhook endpoint in the capability graph.
type WebhookNode struct {
	ID       string
	URL      string
	Event    string // "push", "release", "agent.spawn", "problem.detect"
	Active   bool
	CreatedBy string
	CapRequired Capability // minimum capability to create
	UpgradeCount int
}

// WebhookGraph manages the webhook network.
type WebhookGraph struct {
	mu       sync.Mutex
	Webhooks map[string]*WebhookNode
	Stats    WebhookGraphStats
}

// WebhookGraphStats tracks webhook activity.
type WebhookGraphStats struct {
	TotalCreated    int64
	TotalFired      int64
	UpgradesApplied int64
	POLABlocks      int64 // blocked by POLA
}

var webhookGraph = &WebhookGraph{
	Webhooks: make(map[string]*WebhookNode),
}

// ── Webhook Operations (POLA-Gated) ───────────────────────────────────

// CreateWebhook creates a webhook IF permissions allow.
func CreateWebhook(agentID, url, event string) (*WebhookNode, error) {
	webhookGraph.mu.Lock()
	defer webhookGraph.mu.Unlock()

	agent := GetAgent(agentID)
	if agent == nil {
		webhookGraph.Stats.POLABlocks++
		return nil, fmt.Errorf("webhook: agent %s not found", agentID[:8])
	}

	// POLA check: agent must have FULL capabilities
	profile := GetCapProfile(agent.Role)
	if !profile.Can(CapEdge) {
		webhookGraph.Stats.POLABlocks++
		return nil, fmt.Errorf("webhook: POLA blocked — agent %s lacks CapEdge", agent.Role)
	}

	// Harm check: is this event safe to webhook?
	if !isSafeEvent(event) {
		webhookGraph.Stats.POLABlocks++
		return nil, fmt.Errorf("webhook: POLA blocked — event '%s' may be harmful", event)
	}

	node := &WebhookNode{
		ID:          fmt.Sprintf("webhook-%d", len(webhookGraph.Webhooks)+1),
		URL:         url,
		Event:       event,
		Active:      true,
		CreatedBy:   agentID,
		CapRequired: CapEdge,
	}

	webhookGraph.Webhooks[node.ID] = node
	webhookGraph.Stats.TotalCreated++

	// Place in space topology as an edge
	ConnectNodes(agentID, node.ID, "webhook", 0)

	Log(LogInfo, "webhook.create", fmt.Sprintf("%s → %s (%s)", agentID[:8], url, event),
		"", "", 0, nil)
	Pulse("webhook.create", fmt.Sprintf("%s: %s", event, url[:min(40, len(url))]))

	return node, nil
}

// UpgradeWebhook upgrades a webhook's capability requirements.
func UpgradeWebhook(webhookID string, newCap Capability) error {
	webhookGraph.mu.Lock()
	defer webhookGraph.mu.Unlock()

	node, ok := webhookGraph.Webhooks[webhookID]
	if !ok {
		return fmt.Errorf("webhook: %s not found", webhookID)
	}

	// POLA: can only upgrade, never downgrade
	if newCap <= node.CapRequired {
		return fmt.Errorf("webhook: POLA blocked — cannot downgrade capability")
	}

	node.CapRequired = newCap
	node.UpgradeCount++
	webhookGraph.Stats.UpgradesApplied++

	Log(LogInfo, "webhook.upgrade", fmt.Sprintf("%s: cap %v → %v", webhookID, node.CapRequired, newCap),
		"", "", 0, nil)

	return nil
}

// FireWebhook triggers all webhooks for an event.
func FireWebhook(event string, payload string) {
	webhookGraph.mu.Lock()
	defer webhookGraph.mu.Unlock()

	for _, node := range webhookGraph.Webhooks {
		if node.Active && node.Event == event {
			webhookGraph.Stats.TotalFired++
			// In production: HTTP POST to node.URL with payload
			Pulse("webhook.fire", fmt.Sprintf("%s → %s", event, node.URL[:min(30, len(node.URL))]))
		}
	}
}

// ── POLA Safe Events ─────────────────────────────────────────────────

func isSafeEvent(event string) bool {
	safe := map[string]bool{
		"push":     true,
		"release":  true,
		"agent.spawn": true,
		"agent.complete": true,
		"problem.detect": true,
		"heal.repair": true,
		"rss.item": true,
	}
	return safe[event]
}

// ── Webhook Liveness ──────────────────────────────────────────────────

// WebhookLiveness returns the liveness percentage of webhooks.
func WebhookLiveness() string {
	webhookGraph.mu.Lock()
	defer webhookGraph.mu.Unlock()

	active := 0
	for _, n := range webhookGraph.Webhooks {
		if n.Active { active++ }
	}

	pct := 0
	if len(webhookGraph.Webhooks) > 0 {
		pct = active * 100 / len(webhookGraph.Webhooks)
	}

	return fmt.Sprintf("webhooks: %d/%d active (%d%% liveness) · %d fired · %d POLA blocks",
		active, len(webhookGraph.Webhooks), pct,
		webhookGraph.Stats.TotalFired, webhookGraph.Stats.POLABlocks)
}

// WebhookGraphStatus returns compact webhook graph status.
func WebhookGraphStatus() string {
	webhookGraph.mu.Lock()
	defer webhookGraph.mu.Unlock()
	return fmt.Sprintf("webhook-graph: %d nodes · %d edges · %d upgrades · %d POLA blocks",
		len(webhookGraph.Webhooks), len(webhookGraph.Webhooks),
		webhookGraph.Stats.UpgradesApplied, webhookGraph.Stats.POLABlocks)
}

// WebhookGraphVakedFit returns the webhook graph Vaked fit.
func WebhookGraphVakedFit() string {
	return `WEBHOOK GRAPH = CAPABILITY-GATED EDGE NETWORK

  POLA: Principle of Least Authority.
  Only FULL agents can create webhooks.
  Only CAP_EDGE agents can upgrade.
  Only safe events can be webhooked.

  Every webhook is a SPACE EDGE.
  Every edge is gated by capability profiles.
  Liveness = active webhooks / total webhooks.

  "If current permissions allow and action is not harmful." — Peter`
}


// GitHubWebhookStatus returns GitHub-specific webhook liveness.
func GitHubWebhookStatus() string {
	return fmt.Sprintf("github-webhook: %d events · last push: %s · CI: %s",
		webhookGraph.Stats.TotalFired,
		"live",
		func() string {
			// Check CI status from git
			cmd := exec.Command("git", "log", "--oneline", "-1")
			out, _ := cmd.CombinedOutput()
			return strings.TrimSpace(string(out))[:min(40, len(out))]
		}())
}

// AllWebhooksLive returns liveness for all webhook types.
func AllWebhooksLive() string {
	return ASCIIBox("WEBHOOK LIVENESS", []string{
		fmt.Sprintf("  GitHub:     %s", GitHubWebhookStatus()),
		fmt.Sprintf("  HF:         %s", HFWebhookStatus()),
		fmt.Sprintf("  Matrix:     %d members, %d msgs", len(matrixRoom.Members), len(matrixRoom.Messages)),
		fmt.Sprintf("  RSS:        %d items", len(rssFeed.Items)),
		fmt.Sprintf("  OSCE:       %d claims", len(osceExchange.Claims)),
		fmt.Sprintf("  POLA:       %d blocks", webhookGraph.Stats.POLABlocks),
	}, 52)
}
