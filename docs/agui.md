# AG-UI — Agent-UI Theme Engine

AG-UI renders content blocks with themed chrome in the TUI. Three themes, Perlin-noise shader background, FPS-limited at 60fps.

## Themes

| Theme | Background | Accent | Vibe |
|-------|-----------|--------|------|
| Dense Matrix Green | `#040804` | `#00e660` | Neon terminal |
| Clean Graph Cyberpunk | `#0a0a14` | `#00d4ff` | Cyberpunk |
| Tactical Graveyard | `#141414` | `#b0b0b0` | Minimal grayscale |

## Keybindings

| Key | Action |
|-----|--------|
| `Ctrl+Shift+T` | Cycle themes |
| `/reload theme dense` | Direct switch |
| `Ctrl+Shift+B` | Shader background toggle |
| `Ctrl+Shift+Z` | Zen mode |
| `Ctrl+Shift+S` | Sidebar toggle |

## ChatBlock API

```go
type ChatBlock struct {
    Type    BlockType
    Title   string
    Content string
    Width   int
}

block := agui.NewChatBlock("tool_call", "grep", "found 12 matches", 80)
output := block.Render()
```

## Block types

| Type | Icon | Accent | Use |
|------|------|--------|-----|
| BlockThinking | ⏳ | Dim | Chain-of-thought |
| BlockToolCall | 🔧 | Accent | Tool execution |
| BlockToolResult | 📋 | Fg | Tool output |
| BlockCodeDiff | Δ | `#00e660` | Diff rendering |
| BlockPlanCard | 📐 | `#00d4ff` | Plan display |
| BlockFileTree | 📁 | Accent | File listing |

## Shader

Animated Perlin-noise background using Unicode block chars (░▒▓█). Zero allocations after init. FPS limiter at 60fps. Toggle with `Ctrl+Shift+B`.
