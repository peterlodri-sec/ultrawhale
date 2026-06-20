package native

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ── Grep (ripgrep replacement) ────────────────────────────────────────

// Grep searches files for a pattern. Returns matches with file, line number, and content.
// Zero shell exec — pure Go. Uses regexp for pattern matching, bytes.Index for literal.
type GrepResult struct {
	File    string
	Line    int
	Content string
	Column  int
}

// GrepOptions configures a grep search.
type GrepOptions struct {
	Pattern    string   // regex or literal pattern
	Path       string   // root directory
	Include    string   // glob include filter (e.g. "*.go")
	MaxResults int      // max matches (default 100)
	Literal    bool     // literal match instead of regex
	IgnoreCase bool     // case-insensitive
}

// Grep runs a grep search. Returns up to MaxResults matches.
func Grep(opts GrepOptions) ([]GrepResult, error) {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 100
	}

	var re *regexp.Regexp
	var literal []byte
	if opts.Literal {
		literal = []byte(opts.Pattern)
		if opts.IgnoreCase {
			literal = bytes.ToLower(literal)
		}
	} else {
		pattern := opts.Pattern
		if opts.IgnoreCase {
			pattern = "(?i)" + pattern
		}
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}

	var results []GrepResult
	err := filepath.Walk(opts.Path, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Skip hidden dirs
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Include filter
		if opts.Include != "" {
			matched, _ := filepath.Match(opts.Include, filepath.Base(fullPath))
			if !matched {
				return nil
			}
		}
		// Skip binary
		ext := filepath.Ext(fullPath)
		if isBinaryExt(ext) {
			return nil
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return nil
		}
		defer f.Close()

		relPath, _ := filepath.Rel(opts.Path, fullPath)
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB lines
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			if len(results) >= opts.MaxResults {
				return filepath.SkipAll
			}
			line := scanner.Bytes()
			var matched bool
			var col int
			if literal != nil {
				searchLine := line
				if opts.IgnoreCase {
					searchLine = bytes.ToLower(line)
				}
				idx := bytes.Index(searchLine, literal)
				matched = idx >= 0
				col = idx + 1
			} else if re != nil {
				loc := re.FindIndex(line)
				matched = loc != nil
				if matched {
					col = loc[0] + 1
				}
			}
			if matched {
				results = append(results, GrepResult{
					File:    relPath,
					Line:    lineNo,
					Content: string(line),
					Column:  col,
				})
			}
		}
		return nil
	})
	return results, err
}

func isBinaryExt(ext string) bool {
	switch ext {
	case ".o", ".a", ".so", ".exe", ".bin", ".zip", ".gz", ".tar",
		".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf",
		".class", ".pyc", ".wasm":
		return true
	}
	return false
}
