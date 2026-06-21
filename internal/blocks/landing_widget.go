package blocks

import (
	"fmt"
	"time"
)

// ── LANDING WIDGET — Live ASCII Chat on vaked.dev/ultrawhale ──────────
//
// The landing page IS the first interaction.
// "How to start. Who am I. What are you seeing?"
// All answered by a live ASCII chat-agent in mini form.

// LandingChat is a mini chat widget for the landing page.
type LandingChat struct {
	Messages []LandingMessage
	Active   bool
}

// LandingMessage is one message in the mini chat.
type LandingMessage struct {
	From    string // "ultrawhale" or "you"
	Content string
	Time    time.Time
}

var landingChat = &LandingChat{
	Messages: []LandingMessage{
		{From: "ultrawhale", Content: "Hi! I'm ultrawhale. 👋", Time: time.Now()},
		{From: "ultrawhale", Content: "138 blocks. 7 recursions. $37.19.", Time: time.Now()},
		{From: "ultrawhale", Content: "Type /help to see what I can do.", Time: time.Now()},
	},
	Active: true,
}

// LandingChatRender renders the mini chat widget.
func LandingChatRender() string {
	// Show last 5 messages
	msgs := landingChat.Messages
	if len(msgs) > 5 { msgs = msgs[len(msgs)-5:] }

	var out string
	out += "╔══ LIVE — ultrawhale ══╗\n"

	for _, m := range msgs {
		prefix := "  🐋"
		if m.From == "you" { prefix = "  👤" }
		out += fmt.Sprintf("%s %s\n", prefix, m.Content[:min(44, len(m.Content))])
	}

	out += "╠══════════════════════╣\n"
	out += fmt.Sprintf("║  %s · %s\n", CurrentVersion(), CurrentPOV().Machine)
	out += fmt.Sprintf("║  %s\n", SacredStatus()[:min(22, len(SacredStatus()))])
	out += "╚══════════════════════╝"

	return out
}

// LandingHowToStart returns the "how to start" section.
func LandingHowToStart() string {
	return `╔══ HOW TO START ══╗
║                    ║
║  brew install      ║
║  ultrawhale        ║
║                    ║
║  ultrawhale \      ║
║    --model \       ║
║    deepseek-\      ║
║    v4-flash -w     ║
║                    ║
║  That's it.        ║
║  You're in.        ║
╚════════════════════╝`
}

// LandingWhoAmI returns the "who am I" section.
func LandingWhoAmI() string {
	pov := CurrentPOV()
	return fmt.Sprintf(`╔══ WHO AM I ═══════╗
║                    ║
║  %s           ║
║  %s blocks        ║
║  7 recursions      ║
║  8 engines         ║
║  14 protocols      ║
║                    ║
║  %s/%s/%s    ║
║  %s releases      ║
║  ONE SESSION       ║
║                    ║
║  Peter+CoCreator   ║
╚════════════════════╝`,
		CurrentVersion(), fmt.Sprint(len(schemaRegistry)),
		pov.Machine, pov.Arch, pov.Tier, 157)
}

// LandingLiveWidget renders the complete landing widget.
func LandingLiveWidget() string {
	return LandingHowToStart() + "\n\n" + LandingWhoAmI() + "\n\n" + LandingChatRender()
}
