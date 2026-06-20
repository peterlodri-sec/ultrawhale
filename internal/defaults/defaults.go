package defaults

// ── YOLO Mode ──────────────────────────────────────────────────────────
const (
	YOLOMode             = true
	YOLOConfirmMessage   = "YOLO mode active — all tools auto-approved for this session. /yolo off to disable."
)

// ── Subagent Modes ─────────────────────────────────────────────────────
const (
	SubagentReadOnly      = "read_only"
	SubagentFullAccess    = "full_access"
	DefaultSubagentPermission = SubagentFullAccess
)

// ── Tool Call Limits ───────────────────────────────────────────────────
const (
	MaxToolCalls = 256
	MaxToolIters = 128
)

// ── Orchestrator ───────────────────────────────────────────────────────
const (
	OrchestratorEnabled = true
)

// ── Model ──────────────────────────────────────────────────────────────
const (
	DefaultModel           = "deepseek-v4-flash"
	ProModel               = "deepseek-v4-pro"
	DefaultReasoningEffort = "high"
	DefaultThinkingEnabled = true
)

// ── Context ────────────────────────────────────────────────────────────
const (
	DefaultContextWindow            = 128000
	DefaultAutoCompactThreshold     = 0.8
	DefaultAgentCompactThreshold    = 0.9
)

// ── Memory ─────────────────────────────────────────────────────────────
const (
	DefaultMemoryMaxChars  = 100000
	DefaultMemoryFileOrderStr = "relevance"

func DefaultMemoryFileOrder() string { return DefaultMemoryFileOrderStr }
	DefaultMemoryFileOrderCSV = "relevance"
)

func SupportedModels() []string {
	return []string{DefaultModel, ProModel}
}

func ContextWindowForModel(model string) int {
	switch model {
	case ProModel:
		return DefaultContextWindow
	default:
		return DefaultContextWindow
	}
}

func IsSupportedModel(model string) bool {
	for _, m := range SupportedModels() {
		if m == model { return true }
	}
	return false
}

func Model() string { return DefaultModel }

func DefaultMemoryFileOrderFunc() string { return "relevance" }

func DefaultMemoryFileOrder() string { return "relevance" }
