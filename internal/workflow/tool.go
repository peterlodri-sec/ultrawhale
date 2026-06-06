package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/usewhale/whale/internal/core"
)

type Tool struct {
	runner               *ScriptRunner
	parentSessionIDFunc  func() string
	keywordTriggerEnable bool
}

func NewTool(runner *ScriptRunner, parentSessionIDFunc ...func() string) Tool {
	var fn func() string
	if len(parentSessionIDFunc) > 0 {
		fn = parentSessionIDFunc[0]
	}
	return Tool{runner: runner, parentSessionIDFunc: fn, keywordTriggerEnable: true}
}

type ToolOptions struct {
	ParentSessionIDFunc   func() string
	KeywordTriggerEnabled bool
}

func NewToolWithOptions(runner *ScriptRunner, opts ToolOptions) Tool {
	return Tool{runner: runner, parentSessionIDFunc: opts.ParentSessionIDFunc, keywordTriggerEnable: opts.KeywordTriggerEnabled}
}

func (t Tool) Name() string { return "workflow" }

func (t Tool) Description() string {
	return strings.Join([]string{
		"Launch a restricted Whale workflow script asynchronously for decomposable multi-agent work such as fan-out research, repository inspection, or multi-perspective review.",
		t.workflowUseGuidance(),
		"When the user asks what workflows are available, call this workflow tool with action=\"list\". Do not inspect .whale/workflows, search files, or run shell commands to discover workflow names.",
		"When the user asks whether a named workflow exists, call this workflow tool with action=\"resolve\" and name. The resolver returns project, user, and builtin matches plus available names when missing.",
		"When the user clearly asks to run a named workflow, call this workflow tool directly with name. Do not call request_user_input or ask a chat question for launch confirmation first; this tool returns the single TUI launch confirmation when confirmation is required. Do not first inspect files, search the workspace, or block confirmation because you think an expected input might be missing unless the user asked for a preflight check.",
		"When the user clearly asks to create, generate, or write a new workflow, do not inspect existing workflow directories or load skills first. Generate a Claude Code-compatible raw JavaScript workflow script, pass it as script, and set saveAs to the same kebab-case value as meta.name. The tool will request confirmation; if the user confirms, Whale saves it under the project .whale/workflows directory before launching it.",
		"Use ordinary tools instead for a single quick read, edit, shell-dependent task, or answer.",
		"When an available named workflow fits, pass name instead of generating a new script; include args only when the user supplied useful input or the workflow contract clearly requires it. Do not ask for a missing args value merely because the args field exists. Use scriptPath for an existing file; generate script only for an explicit ad-hoc workflow with no matching named workflow.",
		"Workflow scripts are not Node scripts: export const meta must be a pure literal first statement; phases must be objects like { title: 'Review', detail: '...' }; meta/args/budget/phase/log/agent/workflow/parallel/pipeline are runtime globals; host APIs like require/process/fetch/Date.now/Math.random/new Date are unavailable.",
		"Use phase('Name') only as a statement. Do not call phase('Name', async () => ...); phase() is not a wrapper and returns nothing.",
		"Await async workflow primitives before reading their results: const result = await agent(...), await parallel(...), await pipeline(...), or await workflow(...). Inside parallel(), thunks may return agent(...).",
		"Use agent(prompt, { label, phase, schema, max_tool_calls?, agent?, tools?, disallowedTools?, effort?, permissionMode?, maxTurns? }). The first argument is the full prompt; put labels in opts.label. Do not pass opts.system, opts.prompt, or opts.structured. Do not set opts.model unless the user explicitly asks for a provider-supported model; otherwise let Whale use the current model.",
		"For Claude Code compatibility, do not use Whale-only workflow APIs or fields in generated scripts. Use standard JSON Schema for structured output; enum-only schema properties must also include type: \"string\".",
		"End generated workflows by returning a final JSON-serializable result, usually the synthesis/report object.",
		"Use parallel() with thunks, not promises: () => agent(...). Give every agent() a short unique label, include enough context in each prompt, use JSON Schema for structured output, and add a synthesis/verification agent when combining branches.",
		"Workflow agent leaves are tool-scoped workers. Use agent definitions and opts.tools/opts.disallowedTools to state required tool selectors. Supported selectors include workspace.read, workspace.write, shell.read, shell.run, web.search, web.fetch, mcp.read, and exact tool names; shell.run or workspace.write require an explicit non-read-only permissionMode. If a needed selector is not exposed by the runtime, make the workflow report the missing evidence instead of assuming shell, edit, or host access.",
		"Returns an async launch receipt; tell the user only that /workflows opens the workflow panel. Do not mention /workflows with run ids or hidden subcommands.",
	}, " ")
}

func (t Tool) workflowUseGuidance() string {
	if t.keywordTriggerEnable {
		return "Use this when the user explicitly asks for a workflow, fan-out, multi-agent orchestration, or names/describes an available workflow from the system prompt catalog."
	}
	return "Use this only when the user explicitly asks Whale to run or create a workflow by name or script. Do not infer workflow use from broad task descriptions, ordinary release tasks, or the presence of the word workflow."
}

func (t Tool) Parameters() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"list", "resolve", "run"},
				"description": "Optional workflow control action. Use list to discover available workflows, resolve to check a named workflow without launch confirmation, and run to launch. Omit for backward-compatible run behavior.",
			},
			"script": map[string]any{
				"type":        "string",
				"maxLength":   MaxWorkflowScriptBytes,
				"description": "Self-contained workflow script beginning with a pure literal export const meta = {...}. Use phase(), log(), agent(), workflow(), parallel(thunks), pipeline(), args, and budget; every agent should have a short label.",
			},
			"saveAs": map[string]any{
				"type":        "string",
				"description": "Optional kebab-case workflow name used only when the user asks to create a new workflow. Requires script, must match meta.name, saves to project .whale/workflows/<name>.js, then launches that named workflow.",
			},
			"scriptPath": map[string]any{
				"type":        "string",
				"description": "Path to a workflow script on disk. Takes precedence over script.",
			},
			"args": map[string]any{
				"description": "Optional JSON-serializable args exposed to the script as read-only args. Omit this field when the user did not provide workflow input and the workflow contract does not clearly require it. May be a string, object, array, number, boolean, or null depending on the workflow contract.",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Named workflow from project or user .whale/workflows. Used only when scriptPath and script are omitted.",
			},
			"resumeFromRunId": map[string]any{
				"type":        "string",
				"description": "Optional source run id for same-session resume. Unchanged agent() calls reuse cached results; the first changed call and later calls rerun.",
			},
			"budgetTokens": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"description": "Optional completion-token budget shared by this workflow and child workflows. agent() calls are blocked once spent completion tokens reach the cap.",
			},
		},
	}
}

func (t Tool) Run(ctx context.Context, call core.ToolCall) (core.ToolResult, error) {
	if t.runner == nil {
		return workflowToolError(call, "not_configured", "workflow runner is not configured")
	}
	var input WorkflowInput
	if err := json.Unmarshal([]byte(call.Input), &input); err != nil {
		return workflowToolError(call, "invalid_input", err.Error())
	}
	switch workflowToolAction(input.Action) {
	case "list":
		return t.listWorkflows(ctx, call)
	case "resolve":
		return t.resolveWorkflow(ctx, call, input)
	case "run":
	default:
		return workflowToolError(call, "invalid_input", "workflow action must be one of list, resolve, or run")
	}
	if strings.TrimSpace(input.SaveAs) != "" {
		prepared, err := t.prepareGenerated(ctx, input)
		if err != nil {
			return workflowToolError(call, "workflow_save_failed", err.Error())
		}
		data := workflowConfirmationData(prepared, workflowToolArgsActionString(input.Args), input.ResumeFromRunID)
		data["workflowScript"] = prepared.Script
		data["workflowSaveAs"] = prepared.Definition.Name
		content, err := core.MarshalToolEnvelope(core.ToolEnvelope{
			OK:      true,
			Success: true,
			Code:    "workflow_confirmation_required",
			Summary: fmt.Sprintf("Workflow %q requires user confirmation before save and launch.", prepared.Definition.Name),
			Data:    data,
		})
		if err != nil {
			return core.ToolResult{}, err
		}
		return core.ToolResult{ToolCallID: call.ID, Name: call.Name, Content: content, Metadata: workflowConfirmationMetadata(data)}, nil
	}
	if strings.TrimSpace(input.Script) != "" {
		if err := validateWorkflowScriptForConfirmation(input.Script); err != nil {
			return workflowToolError(call, "workflow_save_failed", err.Error())
		}
		return workflowToolError(call, "workflow_confirmation_required", "workflow scripts must be saved as a named workflow before launch confirmation")
	}
	if strings.TrimSpace(input.ScriptPath) != "" {
		if strings.TrimSpace(input.Name) != "" || strings.TrimSpace(input.Script) != "" {
			return workflowToolError(call, "invalid_input", "scriptPath cannot be combined with name or script")
		}
		resolved, err := ResolveScriptPath(ctx, input.ScriptPath)
		if err != nil {
			return workflowToolError(call, "workflow_failed", err.Error())
		}
		data := workflowConfirmationData(resolved, workflowToolArgsActionString(input.Args), input.ResumeFromRunID)
		data["workflowScriptPath"] = resolved.Definition.Path
		content, err := core.MarshalToolEnvelope(core.ToolEnvelope{
			OK:      true,
			Success: true,
			Code:    "workflow_confirmation_required",
			Summary: fmt.Sprintf("Workflow %q requires user confirmation before launch.", resolved.Definition.Name),
			Data:    data,
		})
		if err != nil {
			return core.ToolResult{}, err
		}
		return core.ToolResult{ToolCallID: call.ID, Name: call.Name, Content: content, Metadata: workflowConfirmationMetadata(data)}, nil
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return workflowToolError(call, "invalid_input", "workflow name is required")
	}
	resolved, err := t.resolveNamedWorkflow(ctx, name)
	if err != nil {
		return workflowToolError(call, "workflow_failed", err.Error())
	}
	data := workflowConfirmationData(resolved, workflowToolArgsActionString(input.Args), input.ResumeFromRunID)
	content, err := core.MarshalToolEnvelope(core.ToolEnvelope{
		OK:      true,
		Success: true,
		Code:    "workflow_confirmation_required",
		Summary: fmt.Sprintf("Workflow %q requires user confirmation before launch.", resolved.Definition.Name),
		Data:    data,
	})
	if err != nil {
		return core.ToolResult{}, err
	}
	return core.ToolResult{ToolCallID: call.ID, Name: call.Name, Content: content, Metadata: workflowConfirmationMetadata(data)}, nil
}

func (t Tool) prepareGenerated(ctx context.Context, input WorkflowInput) (ResolvedScript, error) {
	if t.runner == nil || t.runner.Library == nil {
		return ResolvedScript{}, errors.New("workflow library is not configured")
	}
	if strings.TrimSpace(input.Script) == "" {
		return ResolvedScript{}, errors.New("saveAs requires script")
	}
	if strings.TrimSpace(input.Name) != "" || strings.TrimSpace(input.ScriptPath) != "" {
		return ResolvedScript{}, errors.New("saveAs cannot be combined with name or scriptPath")
	}
	return t.runner.Library.PrepareGenerated(ctx, input.Script, input.SaveAs)
}

func validateWorkflowScriptForConfirmation(script string) error {
	parsed, err := parseWorkflowScript(script)
	if err != nil {
		return err
	}
	if err := validateWorkflowCompile(parsed.Executable); err != nil {
		return err
	}
	return validateGeneratedWorkflowScript(parsed.Executable)
}

func (t Tool) resolveNamedWorkflow(ctx context.Context, name string) (ResolvedScript, error) {
	if t.runner == nil || t.runner.Library == nil {
		return ResolvedScript{}, errors.New("workflow library is not configured")
	}
	return t.runner.Library.Resolve(ctx, name)
}

func workflowToolAction(action string) string {
	action = strings.TrimSpace(strings.ToLower(action))
	if action == "" {
		return "run"
	}
	return action
}

func (t Tool) listWorkflows(ctx context.Context, call core.ToolCall) (core.ToolResult, error) {
	if t.runner == nil || t.runner.Library == nil {
		return workflowToolError(call, "not_configured", "workflow library is not configured")
	}
	defs, err := t.runner.Library.List(ctx)
	if err != nil {
		return workflowToolError(call, "workflow_list_failed", err.Error())
	}
	data := workflowDiscoveryData(t.runner.Library, defs)
	content, err := core.MarshalToolEnvelope(core.ToolEnvelope{
		OK:      true,
		Success: true,
		Code:    "workflow_list",
		Summary: workflowListSummary(defs),
		Data:    data,
	})
	if err != nil {
		return core.ToolResult{}, err
	}
	return core.ToolResult{ToolCallID: call.ID, Name: call.Name, Content: content}, nil
}

func (t Tool) resolveWorkflow(ctx context.Context, call core.ToolCall, input WorkflowInput) (core.ToolResult, error) {
	if t.runner == nil || t.runner.Library == nil {
		return workflowToolError(call, "not_configured", "workflow library is not configured")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return workflowToolError(call, "invalid_input", "workflow name is required for action=resolve")
	}
	defs, err := t.runner.Library.List(ctx)
	if err != nil {
		return workflowToolError(call, "workflow_resolve_failed", err.Error())
	}
	data := workflowDiscoveryData(t.runner.Library, defs)
	data["query"] = name
	for _, def := range defs {
		if def.Name != name {
			continue
		}
		if def.Status != DefinitionReady {
			data["workflow"] = workflowDefinitionData(def)
			return workflowToolErrorWithData(call, "workflow_problem", def.Error, data)
		}
		data["workflow"] = workflowDefinitionData(def)
		content, err := core.MarshalToolEnvelope(core.ToolEnvelope{
			OK:      true,
			Success: true,
			Code:    "workflow_resolved",
			Summary: fmt.Sprintf("Workflow %q resolved from %s.", def.Name, workflowNonEmpty(def.Source, "workflow library")),
			Data:    data,
		})
		if err != nil {
			return core.ToolResult{}, err
		}
		return core.ToolResult{ToolCallID: call.ID, Name: call.Name, Content: content}, nil
	}
	return workflowToolErrorWithData(call, "workflow_not_found", workflowNotFoundMessage(name, defs), data)
}

func (t Tool) parentSessionID() string {
	if t.parentSessionIDFunc == nil {
		return ""
	}
	return strings.TrimSpace(t.parentSessionIDFunc())
}

func workflowDiscoveryData(library *Library, defs []Definition) map[string]any {
	ready := make([]map[string]any, 0, len(defs))
	problems := make([]map[string]any, 0)
	available := make([]string, 0, len(defs))
	for _, def := range defs {
		item := workflowDefinitionData(def)
		if def.Status == DefinitionReady {
			ready = append(ready, item)
			if strings.TrimSpace(def.Name) != "" {
				available = append(available, def.Name)
			}
			continue
		}
		problems = append(problems, item)
	}
	return map[string]any{
		"workflows": ready,
		"problems":  problems,
		"available": available,
		"roots":     workflowRootData(library),
		"count":     len(ready),
	}
}

func workflowDefinitionData(def Definition) map[string]any {
	item := map[string]any{
		"name":   def.Name,
		"source": def.Source,
		"status": string(def.Status),
	}
	if desc := strings.TrimSpace(def.Description); desc != "" {
		item["description"] = desc
	}
	if when := strings.TrimSpace(def.WhenToUse); when != "" {
		item["whenToUse"] = when
	}
	if path := strings.TrimSpace(def.Path); path != "" {
		item["path"] = path
	}
	if root := strings.TrimSpace(def.Root); root != "" {
		item["root"] = root
	}
	if len(def.Phases) > 0 {
		phases := make([]map[string]any, 0, len(def.Phases))
		for _, phase := range def.Phases {
			p := map[string]any{"title": strings.TrimSpace(phase.Title)}
			if detail := strings.TrimSpace(phase.Detail); detail != "" {
				p["detail"] = detail
			}
			phases = append(phases, p)
		}
		item["phases"] = phases
	}
	if def.EstimatedAgents > 0 {
		item["estimatedAgents"] = def.EstimatedAgents
	}
	if def.DefaultBudgetTokens > 0 {
		item["defaultBudgetTokens"] = def.DefaultBudgetTokens
	}
	if err := strings.TrimSpace(def.Error); err != "" {
		item["error"] = err
	}
	return item
}

func workflowRootData(library *Library) []map[string]any {
	if library == nil {
		return nil
	}
	roots := make([]map[string]any, 0, len(library.Roots))
	for _, root := range library.Roots {
		item := map[string]any{
			"source": root.Source,
			"path":   root.Path,
			"rank":   root.Rank,
		}
		exists, status := workflowRootStatus(root.Path)
		item["exists"] = exists
		item["status"] = status
		roots = append(roots, item)
	}
	return roots
}

func workflowRootStatus(path string) (bool, string) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return true, "present"
		}
		return true, "not_directory"
	}
	if os.IsNotExist(err) {
		return false, "missing"
	}
	return false, "error: " + err.Error()
}

func workflowListSummary(defs []Definition) string {
	ready, problems := 0, 0
	for _, def := range defs {
		if def.Status == DefinitionReady {
			ready++
		} else {
			problems++
		}
	}
	if problems > 0 {
		return fmt.Sprintf("%d workflow(s) available; %d workflow definition(s) have problems.", ready, problems)
	}
	return fmt.Sprintf("%d workflow(s) available.", ready)
}

func workflowNotFoundMessage(name string, defs []Definition) string {
	available := make([]string, 0, len(defs))
	for _, def := range defs {
		if def.Status == DefinitionReady && strings.TrimSpace(def.Name) != "" {
			available = append(available, def.Name)
		}
	}
	if len(available) == 0 {
		return fmt.Sprintf("workflow not found: %s; no workflows are currently available", name)
	}
	return fmt.Sprintf("workflow not found: %s; available workflows: %s", name, strings.Join(available, ", "))
}

func workflowNonEmpty(v, fallback string) string {
	if v = strings.TrimSpace(v); v != "" {
		return v
	}
	return fallback
}

func workflowConfirmationData(resolved ResolvedScript, args, resume string) map[string]any {
	data := map[string]any{
		"confirmationRequired": true,
		"workflowName":         resolved.Definition.Name,
		"workflowArgs":         args,
		"userGuidance":         "Tell the user a workflow confirmation has been shown. Do not say the workflow has started until the user confirms it.",
	}
	if description := strings.TrimSpace(resolved.Definition.Description); description != "" {
		data["description"] = description
	}
	if path := strings.TrimSpace(resolved.Definition.Path); path != "" {
		data["scriptPath"] = path
	}
	if resume = strings.TrimSpace(resume); resume != "" {
		data["workflowResume"] = resume
	}
	return data
}

func workflowToolArgsActionString(args any) string {
	switch v := args.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func workflowConfirmationMetadata(data map[string]any) map[string]any {
	meta := map[string]any{
		"workflow_confirmation_required": true,
		"abort_turn_after_tool_result":   true,
	}
	for _, key := range []string{"workflowName", "workflowArgs", "workflowResume", "scriptPath", "workflowSaveAs", "workflowScriptPath"} {
		if v, ok := data[key]; ok {
			meta[key] = v
		}
	}
	return meta
}

func workflowOutputData(out WorkflowOutput) map[string]any {
	data := map[string]any{}
	if out.Status != "" {
		data["status"] = out.Status
	}
	if out.TaskID != "" {
		data["taskId"] = out.TaskID
	}
	if out.RunID != "" {
		data["runId"] = string(out.RunID)
	}
	if out.Summary != "" {
		data["summary"] = out.Summary
	}
	if out.TranscriptDir != "" {
		data["transcriptDir"] = out.TranscriptDir
	}
	if out.ScriptPath != "" {
		data["scriptPath"] = out.ScriptPath
	}
	if out.SessionURL != "" {
		data["sessionUrl"] = out.SessionURL
	}
	if out.Warning != "" {
		data["warning"] = out.Warning
	}
	if out.Error != "" {
		data["error"] = out.Error
	}
	if out.RunID != "" {
		data["userGuidance"] = "Tell the user /workflows opens the workflow panel. Do not suggest /workflows with a run id, events, or cancel subcommands."
	}
	return data
}

func workflowToolMetadata(out WorkflowOutput) map[string]any {
	meta := map[string]any{}
	if out.RunID != "" {
		meta["workflow_run_id"] = string(out.RunID)
	}
	if out.Status != "" {
		meta["workflow_status"] = out.Status
	}
	if out.ScriptPath != "" {
		meta["workflow_script_path"] = out.ScriptPath
	}
	return meta
}

func workflowToolError(call core.ToolCall, code, msg string) (core.ToolResult, error) {
	return workflowToolErrorWithData(call, code, msg, nil)
}

func workflowToolErrorWithData(call core.ToolCall, code, msg string, data map[string]any) (core.ToolResult, error) {
	env := core.NewToolErrorEnvelope(code, msg)
	env.Data = data
	content, err := core.MarshalToolEnvelope(env)
	if err != nil {
		return core.ToolResult{}, err
	}
	return core.ToolResult{ToolCallID: call.ID, Name: call.Name, Content: content, IsError: true}, nil
}
