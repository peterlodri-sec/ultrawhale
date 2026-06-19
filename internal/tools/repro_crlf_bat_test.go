package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usewhale/whale/internal/core"
)

// Repro for the Windows report (session 019ec77f): the write tool created
// build.bat / deploy.bat with LF-only line endings. The on-disk .bat read back
// from that session was pure "\n" (no "\r\n"). Windows cmd.exe misparses
// LF-only batch files — notably the "^" line-continuation in build.bat — which
// is what the user described as "cmd 解析时把文本切碎了".
//
// A newly created Windows script file should be written with CRLF. This test
// drives the real write tool through the registry and inspects the bytes on
// disk. It FAILS on current code (new files are written verbatim, so LF),
// which is exactly the bug we want to reproduce before fixing.
func TestWriteNewBatFileUsesCRLF(t *testing.T) {
	dir := t.TempDir()
	ts, err := NewToolset(dir)
	if err != nil {
		t.Fatalf("new toolset: %v", err)
	}
	reg := core.NewToolRegistry(ts.Tools())

	// Model-authored content uses LF, as the model always emits.
	const content = "@echo off\nchcp 65001 >nul\n%CSC% ^\n  /target:library\n"

	res, err := reg.Dispatch(context.Background(), core.ToolCall{
		ID:    "tc-write-bat",
		Name:  "write",
		Input: `{"file_path":"build.bat","content":"@echo off\nchcp 65001 >nul\n%CSC% ^\n  /target:library\n"}`,
	})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if res.IsError() {
		t.Fatalf("write rejected:\n%s", res.ModelText)
	}
	_ = content

	got, err := os.ReadFile(filepath.Join(dir, "build.bat"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "\n") && !strings.Contains(string(got), "\r\n") {
		t.Fatalf("BUG REPRODUCED: new .bat written with LF-only endings; Windows cmd will misparse it.\nbytes=%q", string(got))
	}
}
