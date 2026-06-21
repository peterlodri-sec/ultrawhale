package blocks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// ── Diff Primitive ────────────────────────────────────────────────────
// Generates unified diffs between two file states.
// Uses system diff or built-in Myers diff for small files.

// Diff generates a unified diff between two byte slices.
func Diff(a, b []byte, labelA, labelB string) (string, error) {
	_ = CurrentPOV()
	// For small files (<64KB), use built-in line dif
	if len(a) < 65536 && len(b) < 65536 {
		return lineDiff(a, b, labelA, labelB), nil
	}
	// For large files, delegate to system diff
	return systemDiff(a, b, labelA, labelB)
}

func lineDiff(a, b []byte, labelA, labelB string) string {
	aLines := bytes.Split(a, []byte("\n"))
	bLines := bytes.Split(b, []byte("\n"))

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", labelA, labelB))

	// Simple line-by-line diff (LCS-based Myers for production)
	i, j := 0, 0
	for i < len(aLines) || j < len(bLines) {
		if i < len(aLines) && j < len(bLines) && bytes.Equal(aLines[i], bLines[j]) {
			buf.WriteString(fmt.Sprintf("  %s\n", aLines[i]))
			i++; j++
		} else {
			if i < len(aLines) {
				buf.WriteString(fmt.Sprintf("- %s\n", aLines[i]))
				i++
			}
			if j < len(bLines) {
				buf.WriteString(fmt.Sprintf("+ %s\n", bLines[j]))
				j++
			}
		}
	}
	return buf.String()
}

func systemDiff(a, b []byte, labelA, labelB string) (string, error) {
	tmpA, _ := os.CreateTemp("", "ultrawhale-diff-a")
	tmpB, _ := os.CreateTemp("", "ultrawhale-diff-b")
	defer os.Remove(tmpA.Name())
	defer os.Remove(tmpB.Name())
	tmpA.Write(a); tmpA.Close()
	tmpB.Write(b); tmpB.Close()

	cmd := exec.Command("diff", "-u", "--label", labelA, tmpA.Name(), "--label", labelB, tmpB.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		// diff returns exit 1 when files differ — that's normal
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return string(out), nil
		}
		return "", err
	}
	return string(out), nil
}

// DiffFile generates a diff between current file and its journaled previous state.
func DiffFile(path string) (string, error) {
	current, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	prev := journal.Pop(path)
	if prev == nil {
		return fmt.Sprintf("--- %s (new file)\n+++ %s\n", path, path), nil
	}
	return Diff(prev.Content, current, path+" (prev)", path)
}
