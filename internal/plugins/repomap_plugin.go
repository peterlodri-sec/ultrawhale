package plugins

import (
	"context"
	"github.com/usewhale/whale/internal/agent"
	"github.com/usewhale/whale/internal/repomap"
)

type repomapPlugin struct {
	inner *repomap.Plugin
}

func (rp *repomapPlugin) Manifest() Manifest {
	return Manifest{
		ID:          repomap.PluginID,
		Name:        "Repo Map",
		Version:     "0.1.0",
		Description: "Builds a dependency graph of the workspace and injects it into the LLM context.",
		Official:    true,
		Capabilities: []Capability{
			CapabilityStartupContext,
			CapabilityHooks,
		},
		Permissions: []Permission{
			PermissionReadWorkspace,
		},
	}
}

func (rp *repomapPlugin) StartupContext(c context.Context, ctx Context) (string, error) {
	rp.inner.SetRoot(ctx.WorkspaceRoot)
	return rp.inner.StartupContext(c, ctx.WorkspaceRoot)
}

func (rp *repomapPlugin) Hooks(ctx Context) []agent.HookHandler {
	return rp.inner.Hooks()
}

func (rp *repomapPlugin) Doctor(c context.Context, ctx Context) []Diagnostic {
	return []Diagnostic{{
		PluginID: repomap.PluginID,
		Level:    DiagnosticOK,
		Label:    "repo map",
		Detail:   rp.inner.Doctor(),
	}}
}
