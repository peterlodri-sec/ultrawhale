//go:build amd64

package asm

//go:noescape
func findSubstringAmd64(haystack []byte, needle []byte) int

//go:noescape
func countLinesAmd64(data []byte) int

// FindString uses ASM-accelerated substring search on amd64.
func FindString(haystack, needle string) int {
	return findSubstringAmd64([]byte(haystack), []byte(needle))
}

// CountLines uses POPCNT-accelerated line counting on amd64.
func CountLines(data string) int {
	return countLinesAmd64([]byte(data))
}
