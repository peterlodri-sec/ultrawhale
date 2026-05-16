//go:build !unix && !windows

package shell

import (
	"context"
	"os/exec"
	"time"
)

// ConfigureCommand applies platform process settings for shell commands.
func ConfigureCommand(cmd *exec.Cmd) {
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
	cancel := cmd.Cancel
	if err := cmd.Start(); err != nil {
		return err
	}
	return waitCommandContext(ctx, cmd, cancel)
}
