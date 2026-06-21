package blocks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Glob returns files matching a pattern. Journaled.
func Glob(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil { return nil, err }

	// Filter: only regular files, skip dirs + symlinks
	var files []string
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil { continue }
		if info.Mode().IsRegular() { files = append(files, m) }
	}

	Log(LogInfo, "blocks.Glob", fmt.Sprintf("%s → %d files", pattern, len(files)),
		"", "", 0, nil)
	return files, nil
}

// GlobRecursive recursively finds files matching a pattern.
func GlobRecursive(root, pattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
