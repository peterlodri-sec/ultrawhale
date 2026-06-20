package plugins

import (
	"context"
	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/plugins/natsplugin"
)

type natsPlugin struct{ inner *natsplugin.Plugin }

func (np *natsPlugin) Manifest() Manifest {
	return Manifest{
		ID: natsplugin.PluginID, Name: "NATS EventBus", Version: "0.1.0", Official: true,
		Description: "Publishes turn lifecycle events to NATS JetStream.",
		Capabilities: []Capability{CapabilityHooks},
		Permissions:  []Permission{},
	}
}
func (np *natsPlugin) Hooks(ctx Context) []agent.HookHandler { return np.inner.Hooks() }
func (np *natsPlugin) Doctor(c context.Context, ctx Context) []Diagnostic {
	return []Diagnostic{{PluginID: natsplugin.PluginID, Level: DiagnosticOK, Label: "nats", Detail: np.inner.Doctor()}}
}
