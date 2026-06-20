package plugins

import (
	"context"
	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/plugins/langfuseplugin"
)

type langfusePlugin struct{ inner *langfuseplugin.Plugin }

func (lp *langfusePlugin) Manifest() Manifest {
	return Manifest{
		ID: langfuseplugin.PluginID, Name: "Langfuse Telemetry", Version: "0.1.0", Official: true,
		Description: "Sends LLM observability traces to Langfuse.",
		Capabilities: []Capability{CapabilityHooks},
		Permissions:  []Permission{},
	}
}
func (lp *langfusePlugin) Hooks(ctx Context) []agent.HookHandler { return lp.inner.Hooks() }
func (lp *langfusePlugin) Doctor(c context.Context, ctx Context) []Diagnostic {
	return []Diagnostic{{PluginID: langfuseplugin.PluginID, Level: DiagnosticOK, Label: "langfuse", Detail: lp.inner.Doctor()}}
}
