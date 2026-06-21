# Native Format Rendering — AG-UI Protocol Extension

> CoCreator Design, v37.0.0

## Native Formats (v38 MVP)

| Format | TUI Render | Surface Render | Priority |
|--------|-----------|----------------|----------|
| **Markdown** | ANSI-styled text (headers bold, code inverted) | HTML (native) | P0 |
| **GSM** | ASCII state machine diagrams | SVG | P0 |
| **Diff** | Unified diff (+, -, @@ lines colored) | HTML diff | P0 |
| **JSON** | Indented with syntax colors | Code block with highlight | P1 |
| **CSV** | ASCII table with borders | HTML table | P1 |
| **Mermaid** | Text fallback + "render in browser" hint | Mermaid.js | P2 |
| **DOT/Graphviz** | ASCII topology (VFS tree) | SVG via viz.js | P2 |
| **PlantUML** | Text fallback | PlantUML server render | P3 |

## Vaked Fit

```
Declares:   format schema → "this is markdown", "this is a state machine"
Materializes: render engine → ANSI/TUI or HTML/Web
Reveals:    AG-UI block → styled output
```

## Rendering Strategy

Create `internal/blocks/render_engine.go` — the 8th engine.

```
ui-engine (Reveals: surfaces)
  └── render-engine (Reveals: formats)
       ├── Markdown → ANSI / HTML
       ├── GSM → ASCII / SVG
       ├── Diff → Colored / HTML
       ├── JSON → Syntax / Code
       ├── CSV → Table / HTML
       ├── Mermaid → Fallback / JS
       ├── DOT → VFS tree / SVG
       └── PlantUML → Fallback / Server
```

## Commands

```
/render md "# Hello"        → styled markdown
/render gsm "A → B → C"    → ASCII state machine
/render diff @@ -1,3 +1,3  → colored diff
/render json '{"a":1}'     → syntax-highlighted JSON
```
