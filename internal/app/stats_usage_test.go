package app

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/usewhale/whale/internal/telemetry"
)

func TestSessionUsageSummaryReportsSubagentShapeDriftOnlyWhenHashesChange(t *testing.T) {
	dir := t.TempDir()
	usagePath := filepath.Join(dir, "usage.jsonl")
	parentID := "parent-session"

	writeUsageRecord(t, usagePath, telemetry.UsageRecord{
		Session:          parentID,
		Model:            "deepseek-v4-flash",
		PromptTokens:     1000,
		CompletionTokens: 100,
		PromptCacheHit:   800,
		PromptCacheMiss:  200,
	})
	writeUsageRecord(t, usagePath, telemetry.UsageRecord{
		Session:          "child-1",
		Model:            "deepseek-v4-flash",
		Kind:             "subagent",
		ParentSessionID:  parentID,
		PromptTokens:     2000,
		CompletionTokens: 200,
		PromptCacheHit:   1500,
		PromptCacheMiss:  500,
		CacheShape:       &telemetry.CacheShape{RequestHash: "request-a", SystemHash: "system-a", ToolsHash: "tools-a"},
	})
	writeUsageRecord(t, usagePath, telemetry.UsageRecord{
		Session:          "child-2",
		Model:            "deepseek-v4-flash",
		Kind:             "subagent",
		ParentSessionID:  parentID,
		PromptTokens:     3000,
		CompletionTokens: 300,
		PromptCacheHit:   2500,
		PromptCacheMiss:  500,
		CacheShape:       &telemetry.CacheShape{RequestHash: "request-a", SystemHash: "system-a", ToolsHash: "tools-a"},
	})

	stable := formatSessionUsageSummary(readSessionUsageSummary(usagePath, parentID))
	if strings.Contains(stable, "subagent shape drift") {
		t.Fatalf("stable subagent shapes should not report drift: %s", stable)
	}

	writeUsageRecord(t, usagePath, telemetry.UsageRecord{
		Session:          "child-3",
		Model:            "deepseek-v4-flash",
		Kind:             "subagent",
		ParentSessionID:  parentID,
		PromptTokens:     4000,
		CompletionTokens: 400,
		PromptCacheHit:   2500,
		PromptCacheMiss:  1500,
		CacheShape:       &telemetry.CacheShape{RequestHash: "request-b", SystemHash: "system-b", ToolsHash: "tools-a"},
	})

	drift := formatSessionUsageSummary(readSessionUsageSummary(usagePath, parentID))
	for _, want := range []string{
		"subagents 3 turns",
		"subagent shape drift 2 request/2 system",
	} {
		if !strings.Contains(drift, want) {
			t.Fatalf("summary missing %q:\n%s", want, drift)
		}
	}
	if strings.Contains(drift, "tools") {
		t.Fatalf("stable tools hash should not be reported as drift: %s", drift)
	}
}
