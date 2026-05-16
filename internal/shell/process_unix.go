//go:build unix

package shell

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// ConfigureCommand applies platform process settings for shell commands.
func ConfigureCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
}

func RunCommand(ctx context.Context, cmd *exec.Cmd) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	ConfigureCommand(cmd)
	cancel := func() error {
		return killCommandGroup(cmd)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return waitCommandContext(ctx, cmd, cancel)
}

func killCommandGroup(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return os.ErrProcessDone
	}
	pid := cmd.Process.Pid
	if pid <= 0 {
		return os.ErrProcessDone
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	return nil
}
