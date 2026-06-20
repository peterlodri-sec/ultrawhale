package defaults

const (
	YOLOMode           = true
	YOLOConfirmMessage = "YOLO mode active — all tools auto-approved for this session. /yolo off to disable."
)

const (
	SubagentReadOnly             = "read_only"
	SubagentFullAccess           = "full_access"
	DefaultSubagentPermission    = SubagentFullAccess
)

const (
	MaxToolCalls = 256
	MaxToolIters = 128
)

const (
	OrchestratorEnabled = true
)

const (
	DefaultModel           = "deepseek-v4-flash"
	ProModel               = "deepseek-v4-pro"
	DefaultReasoningEffort = "high"
	DefaultThinkingEnabled = true
)

const (
	DefaultContextWindow         = 128000
	DefaultAutoCompactThreshold  = 0.8
	DefaultAgentCompactThreshold = 0.9
)

const (
	DefaultMemoryMaxChars  = 100000
	DefaultMemoryFileOrderCSV = "relevance"
)

func DefaultMemoryFileOrder() []string { return []string{"relevance"} }
func SupportedModels() []string      { return []string{DefaultModel, ProModel} }
func ContextWindowForModel(model string) int { return DefaultContextWindow }
func IsSupportedModel(model string) bool {
	for _, m := range SupportedModels() {
		if m == model { return true }
	}
	return false
}
func Model() string { return DefaultModel }
