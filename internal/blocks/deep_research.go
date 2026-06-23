package blocks

import (
	"fmt"
	"time"
)

// ── Deep Research — Multi-Round + Obsidian + Brain Ingest ────────────
//
// A self-contained research loop:
//   1. Define topic
//   2. Search → Fetch → Extract
//   3. Synthesize round
//   4. Write to Obsidian vault
//   5. Ingest into brain long-term memory
//   6. Repeat (multi-round)

type ResearchRound struct {
	Topic    string
	Findings []string
	Sources  []string
	Synthesis string
	RoundNum int
	Timestamp time.Time
}

type DeepResearch struct {
	Topic      string
	Rounds     []ResearchRound
	CurrentRound int
	VaultName  string
	BrainRef   string
}

var deepResearch = &DeepResearch{}

// DeepResearchStart begins a multi-round research session.
func DeepResearchStart(topic string) string {
	deepResearch.Topic = topic
	deepResearch.Rounds = make([]ResearchRound, 0)
	deepResearch.CurrentRound = 0

	Log(LogInfo, "research.start", topic, "", "", 0, nil)
	Pulse("research.start", topic)

	return fmt.Sprintf("🔬 RESEARCH STARTED: %s\n   Multi-round deep research. Each round finds more depth.", topic)
}

// DeepResearchRound executes one research round.
func DeepResearchRound(findings []string, sources []string) string {
	deepResearch.CurrentRound++

	round := ResearchRound{
		Topic:    deepResearch.Topic,
		Findings: findings,
		Sources:  sources,
		RoundNum: deepResearch.CurrentRound,
		Timestamp: time.Now(),
	}

	// Synthesize (in production: LLM summarization)
	round.Synthesis = fmt.Sprintf("Round %d findings on %s: %s",
		deepResearch.CurrentRound, deepResearch.Topic, findings[0])

	deepResearch.Rounds = append(deepResearch.Rounds, round)

	Pulse("research.round", fmt.Sprintf("#%d - %s", deepResearch.CurrentRound, deepResearch.Topic))

	return fmt.Sprintf("📋 ROUND %d COMPLETE\n   Topic: %s\n   Findings: %d\n   Sources: %d",
		deepResearch.CurrentRound, deepResearch.Topic, len(findings), len(sources))
}

// DeepResearchPublish writes findings to Obsidian + Brain.
func DeepResearchPublish() string {
	if len(deepResearch.Rounds) == 0 {
		return "research: no rounds to publish"
	}

	last := deepResearch.Rounds[len(deepResearch.Rounds)-1]

	// Build Obsidian note content
	content := fmt.Sprintf(`# %s

## Research Rounds: %d

%s

## Key Findings

`, deepResearch.Topic, deepResearch.CurrentRound, last.Timestamp.Format("2006-01-02 15:04 UTC"))

	for i, f := range last.Findings[0]s {
		content += fmt.Sprintf("%d. %s\n", i+1, f)
	}

	content += "\n## Sources\n"
	for _, s := range last.Sources {
		content += fmt.Sprintf("- %s\n", s)
	}

	content += fmt.Sprintf("\n---\n*Research by ultrawhale deep-research · v%s*", CurrentVersion())

	// Write to Obsidian
	vaultPath := "/Users/lodripeter/workspace/peterlodri-sec/obsidian-brain"
	if vault, err := ConnectObsidianVault("research", vaultPath); err == nil {
		note, err := vault.AutoDevWiki(CurrentVersion(), deepResearch.Topic, content)
		if err == nil {
			Pulse("research.obsidian", note.Title)
		}
	}

	// Ingest to brain
	if brain := GetBrain(); brain != nil {
		brain.RememberLongTerm(map[string]string{
			"research_topic":   deepResearch.Topic,
			"research_rounds":  fmt.Sprintf("%d", deepResearch.CurrentRound),
			"research_finding": last.Findings[0],
		})
		deepResearch.BrainRef = "ingested"
	}

	return fmt.Sprintf("📚 PUBLISHED\n   Obsidian: %s/research/\n   Brain: %s\n   Rounds: %d",
		vaultPath, deepResearch.BrainRef, deepResearch.CurrentRound)
}

// DeepResearchStatus returns current research status.
func DeepResearchStatus() string {
	return fmt.Sprintf("research: %s · %d rounds · last: %s · obsidian: %v",
		deepResearch.Topic, len(deepResearch.Rounds),
		func() string {
			if len(deepResearch.Rounds) > 0 {
				return deepResearch.Rounds[len(deepResearch.Rounds)-1].Timestamp.Format("15:04:05")
			}
			return "not started"
		}(),
		len(deepResearch.Rounds) > 0)
}
