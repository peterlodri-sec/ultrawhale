package plugins

import (
	"context"
	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/plugins/superpowers"
)

type superpowersPlugin struct{ inner *superpowers.Plugin }

func (sp *superpowersPlugin) Manifest() Manifest {
	return Manifest{
		ID: superpowers.PluginID, Name: "Superpowers", Version: "0.1.0", Official: true,
		Description: "Auto-discovers and wires vaked infrastructure.",
		Capabilities: []Capability{CapabilityHooks},
		Permissions:  []Permission{},
	}
}
func (sp *superpowersPlugin) Hooks(ctx Context) []agent.HookHandler { return sp.inner.Hooks() }
func (sp *superpowersPlugin) Doctor(c context.Context, ctx Context) []Diagnostic {
	return []Diagnostic{{PluginID: superpowers.PluginID, Level: DiagnosticOK, Label: "superpowers", Detail: sp.inner.Doctor()}}
}
