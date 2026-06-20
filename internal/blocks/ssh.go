package blocks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ── SSH Primitive — Tool-Level Remote Execution ──────────────────────
//
// Tool-level: wraps local 'ssh' command. No agent deployed to remote.
// Key management: bao secrets vault (VAULT_TOKEN) or local ~/.ssh/.
// POV: always the parent who originated the SSH command.
// Lifecycle: PID-file lock, pause/stop/start/restart. No rollback.
// Scope: available to all agents (AGENT scope).

// SSHBlock represents a single SSH execution.
type SSHBlock struct {
	ID        string        // unique run ID
	Host      string        // "dev-cx53" or "root@167.233.105.32"
	Command   string        // the remote command
	KeyPath   string        // SSH private key path (from bao or local)
	User      string        // SSH user (default: "dev")
	Timeout   time.Duration // command timeout
	PID       int           // local SSH process PID
	PIDFile   string        // lock file: ~/.whale/ssh/{id}.pid
	Status    string        // "running", "completed", "failed", "paused", "stopped"
	Result    []byte        // stdout+stderr
	Ref       string        // sha256 of result
	ExitCode  int
	POV       POV           // parent originator POV (never remote)
	StartTime time.Time
	EndTime   time.Time

	mu       sync.Mutex
	cmd      *exec.Cmd
	cancelFn func()
}

// ── SSH Config ────────────────────────────────────────────────────────

// SSHConfig holds SSH key and connection preferences.
type SSHConfig struct {
	KeyPath    string // path to private key
	DefaultUser string // "dev" or "root"
	KnownHosts  string // ~/.ssh/known_hosts
	BaoKeyPath string // bao secret path (e.g. "secret/data/ssh/ultrawhale")
}

// LoadSSHConfig loads SSH config from bao or local files.
func LoadSSHConfig() *SSHConfig {
	home, _ := os.UserHomeDir()

	cfg := &SSHConfig{
		DefaultUser: "dev",
		KnownHosts:  filepath.Join(home, ".ssh", "known_hosts"),
		BaoKeyPath:  "secret/data/ssh/ultrawhale",
	}

	// Try bao first
	if os.Getenv("VAULT_TOKEN") != "" && os.Getenv("VAULT_TOKEN") != "" {
		key, err := fetchBaoSSHKey(cfg.BaoKeyPath)
		if err == nil {
			cfg.KeyPath = key
			return cfg
		}
	}

	// Fallback: local key
	localKey := filepath.Join(home, ".ssh", "id_ed25519_ultrawhale")
	if _, err := os.Stat(localKey); err == nil {
		cfg.KeyPath = localKey
		return cfg
	}

	// Generate one if missing
	cfg.KeyPath = localKey
	generateSSHKey(localKey)
	return cfg
}

func fetchBaoSSHKey(path string) (string, error) {
	// Stub: fetch from bao.crabcc.app
	_ = path
	return "", fmt.Errorf("bao SSH key fetch not implemented")
}

func generateSSHKey(path string) {
	os.MkdirAll(filepath.Dir(path), 0o700)
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", path, "-N", "", "-C", "ultrawhale")
	cmd.Run()
	Log(LogInfo, "blocks.SSH", "keygen", path, "", 0, nil)
}

// ── SSH Execution ─────────────────────────────────────────────────────

var sshStore = &sshRunStore{runs: make(map[string]*SSHBlock)}

type sshRunStore struct {
	mu   sync.Mutex
	runs map[string]*SSHBlock
}

// SSHExec executes a command on a remote host via SSH.
func SSHExec(host, command string) (*SSHBlock, error) {
	cfg := LoadSSHConfig()
	user := detectSSHUser(host)
	if strings.Contains(host, "@") {
		parts := strings.SplitN(host, "@", 2)
		user = parts[0]
		host = parts[1]
	}

	id := fmt.Sprintf("ssh-%d", time.Now().UnixNano())
	run := &SSHBlock{
		ID:        id,
		Host:      host,
		Command:   command,
		KeyPath:   cfg.KeyPath,
		User:      user,
		Timeout:   30 * time.Second,
		Status:    "running",
		POV:       CurrentPOV(),
		StartTime: time.Now(),
		PIDFile:   filepath.Join(os.Getenv("HOME"), ".whale", "ssh", id+".pid"),
	}

	// Create PID file lock
	os.MkdirAll(filepath.Dir(run.PIDFile), 0o700)
	os.WriteFile(run.PIDFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0o600)

	sshStore.mu.Lock()
	sshStore.runs[id] = run
	sshStore.mu.Unlock()

	// Build SSH command
	args := []string{
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", fmt.Sprintf("UserKnownHostsFile=%s", cfg.KnownHosts),
		"-i", cfg.KeyPath,
		"-l", user,
		host,
		command,
	}

	run.cmd = exec.Command("ssh", args...)

	// Start asynchronously
	go func() {
		output, err := run.cmd.CombinedOutput()
		run.EndTime = time.Now()
		run.mu.Lock()
		defer run.mu.Unlock()

		run.Result = output
		run.Ref = Ref(output)
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					run.ExitCode = status.ExitStatus()
				}
			}
			run.Status = "failed"
		} else {
			run.Status = "completed"
			run.ExitCode = 0
		}
		run.PID = run.cmd.Process.Pid

		// Clean PID file
		os.Remove(run.PIDFile)

		// Log the execution
		Log(LogInfo, "blocks.SSH", fmt.Sprintf("%s@%s: %s", user, host, command),
			run.Ref, "", run.EndTime.Sub(run.StartTime),
			err)
	}()

	run.PID = run.cmd.Process.Pid
	return run, nil
}

// ── Lifecycle: Pause / Stop / Start / Restart ─────────────────────────

// PauseSSH sends SIGSTOP to the SSH process.
func PauseSSH(id string) error {
	run := getSSHRun(id)
	if run == nil { return fmt.Errorf("ssh run not found: %s", id) }
	if run.Status != "running" { return fmt.Errorf("cannot pause %s run", run.Status) }

	if run.cmd != nil && run.cmd.Process != nil {
		run.cmd.Process.Signal(syscall.SIGSTOP)
	}
	run.Status = "paused"
	Log(LogInfo, "blocks.SSH", fmt.Sprintf("paused %s", id), "", "", 0, nil)
	return nil
}

// ResumeSSH sends SIGCONT to the SSH process.
func ResumeSSH(id string) error {
	run := getSSHRun(id)
	if run == nil { return fmt.Errorf("ssh run not found: %s", id) }
	if run.Status != "paused" { return fmt.Errorf("cannot resume %s run", run.Status) }

	if run.cmd != nil && run.cmd.Process != nil {
		run.cmd.Process.Signal(syscall.SIGCONT)
	}
	run.Status = "running"
	Log(LogInfo, "blocks.SSH", fmt.Sprintf("resumed %s", id), "", "", 0, nil)
	return nil
}

// StopSSH sends SIGTERM to the SSH process.
func StopSSH(id string) error {
	run := getSSHRun(id)
	if run == nil { return fmt.Errorf("ssh run not found: %s", id) }
	if run.Status == "stopped" || run.Status == "completed" || run.Status == "failed" {
		return fmt.Errorf("already %s", run.Status)
	}

	if run.cmd != nil && run.cmd.Process != nil {
		run.cmd.Process.Signal(syscall.SIGTERM)
	}
	run.Status = "stopped"
	run.EndTime = time.Now()
	os.Remove(run.PIDFile)
	Log(LogInfo, "blocks.SSH", fmt.Sprintf("stopped %s", id), "", "", time.Since(run.StartTime), nil)
	return nil
}

// RestartSSH re-runs the same command on the same host.
func RestartSSH(id string) (*SSHBlock, error) {
	run := getSSHRun(id)
	if run == nil { return nil, fmt.Errorf("ssh run not found: %s", id) }

	// Stop existing
	_ = StopSSH(id)

	// Re-execute
	return SSHExec(
		fmt.Sprintf("%s@%s", run.User, run.Host),
		run.Command,
	)
}

// ── Status ────────────────────────────────────────────────────────────

func getSSHRun(id string) *SSHBlock {
	sshStore.mu.Lock()
	defer sshStore.mu.Unlock()
	return sshStore.runs[id]
}

// SSHList returns all SSH runs.
func SSHList() []*SSHBlock {
	sshStore.mu.Lock()
	defer sshStore.mu.Unlock()
	result := make([]*SSHBlock, 0, len(sshStore.runs))
	for _, r := range sshStore.runs {
		result = append(result, r)
	}
	return result
}

// SSHStatus returns compact status.
func SSHStatus() string {
	sshStore.mu.Lock()
	defer sshStore.mu.Unlock()
	return fmt.Sprintf("ssh: %d runs", len(sshStore.runs))
}

func detectSSHUser(host string) string {
	if strings.Contains(host, "@") { return "dev" }
	// Common hosts
	switch {
	case strings.Contains(host, "dev-cx53"): return "dev"
	case strings.Contains(host, "167.233"): return "root"
	default: return "dev"
	}
}
