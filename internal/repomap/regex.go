//go:build ignore

package repomap

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Regex-based symbol extraction as a zero-dependency fallback.
// Faster startup (no CGO, no grammar loading) and catches 90%+ of symbols.
// Tree-sitter path (parser.go, extract.go) is the full implementation when grammars are available.

// Regex patterns for symbol extraction per language.
var (
	goFuncRe     = regexp.MustCompile(`^func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`)
	goTypeRe     = regexp.MustCompile(`^type\s+(\w+)\s+`)
	goVarRe      = regexp.MustCompile(`^\s*var\s+(\w+)\s+`)
	goImportRe   = regexp.MustCompile(`"([^"]+)"`)
	pyFuncRe     = regexp.MustCompile(`^def\s+(\w+)\s*\(`)
	pyClassRe    = regexp.MustCompile(`^class\s+(\w+)\s*[(:]`)
	pyImportRe   = regexp.MustCompile(`^(?:from|import)\s+([\w.]+)`)
	zigFuncRe    = regexp.MustCompile(`^(?:pub\s+)?fn\s+(\w+)\s*\(`)
	zigVarRe     = regexp.MustCompile(`^(?:pub\s+)?(?:const|var)\s+(\w+)\s*[=:;]`)
	rustFuncRe   = regexp.MustCompile(`^(?:pub(?:\s*\(\s*crate\s*\))?\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+(\w+)\s*[<(]`)
	rustStructRe = regexp.MustCompile(`^(?:pub\s+)?(?:struct|enum|trait)\s+(\w+)`)
	rustImplRe   = regexp.MustCompile(`^impl\s+(?:(\w+)\s+for\s+)?(\w+)`)
	tsFuncRe     = regexp.MustCompile(`(?:function|async function)\s+(\w+)\s*\(`)
	tsClassRe    = regexp.MustCompile(`class\s+(\w+)`)
	tsExportRe   = regexp.MustCompile(`export\s+(?:default\s+)?(?:function|class|const|let|var)\s+(\w+)`)
)

// ExtractSymbolsRegex parses a file and returns symbols using regex patterns.
// This is the zero-dependency path — works without tree-sitter grammars.
func ExtractSymbolsRegex(path string, src []byte) []Symbol {
	ext := filepath.Ext(path)
	if strings.HasSuffix(path, ".d.ts") || strings.HasSuffix(path, ".test.ts") {
		ext = ".ts"
	}
	if strings.HasSuffix(path, ".test.js") {
		ext = ".js"
	}

	switch ext {
	case ".go":
		return extractGoRegex(path, src)
	case ".py":
		return extractPythonRegex(path, src)
	case ".zig":
		return extractZigRegex(path, src)
	case ".rs":
		return extractRustRegex(path, src)
	case ".ts", ".tsx", ".js", ".jsx", ".mjs":
		return extractTSRegex(path, src)
	default:
		return nil
	}
}

func extractGoRegex(path string, src []byte) []Symbol {
	var syms []Symbol
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		// Function declaration
		if m := goFuncRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !isKeyword(name) {
				exported := name[0] >= 'A' && name[0] <= 'Z'
				syms = append(syms, Symbol{
					Name: name, Kind: "func", File: path, Line: lineNo,
					Exported: exported,
					Calls:    extractGoCalls(line),
				})
			}
			continue
		}

		// Type declaration
		if m := goTypeRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !isKeyword(name) {
				exported := name[0] >= 'A' && name[0] <= 'Z'
				syms = append(syms, Symbol{
					Name: name, Kind: "type", File: path, Line: lineNo, Exported: exported,
				})
			}
			continue
		}

		// Import
		if strings.HasPrefix(strings.TrimSpace(line), "\"") || strings.HasPrefix(strings.TrimSpace(line), "import") {
			for _, m := range goImportRe.FindAllStringSubmatch(line, -1) {
				pkg := m[1]
				syms = append(syms, Symbol{
					Name: pkg, Kind: "import", File: path, Line: lineNo,
				})
			}
		}
	}

	return syms
}

func extractGoCalls(line string) []string {
	// Extract identifiers from function body lines — simplified
	re := regexp.MustCompile(`\b([A-Z]\w*|[a-z]\w+)\s*\(`)
	matches := re.FindAllStringSubmatch(line, -1)
	seen := make(map[string]bool)
	var calls []string
	for _, m := range matches {
		name := m[1]
		if !isKeyword(name) && !seen[name] {
			seen[name] = true
			calls = append(calls, name)
		}
	}
	return calls
}

func extractPythonRegex(path string, src []byte) []Symbol {
	var syms []Symbol
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())

		if m := pyFuncRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !isKeyword(name) && !strings.HasPrefix(name, "_") {
				syms = append(syms, Symbol{
					Name: name, Kind: "func", File: path, Line: lineNo, Exported: !strings.HasPrefix(name, "_"),
				})
			}
			continue
		}

		if m := pyClassRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			syms = append(syms, Symbol{
				Name: name, Kind: "class", File: path, Line: lineNo, Exported: true,
			})
			continue
		}

		if m := pyImportRe.FindStringSubmatch(line); m != nil {
			syms = append(syms, Symbol{
				Name: m[1], Kind: "import", File: path, Line: lineNo,
			})
		}
	}

	return syms
}

func extractZigRegex(path string, src []byte) []Symbol {
	var syms []Symbol
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())

		if m := zigFuncRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !isKeyword(name) {
				syms = append(syms, Symbol{
					Name: name, Kind: "func", File: path, Line: lineNo,
					Exported: strings.HasPrefix(line, "pub"),
				})
			}
			continue
		}

		if m := zigVarRe.FindStringSubmatch(line); m != nil {
			syms = append(syms, Symbol{
				Name: m[1], Kind: "var", File: path, Line: lineNo,
			})
		}
	}

	return syms
}

func extractRustRegex(path string, src []byte) []Symbol {
	var syms []Symbol
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())

		if m := rustFuncRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !isKeyword(name) {
				syms = append(syms, Symbol{
					Name: name, Kind: "func", File: path, Line: lineNo,
					Exported: strings.HasPrefix(line, "pub"),
				})
			}
			continue
		}

		if m := rustStructRe.FindStringSubmatch(line); m != nil {
			syms = append(syms, Symbol{
				Name: m[1], Kind: "type", File: path, Line: lineNo,
				Exported: strings.HasPrefix(line, "pub"),
			})
		}
	}

	return syms
}

func extractTSRegex(path string, src []byte) []Symbol {
	var syms []Symbol
	scanner := bufio.NewScanner(bytes.NewReader(src))
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())

		if m := tsExportRe.FindStringSubmatch(line); m != nil {
			syms = append(syms, Symbol{
				Name: m[1], Kind: "export", File: path, Line: lineNo, Exported: true,
			})
			continue
		}

		if m := tsFuncRe.FindStringSubmatch(line); m != nil {
			syms = append(syms, Symbol{
				Name: m[1], Kind: "func", File: path, Line: lineNo,
			})
			continue
		}

		if m := tsClassRe.FindStringSubmatch(line); m != nil {
			syms = append(syms, Symbol{
				Name: m[1], Kind: "class", File: path, Line: lineNo, Exported: true,
			})
		}
	}

	return syms
}

// BuildGraphRegex is BuildGraph using regex extraction (zero-dependency path).
func BuildGraphRegex(root string, paths []string, cache *Cache) (*Graph, error) {
	g := NewGraph()

	for _, relPath := range paths {
		fullPath := filepath.Join(root, relPath)

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			continue
		}

		g.AddFile(relPath)

		if cache != nil {
			if syms, ok := cache.Get(relPath); ok {
				for _, sym := range syms {
					g.AddSymbol(sym)
				}
				continue
			}
		}

		src, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		syms := ExtractSymbolsRegex(relPath, src)
		if cache != nil {
			_ = cache.Put(relPath, syms)
		}

		for _, sym := range syms {
			g.AddSymbol(sym)
		}
	}

	// Build import edges
	for _, sym := range g.Symbols {
		if sym.Kind == "import" {
			resolved := resolveImport(sym.Name, g.Files)
			if resolved != "" {
				g.AddEdge(sym.File, resolved, "import")
			}
		}
	}

	return g, nil
}
