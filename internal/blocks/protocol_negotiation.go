package blocks

import (
	"fmt"
	"strings"
)

// ── Protocol Negotiation ──────────────────────────────────────────────
// v22: Auto-negotiate the best protocol for each connection.
// Client sends Accept header, server selects best match.

// NegotiatedProtocol is the result of protocol negotiation.
type NegotiatedProtocol struct {
	Protocol string // "http/1.1", "h2", "websocket", "h3"
	Version  string // protocol version
	Upgraded bool   // true if the connection was upgraded
}

// NegotiateProtocol selects the best protocol based on client capabilities.
func NegotiateProtocol(acceptHeader, connectionHeader, upgradeHeader string) NegotiatedProtocol {
	accept := strings.ToLower(acceptHeader)
	connection := strings.ToLower(connectionHeader)
	upgrade := strings.ToLower(upgradeHeader)

	// WebSocket upgrade
	if strings.Contains(connection, "upgrade") && strings.Contains(upgrade, "websocket") {
		return NegotiatedProtocol{Protocol: "websocket", Version: "13", Upgraded: true}
	}

	// HTTP/2 (

}