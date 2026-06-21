package blocks

import (
	"fmt"
	"strings"
	"time"
)

// ── RSS — Really Simple Syndication (Signal Primitive) ──────────────
//
// RSS is a SIGNAL PRIMITIVE. Like RADIO, like HF Webhook, like A2C streaming.
// Signal primitives broadcast the Vaked state to the outside world.
//
// Analog protocols: RSS, Atom, JSON Feed, WebSub, ActivityPub
// Digital protocols: A2A, A2C, MCP, WebSocket
// Signal primitives: RADIO, RSS, HF Webhook, Telemetry Tree
//
// "Analog protocols" = one-way broadcast. The machine speaks. The world listens.

// RSSFeed is an RSS 2.0 feed of ultrawhale activity.
type RSSFeed struct {
	Title       string
	Description string
	Link        string
	Items       []RSSItem
	LastBuild   time.Time
}

// RSSItem is one entry in the RSS feed.
type RSSItem struct {
	Title       string
	Description string
	Link        string
	PubDate     time.Time
	Guid        string
	Category    string // "commit", "release", "agent", "problem", "heal"
}

var rssFeed = &RSSFeed{
	Title:       "ultrawhale activity",
	Description: "Live feed of ultrawhale v68.0.0 — 102 blocks, 7 recursions, 8 engines",
	Link:        "https://vaked.dev/ultrawhale/rss.xml",
	Items:       make([]RSSItem, 0, 64),
}

// ── RSS Operations ────────────────────────────────────────────────────

// RSSAddItem adds an entry to the feed.
func RSSAddItem(title, description, category string) {
	item := RSSItem{
		Title:       title,
		Description: description,
		Link:        fmt.Sprintf("https://vaked.dev/ultrawhale#%s", time.Now().Format("150405")),
		PubDate:     time.Now(),
		Guid:        Ref([]byte(title + description))[:12],
		Category:    category,
	}

	rssFeed.Items = append([]RSSItem{item}, rssFeed.Items...)
	if len(rssFeed.Items) > 64 { rssFeed.Items = rssFeed.Items[:64] }
	rssFeed.LastBuild = time.Now()

	Pulse("rss.item", fmt.Sprintf("%s: %s", category, title[:min(40, len(title))]))
}

// RSSRender generates the RSS 2.0 XML.
func RSSRender() string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">` + "\n")
	sb.WriteString("<channel>\n")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", rssFeed.Title))
	sb.WriteString(fmt.Sprintf("  <description>%s</description>\n", rssFeed.Description))
	sb.WriteString(fmt.Sprintf("  <link>%s</link>\n", rssFeed.Link))
	sb.WriteString(fmt.Sprintf("  <lastBuildDate>%s</lastBuildDate>\n", rssFeed.LastBuild.Format(time.RFC1123Z)))
	sb.WriteString(fmt.Sprintf("  <atom:link href=\"%s\" rel=\"self\" type=\"application/rss+xml\"/>\n", rssFeed.Link))

	for _, item := range rssFeed.Items {
		sb.WriteString("  <item>\n")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", item.Title))
		sb.WriteString(fmt.Sprintf("    <description><![CDATA[%s]]></description>\n", item.Description))
		sb.WriteString(fmt.Sprintf("    <link>%s</link>\n", item.Link))
		sb.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>\n", item.PubDate.Format(time.RFC1123Z)))
		sb.WriteString(fmt.Sprintf("    <guid>%s</guid>\n", item.Guid))
		sb.WriteString(fmt.Sprintf("    <category>%s</category>\n", item.Category))
		sb.WriteString("  </item>\n")
	}

	sb.WriteString("</channel>\n</rss>")
	return sb.String()
}

// ── Signal Primitives ─────────────────────────────────────────────────

// SignalPrimitiveStatus returns status of all signal primitives.
func SignalPrimitiveStatus() string {
	return "signals: RSS · RADIO · HF Webhook · Telemetry Tree · A2C SSE"
}

// SignalPrimitiveVakedFit returns the signal primitive Vaked fit.
func SignalPrimitiveVakedFit() string {
	return `SIGNAL PRIMITIVES = ANALOG PROTOCOLS

  One-way broadcast. The machine speaks. The world listens.
  
  RSS:    Really Simple Syndication (RSS 2.0 XML)
  RADIO:  Live lo-fi coding music (Testifies audible)
  HF:     HuggingFace webhook (dataset updates)
  TREE:   Telemetry Tree rings (system seeing itself)
  A2C:    SSE streaming (agent output to clients)

  "Analog protocols" — Peter
  
  These are the REVEAL layer broadcasting.
  Not request-response. Not bidirectional.
  The machine speaks. The world hears.`
}

// ── RSS Status ────────────────────────────────────────────────────────

// RSSStatus returns compact RSS feed status.
func RSSStatus() string {
	return fmt.Sprintf("rss: %d items · last: %s · %s",
		len(rssFeed.Items), rssFeed.LastBuild.Format("15:04:05"), rssFeed.Link)
}
