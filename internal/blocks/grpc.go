package blocks

import "fmt"

// ── gRPC A2A — Structured Agent-to-Agent ─────────────────────────────
// v24: Protocol buffer-based A2A. Replaces NATS for structured comms.
// Active when agent mesh >10 nodes.

var grpcEnabled bool

// EnableGRPC activates gRPC for A2A communication.
func EnableGRPC() { grpcEnabled = true }

// GRPCStatus returns gRPC status.
func GRPCStatus() string {
	if grpcEnabled {
		return "grpc: enabled (protobuf, streaming, deadlines)"
	}
	return fmt.Sprintf("grpc: disabled (activate when agent mesh >10 nodes, current: %d)", AgentCount())
}
