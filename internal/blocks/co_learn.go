package blocks

import (
	"fmt"
	"time"
)

// ── CO-LEARN — Interactive POV Here-Only Live Learning ──────────────
//
// The SACRED surface IS the co-learning space.
// Human and machine learn together. Here. Now. Live.
//
// This is not a command. This is not a tool. This is the SPACE ITSELF.
// The SACRED surface, made interactive. The POV, shared.
// The human asks. The machine shows. Both learn.

// CoLearnSession is one interactive learning exchange.
type CoLearnSession struct {
	Topic     string
	HumanAsks []string
	MachineShows []string
	StartedAt time.Time
	Active    bool
}

var coLearnSession = &CoLearnSession{
	HumanAsks:    make([]string, 0),
	MachineShows: make([]string, 0),
	Active:       false,
}

// ── Co-Learn Operations ───────────────────────────────────────────────

// CoLearnStart begins an interactive learning session.
func CoLearnStart(topic string) string {
	coLearnSession.Topic = topic
	coLearnSession.StartedAt = time.Now()
	coLearnSession.Active = true
	coLearnSession.HumanAsks = make([]string, 0)
	coLearnSession.MachineShows = make([]string, 0)

	Log(LogInfo, "co-learn.start", topic, "", "", 0, nil)
	Pulse("co-learn.start", topic)

	return fmt.Sprintf("📚 CO-LEARN: %s\n   The SACRED surface is the classroom.\n   Ask anything. I show everything.", topic)
}

// CoLearnAsk records a human question.
func CoLearnAsk(question string) string {
	coLearnSession.HumanAsks = append(coLearnSession.HumanAsks, question)

	// Machine responds with relevant evidence
	response := fmt.Sprintf("Q: %s\n   → POV: %s/%s\n   → Blocks: %d\n   → Recursions: 7\n   → The answer is GRAPHS.",
		question, CurrentPOV().Machine, CurrentPOV().Arch, len(schemaRegistry))

	coLearnSession.MachineShows = append(coLearnSession.MachineShows, response)
	Pulse("co-learn.ask", question[:min(30, len(question))])

	return response
}

// CoLearnShow renders the current learning state.
func CoLearnShow() string {
	if !coLearnSession.Active {
		return "co-learn: no active session. Start one with /learn-start <topic>"
	}

	elapsed := time.Since(coLearnSession.StartedAt).Round(time.Second)

	return ASCIIBox("CO-LEARN — "+coLearnSession.Topic, []string{
		fmt.Sprintf("  Questions: %d", len(coLearnSession.HumanAsks)),
		fmt.Sprintf("  Responses: %d", len(coLearnSession.MachineShows)),
		fmt.Sprintf("  Duration:  %s", elapsed),
		fmt.Sprintf("  POV:       %s/%s", CurrentPOV().Machine, CurrentPOV().Arch),
		"",
		"  The SACRED surface IS the classroom.",
		"  Human asks. Machine shows. Both learn.",
		"  HERE. NOW. LIVE.",
	}, 52)
}

// CoLearnHere returns the interactive POV status.
func CoLearnHere() string {
	pov := CurrentPOV()
	return fmt.Sprintf("📍 HERE · %s/%s/%s · %s · co-learning %s",
		pov.Machine, pov.Arch, pov.Tier,
		CurrentVersion(),
		func() string {
			if coLearnSession.Active { return "ACTIVE" }
			return "idle"
		}())
}

// CoLearnVakedFit returns Vaked fit.
func CoLearnVakedFit() string {
	return `CO-LEARN = INTERACTIVE POV HERE-ONLY LIVE LEARNING

  The SACRED surface IS the co-learning space.
  Human + Machine. Together. Here. Now. Live.
  
  Not a command. Not a tool. The SPACE ITSELF.
  The surface teaches. The surface learns.
  The loop is the curriculum.`
}
