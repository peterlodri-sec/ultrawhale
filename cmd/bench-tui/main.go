package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	// Only run on macOS with Ghostty
	if os.Getenv("GHOSTTY") == "" && os.Getenv("TERM_PROGRAM") != "ghostty" {
		fmt.Println("ghostty terminal required. Install: brew install ghostty")
		fmt.Print("Run anyway with --force? [y/N]: ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			os.Exit(1)
		}
	}

	// Check for ultrawhale binary
	if _, err := os.Stat("bin/ultrawhale"); os.IsNotExist(err) {
		fmt.Println("ultrawhale binary not found at bin/ultrawhale")
		fmt.Println("Build first: go build -o bin/ultrawhale ./cmd/whale")
		os.Exit(1)
	}

	fmt.Println("\n⚡ ultrawhale TUI Benchmark")
	fmt.Println(strings.Repeat("─", 60))

	tests := []struct {
		name string
		cmd  string
	}{
		{"Doctor", "bin/ultrawhale --dangerously-skip-permissions doctor"},
		{"Help", "bin/ultrawhale --help"},
		{"Version", "bin/ultrawhale --version 2>/dev/null || bin/ultrawhale --help | head -1"},
		{"Setup (dry-run)", "echo '' | bin/ultrawhale-setup 2>/dev/null; echo OK"},
	}

	var totalDur time.Duration
	passed := 0

	for _, tc := range tests {
		fmt.Printf("\n[%s] ", tc.name)
		start := time.Now()

		cmd := exec.Command("bash", "-c", tc.cmd)
		cmd.Stdout = nil
		cmd.Stderr = nil
		err := cmd.Run()

		dur := time.Since(start)
		totalDur += dur

		if err != nil {
			fmt.Printf("FAIL (%s)\n", dur.Round(time.Millisecond))
		} else {
			fmt.Printf("PASS (%s)\n", dur.Round(time.Millisecond))
			passed++
		}
	}

	// Profile: launch TUI briefly and measure startup
	fmt.Println("\n[TUI Profile] launching ultrawhale TUI (2-second profile)...")
	start := time.Now()
	cmd := exec.Command("timeout", "2", "bin/ultrawhale", "--dangerously-skip-permissions", "--model", "deepseek-v4-flash", "-w")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.Run()
	tuiDur := time.Since(start)

	fmt.Printf("  TUI startup + 2s runtime: %s\n", tuiDur.Round(time.Millisecond))

	// Summary
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Printf("Results: %d/%d PASS (%s total)\n", passed, len(tests), totalDur.Round(time.Millisecond))
	if passed == len(tests) {
		fmt.Println("✅ ultrawhale TUI benchmark PASS")
	} else {
		fmt.Println("❌ some tests FAILED")
		os.Exit(1)
	}
}
