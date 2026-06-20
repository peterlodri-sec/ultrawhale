package repomap

import (
	"sort"
	"strings"
)

// Symbol represents a named code entity (function, type, variable, import).
type Symbol struct {
	Name     string   `json:"name"`
	Kind     string   `json:"kind"` // func, type, var, const, import, module
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Calls    []string `json:"calls,omitempty"`    // symbols this symbol references
	Exported bool     `json:"exported,omitempty"` // starts with uppercase (Go/Rust) or __all__ (Python)
}

// Edge represents a dependency between two files.
type Edge struct {
	From string `json:"from"` // file path
	To   string `json:"to"`   // file path
	Kind string `json:"kind"` // import, call, type_ref
}

// Graph is the complete repository dependency graph.
type Graph struct {
	Files   []string          `json:"files"`
	Symbols map[string]Symbol `json:"symbols"` // keyed by "file:name"
	Edges   []Edge            `json:"edges"`
}

// NewGraph creates an empty graph.
func NewGraph() *Graph {
	return &Graph{
		Symbols: make(map[string]Symbol),
	}
}

// AddFile registers a file in the graph.
func (g *Graph) AddFile(path string) {
	for _, f := range g.Files {
		if f == path {
			return
		}
	}
	g.Files = append(g.Files, path)
}

// AddSymbol adds a symbol to the graph.
func (g *Graph) AddSymbol(sym Symbol) {
	key := sym.File + ":" + sym.Name
	g.Symbols[key] = sym
}

// AddEdge adds a dependency edge between two files.
func (g *Graph) AddEdge(from, to, kind string) {
	// Deduplicate
	for _, e := range g.Edges {
		if e.From == from && e.To == to && e.Kind == kind {
			return
		}
	}
	g.Edges = append(g.Edges, Edge{From: from, To: to, Kind: kind})
}

// SymbolsInFile returns all symbols defined in a given file.
func (g *Graph) SymbolsInFile(file string) []Symbol {
	var out []Symbol
	for _, sym := range g.Symbols {
		if sym.File == file {
			out = append(out, sym)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Line < out[j].Line })
	return out
}

// CallersOf returns all symbols that reference the given symbol name.
func (g *Graph) CallersOf(name string) []Symbol {
	var out []Symbol
	for _, sym := range g.Symbols {
		for _, call := range sym.Calls {
			if call == name {
				out = append(out, sym)
				break
			}
		}
	}
	return out
}

// DependentsOf returns all files that import or reference the given file.
func (g *Graph) DependentsOf(file string) []string {
	seen := make(map[string]bool)
	for _, e := range g.Edges {
		if e.To == file {
			seen[e.From] = true
		}
	}
	var out []string
	for f := range seen {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// ToJSON returns a compact JSON representation suitable for LLM context injection.
// Format matches Aider's repo map: {files: [...], symbols: {...}, edges: [...]}
func (g *Graph) ToJSON() string {
	var sb strings.Builder
	sb.WriteString("{")
	sb.WriteString("\"files\":[")
	for i, f := range g.Files {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString("\"" + f + "\"")
	}
	sb.WriteString("],\"symbols\":{")
	first := true
	for _, sym := range g.Symbols {
		if !first {
			sb.WriteByte(',')
		}
		first = false
		sb.WriteString("\"" + sym.File + ":" + sym.Name + "\":")
		sb.WriteString(symToJSON(sym))
	}
	sb.WriteString("},\"edges\":[")
	for i, e := range g.Edges {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString("{\"from\":\"" + e.From + "\",\"to\":\"" + e.To + "\",\"kind\":\"" + e.Kind + "\"}")
	}
	sb.WriteString("]}")
	return sb.String()
}

func symToJSON(sym Symbol) string {
	var sb strings.Builder
	sb.WriteString("{\"name\":\"" + sym.Name + "\"")
	sb.WriteString(",\"kind\":\"" + sym.Kind + "\"")
	sb.WriteString(",\"file\":\"" + sym.File + "\"")
	sb.WriteString(",\"line\":" + itoa(sym.Line))
	if sym.Exported {
		sb.WriteString(",\"exported\":true")
	}
	if len(sym.Calls) > 0 {
		sb.WriteString(",\"calls\":[")
		for i, c := range sym.Calls {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString("\"" + c + "\"")
		}
		sb.WriteString("]")
	}
	sb.WriteString("}")
	return sb.String()
}

// itoa is a simple int-to-string for small integers (line numbers).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
