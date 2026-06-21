//go:build arm64

package asm

// FindString uses Go stdlib on arm64 (Apple Silicon NEON auto-used).
func FindString(haystack, needle string) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle { return i }
	}
	return -1
}

// CountLines counts newlines — ARM64 NEON is auto-used by Go compiler.
func CountLines(data string) int {
	count := 0
	for _, ch := range data {
		if ch == '\n' { count++ }
	}
	return count
}
