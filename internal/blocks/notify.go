package blocks

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// ── Notify Primitive ──────────────────────────────────────────────────
// Desktop notifications for long-running task completion.
// macOS: osascript display notification
// Linux: notify-send

// Notify sends a desktop notification.
func Notify(title, message string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification "%s" with title "%s"`, message, title))
	case "linux":
		cmd = exec.Command("notify-send", title, message)
	default:
		return fmt.Errorf("notify: unsupported OS %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	Log(LogInfo, "blocks.Notify", fmt.Sprintf("%s [%s]", title, CurrentPOV().Machine), "", "", 0, nil)
	return nil
}

// NotifyTaskComplete sends a notification when a long-running task finishes.
func NotifyTaskComplete(taskName string, dur time.Duration, err error) {
	title := fmt.Sprintf("ultrawhale: %s complete", taskName)
	message := fmt.Sprintf("Duration: %s", dur.Round(time.Second))
	if err != nil {
		title = fmt.Sprintf("ultrawhale: %s failed", taskName)
		message = err.Error()
	}
	Notify(title, message)
}
