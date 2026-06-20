package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type ExecRequest struct {
	Command   string `json:"command"`
	POV       map[string]string `json:"pov"`
	TimeoutSec int  `json:"timeout_sec"`
}

type ExecResponse struct {
	Result     string `json:"result"`
	ExitCode   int    `json:"exit_code"`
	Ref        string `json:"ref"`
	DurationMs int64  `json:"duration_ms"`
}

var startTime = time.Now()

func main() {
	port := "9797"
	if p := os.Getenv("ULTRAWHALE_SHELL_PORT"); p != "" { port = p }

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok", "version": "v4.10.0",
			"uptime": time.Since(startTime).String(),
		})
	})

	http.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {
		var req ExecRequest
		json.NewDecoder(r.Body).Decode(&req)
		start := time.Now()
		cmd := exec.Command("sh", "-c", req.Command)
		out, err := cmd.CombinedOutput()
		resp := ExecResponse{Result: string(out), DurationMs: time.Since(start).Milliseconds()}
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				resp.ExitCode = exitErr.ExitCode()
			}
		}
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Printf("ultrawhale-shell v4.10.0 on :%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
