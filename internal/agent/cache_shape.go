package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/usewhale/whale/internal/compact"
	"github.com/usewhale/whale/internal/core"
	"github.com/usewhale/whale/internal/telemetry"
)

const cacheShapeTailMessages = 8

type cacheShapeMessage struct {
	Role             core.Role            `json:"role"`
	Text             string               `json:"text,omitempty"`
	ReasoningContent string               `json:"reasoning_content,omitempty"`
	ToolCallID       string               `json:"tool_call_id,omitempty"`
	ToolCalls        []cacheShapeToolCall `json:"tool_calls,omitempty"`
}

type cacheShapeToolSpec struct {
	Type     string                 `json:"type"`
	Function cacheShapeToolFunction `json:"function"`
}

type cacheShapeToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type cacheShapeToolCall struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

type pendingCacheShapeToolCall struct {
	ID                   string
	ConsumesStoredResult bool
}

func buildCacheShape(history []core.Message, tools []core.Tool, assistantPrefix string) *telemetry.CacheShape {
	var system []cacheShapeMessage
	var log []cacheShapeMessage
	var pendingToolCalls []pendingCacheShapeToolCall
	providerMessageIndex := 0
	flushPending := func() []cacheShapeMessage {
		if len(pendingToolCalls) == 0 {
			return nil
		}
		out := make([]cacheShapeMessage, 0, len(pendingToolCalls))
		for _, pending := range pendingToolCalls {
			out = append(out, syntheticMissingToolResultShape(pending.ID))
		}
		pendingToolCalls = nil
		return out
	}
	for _, msg := range history {
		shaped := shapeMessage(msg, &pendingToolCalls, providerMessageIndex, flushPending)
		if len(shaped) == 0 {
			continue
		}
		if msg.Role == core.RoleSystem {
			system = append(system, shaped...)
		} else {
			log = append(log, shaped...)
		}
		providerMessageIndex += len(shaped)
	}
	if flushed := flushPending(); len(flushed) > 0 {
		log = append(log, flushed...)
		providerMessageIndex += len(flushed)
	}

	tailLen := min(cacheShapeTailMessages, len(log))
	head := log[:len(log)-tailLen]
	tail := log[len(log)-tailLen:]
	shape := &telemetry.CacheShape{
		SystemHash:   hashJSON(system),
		ToolsHash:    hashJSON(shapeToolPayload(tools)),
		LogMessages:  len(log),
		TailMessages: tailLen,
	}
	if strings.TrimSpace(assistantPrefix) != "" {
		shape.AssistantPrefixHash = hashJSON(assistantPrefix)
	}
	if len(head) > 0 {
		shape.LogHeadHash = hashJSON(head)
	}
	if len(tail) > 0 {
		shape.LogTailHash = hashJSON(tail)
	}
	shape.RequestHash = hashJSON(struct {
		SystemHash          string `json:"system_hash,omitempty"`
		ToolsHash           string `json:"tools_hash,omitempty"`
		FewShotHash         string `json:"fewshot_hash,omitempty"`
		AssistantPrefixHash string `json:"assistant_prefix_hash,omitempty"`
		LogHeadHash         string `json:"log_head_hash,omitempty"`
		LogTailHash         string `json:"log_tail_hash,omitempty"`
		LogMessages         int    `json:"log_messages,omitempty"`
		TailMessages        int    `json:"tail_messages,omitempty"`
	}{
		SystemHash:          shape.SystemHash,
		ToolsHash:           shape.ToolsHash,
		FewShotHash:         shape.FewShotHash,
		AssistantPrefixHash: shape.AssistantPrefixHash,
		LogHeadHash:         shape.LogHeadHash,
		LogTailHash:         shape.LogTailHash,
		LogMessages:         shape.LogMessages,
		TailMessages:        shape.TailMessages,
	})
	return shape
}

func shapeMessage(msg core.Message, pendingToolCalls *[]pendingCacheShapeToolCall, providerMessageIndex int, flushPending func() []cacheShapeMessage) []cacheShapeMessage {
	switch msg.Role {
	case core.RoleSystem:
		out := flushPending()
		return append(out, cacheShapeMessage{Role: core.RoleSystem, Text: msg.Text})
	case core.RoleUser:
		out := flushPending()
		return append(out, cacheShapeMessage{Role: core.RoleUser, Text: msg.Text})
	case core.RoleAssistant:
		out := flushPending()
		shaped := cacheShapeMessage{
			Role:      core.RoleAssistant,
			Text:      msg.Text,
			ToolCalls: shapeToolCalls(msg.ToolCalls, providerMessageIndex),
		}
		if len(msg.ToolCalls) > 0 {
			shaped.ReasoningContent = msg.Reasoning
		}
		for callIdx, tc := range msg.ToolCalls {
			*pendingToolCalls = append(*pendingToolCalls, pendingCacheShapeToolCall{
				ID:                   cacheShapeToolCallID(tc.ID, providerMessageIndex, callIdx),
				ConsumesStoredResult: strings.TrimSpace(tc.ID) != "",
			})
		}
		return append(out, shaped)
	case core.RoleTool:
		return shapeToolResults(msg.ToolResults, pendingToolCalls)
	default:
		return nil
	}
}

func shapeToolCalls(calls []core.ToolCall, providerMessageIndex int) []cacheShapeToolCall {
	if len(calls) == 0 {
		return nil
	}
	out := make([]cacheShapeToolCall, 0, len(calls))
	for callIdx, call := range calls {
		out = append(out, cacheShapeToolCall{
			ID:        cacheShapeToolCallID(call.ID, providerMessageIndex, callIdx),
			Name:      call.Name,
			Arguments: call.Input,
		})
	}
	return out
}

func cacheShapeToolCallID(id string, providerMessageIndex, callIdx int) string {
	if strings.TrimSpace(id) == "" {
		return fmt.Sprintf("whale_synthetic_call_%d_%d", providerMessageIndex, callIdx)
	}
	return id
}

func shapeToolResults(results []core.ToolResult, pendingToolCalls *[]pendingCacheShapeToolCall) []cacheShapeMessage {
	if len(results) == 0 || len(*pendingToolCalls) == 0 {
		return nil
	}
	out := make([]cacheShapeMessage, 0, len(results))
	for _, result := range results {
		for len(*pendingToolCalls) > 0 && !(*pendingToolCalls)[0].ConsumesStoredResult {
			out = append(out, syntheticMissingToolResultShape((*pendingToolCalls)[0].ID))
			*pendingToolCalls = (*pendingToolCalls)[1:]
		}
		if len(*pendingToolCalls) == 0 || strings.TrimSpace(result.ToolCallID) == "" {
			continue
		}
		match := -1
		for i, pending := range *pendingToolCalls {
			if pending.ConsumesStoredResult && pending.ID == result.ToolCallID {
				match = i
				break
			}
		}
		if match < 0 {
			continue
		}
		for i := 0; i < match; i++ {
			out = append(out, syntheticMissingToolResultShape((*pendingToolCalls)[i].ID))
		}
		id := (*pendingToolCalls)[match].ID
		out = append(out, cacheShapeMessage{
			Role:       core.RoleTool,
			ToolCallID: id,
			Text:       compact.ToolResultReplayContent(result.Content),
		})
		*pendingToolCalls = (*pendingToolCalls)[match+1:]
	}
	return out
}

func syntheticMissingToolResultShape(id string) cacheShapeMessage {
	return cacheShapeMessage{
		Role:       core.RoleTool,
		ToolCallID: id,
		Text:       `{"success":false,"error":"missing tool result recovered before provider send","code":"missing_tool_result_recovered"}`,
	}
}

func shapeToolPayload(tools []core.Tool) []cacheShapeToolSpec {
	if len(tools) == 0 {
		return nil
	}
	out := make([]cacheShapeToolSpec, 0, len(tools))
	for _, tool := range tools {
		spec := core.DescribeTool(tool)
		out = append(out, cacheShapeToolSpec{
			Type: "function",
			Function: cacheShapeToolFunction{
				Name:        spec.Name,
				Description: spec.Description,
				Parameters:  stableMap(core.FlattenSchemaForModel(spec.Parameters)),
			},
		})
	}
	return out
}

func stableMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = stableValue(v)
	}
	return out
}

func stableValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return stableMap(x)
	case []any:
		out := make([]any, 0, len(x))
		for _, item := range x {
			out = append(out, stableValue(item))
		}
		return out
	default:
		return x
	}
}

func hashJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte("null")
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:16]
}
