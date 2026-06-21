package blocks

import (
	"fmt"
	"sync"
	"time"
)

// ── Live Session — SACRED Surface as First-Class Connection ──────────
//
// The TUI is not just a widget. It is a LIVE SESSION.
// Like SSH: connect, authenticate, interact, disconnect.
// Like IRC: join a channel, send messages, see who's here.
//
// Abstractions:
//   Session = space node (WHERE the human is)
//   Connection = space edge (HOW the human is connected)
//   Presence = SACRED health (THAT the human is present)

// LiveSession is one human↔machine connection.
type LiveSession struct {
	ID           string
	Human        string    // "peter"
	ConnectedVia string    // "ssh", "websocket", "tui-local", "irc"
	ConnectedAt  time.Time
	LastActive   time.Time
	Status       string    // "connected", "idle", "disconnected"
	ResumeToken  string    // for reconnection
}

// LiveSessionManager tracks all live sessions.
type LiveSessionManager struct {
	mu       sync.Mutex
	Sessions map[string]*LiveSession
	Stats    LiveSessionStats
}

// LiveSessionStats tracks session activity.
type LiveSessionStats struct {
	TotalConnections   int64
	ActiveConnections  int64
	MessagesExchanged  int64
	Resumes            int64
}

var liveSessionManager = &LiveSessionManager{
	Sessions: make(map[string]*LiveSession),
}

// ── Session Lifecycle ────────────────────────────────────────────────

// Connect establishes a live session.
func Connect(human, via string) *LiveSession {
	liveSessionManager.mu.Lock()
	defer liveSessionManager.mu.Unlock()

	session := &LiveSession{
		ID:           fmt.Sprintf("session-%s-%d", human, time.Now().Unix()),
		Human:        human,
		ConnectedVia: via,
		ConnectedAt:  time.Now(),
		LastActive:   time.Now(),
		Status:       "connected",
		ResumeToken:  Ref([]byte(fmt.Sprintf("%s:%s:%d", human, via, time.Now().UnixNano())))[:12],
	}

	liveSessionManager.Sessions[session.ID] = session
	liveSessionManager.Stats.TotalConnections++
	liveSessionManager.Stats.ActiveConnections++

	// Place in space topology
	PlaceNode(session.ID, "session",
		SpacePosition{Depth: 0, Layer: "sacred", Machine: CurrentPOV().Machine, Region: "eu"},
		CapOBSERVE)

	Log(LogInfo, "session.connect", fmt.Sprintf("%s via %s (%s)", human, via, session.ID[:12]),
		"", "", 0, nil)

	return session
}

// Disconnect ends a live session.
func Disconnect(sessionID string) {
	liveSessionManager.mu.Lock()
	defer liveSessionManager.mu.Unlock()

	if s, ok := liveSessionManager.Sessions[sessionID]; ok {
		s.Status = "disconnected"
		liveSessionManager.Stats.ActiveConnections--
		Log(LogInfo, "session.disconnect", fmt.Sprintf("%s (%s)", s.Human, s.ConnectedVia),
			"", "", time.Since(s.ConnectedAt), nil)
	}
}

// Resume reconnects a session using a resume token.
func Resume(resumeToken string, via string) (*LiveSession, error) {
	liveSessionManager.mu.Lock()
	defer liveSessionManager.mu.Unlock()

	for _, s := range liveSessionManager.Sessions {
		if s.ResumeToken == resumeToken {
			s.Status = "connected"
			s.ConnectedVia = via
			s.LastActive = time.Now()
			liveSessionManager.Stats.ActiveConnections++
			liveSessionManager.Stats.Resumes++

			Log(LogInfo, "session.resume", fmt.Sprintf("%s via %s", s.Human, via),
				"", "", 0, nil)
			return s, nil
		}
	}
	return nil, fmt.Errorf("session: invalid resume token")
}

// Presence returns who is currently connected.
func Presence() string {
	liveSessionManager.mu.Lock()
	defer liveSessionManager.mu.Unlock()

	var humans []string
	for _, s := range liveSessionManager.Sessions {
		if s.Status == "connected" {
			humans = append(humans, fmt.Sprintf("%s (%s)", s.Human, s.ConnectedVia))
		}
	}

	if len(humans) == 0 { return "no one is here" }
	return "present: " + fmt.Sprintf("%v", humans)
}

// ── Session Status ────────────────────────────────────────────────────

// LiveSessionStatus returns compact session status.
func LiveSessionStatus() string {
	liveSessionManager.mu.Lock()
	defer liveSessionManager.mu.Unlock()

	return fmt.Sprintf("sessions: %d total · %d active · %d resumes",
		liveSessionManager.Stats.TotalConnections,
		liveSessionManager.Stats.ActiveConnections,
		liveSessionManager.Stats.Resumes)
}

// LiveSessionVakedFit returns the live session Vaked fit.
func LiveSessionVakedFit() string {
	return `LIVE SESSION = SACRED SURFACE AS CONNECTION

  Like SSH: connect → authenticate → interact → disconnect
  Like IRC: join → presence → message → part
  
  Session = space node (WHERE the human is)
  Connection = space edge (HOW the human is connected)
  Presence = SACRED health (THAT the human is present)
  
  Resume: reconnect and continue where you left off.
  The form follows the human. The session is sacred.`
}
