package plugins

import (
	"context"
	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/plugins/agentfield"
)

type agentfieldPlugin struct{ inner *agentfield.Plugin }

func (p *agentfieldPlugin) Manifest() Manifest {
	return Manifest{
		ID: agentfield.PluginID, Name: "AgentField", Version: "0.1.0", Official: true,
		Description: "Local-first AgentField control plane with DID identity and callable API.",
		Capabilities: []Capability{CapabilityHooks},
	}
}
func (p *agentfieldPlugin) Hooks(ctx Context) []agent.HookHandler { return p.inner.Hooks() }
func (p *agentfieldPlugin) Doctor(c context.Context, ctx Context) []Diagnostic {
	if p.inner == nil { return nil }
	return []Diagnostic{{PluginID: agentfield.PluginID, Level: DiagnosticOK, Label: "agentfield", Detail: p.inner.Doctor()}}
}
