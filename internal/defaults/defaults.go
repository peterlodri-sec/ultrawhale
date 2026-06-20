package defaults

// ── YOLO Mode ──────────────────────────────────────────────────────────
// One-time user confirmation on TUI start, then full auto for the session.

const (
	// YOLOMode: when true, user confirms ONCE at TUI start, then all tools auto-approved.
	YOLOMode = true

	// YOLOConfirmMessage shown on first prompt.
	YOLOConfirmMessage = "YOLO mode active — all tools auto-approved for this session. /yolo off to disable."
)

// ── Subagent Modes ─────────────────────────────────────────────────────
// Only two modes: read-only and full-access-auto.

const (
	SubagentReadOnly    = "read_only"
	SubagentFullAccess  = "full_access" // auto-approve all tools
)

// Default subagent permission profile.
const DefaultSubagentPermission = SubagentFullAccess

// ── Tool Call Limits ───────────────────────────────────────────────────

const (
	DefaultMaxToolCalls = 256
	DefaultMaxToolIters = 128
)

// ── Orchestrator ───────────────────────────────────────────────────────

const (
	OrchestratorEnabled = true // always delegate prompts to subagents
)

const DefaultModel = "deepseek-v4-flash"

const DefaultThinkingEnabled = true
