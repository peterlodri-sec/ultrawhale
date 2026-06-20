package repomap

import "strings"

// isKeyword returns true for common language keywords that should not be treated as symbols.
func isKeyword(s string) bool {
	switch s {
	case "if", "else", "for", "while", "return", "break", "continue",
		"switch", "case", "default", "defer", "go", "select", "chan",
		"func", "var", "const", "type", "import", "package", "range",
		"true", "false", "nil", "null", "undefined", "None", "True", "False",
		"def", "class", "self", "cls", "async", "await", "yield",
		"pub", "fn", "let", "mut", "use", "mod", "struct", "enum", "trait",
		"export", "from", "in", "is", "not", "and", "or", "as",
		"try", "catch", "finally", "throw", "new", "delete", "typeof",
		"i", "j", "k", "x", "y", "z", "n", "v", "ok", "err":
		return true
	}
	return false
}

// resolveImport tries to match an import path to a known file in the repo.
func resolveImport(importPath string, files []string) string {
	candidates := []string{
		strings.TrimPrefix(importPath, "github.com/usewhale/whale/"),
		strings.ReplaceAll(importPath, ".", "/"),
		importPath,
	}

	for _, c := range candidates {
		for _, f := range files {
			if strings.HasPrefix(f, c) || strings.Contains(f, "/"+c+"/") || strings.HasSuffix(f, "/"+c) {
				return f
			}
		}
	}
	return ""
}
