package osx

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// OpenBrowser go/src/cmd/internal/browser/browser.go
func OpenBrowser(url string) bool {
	var cmds [][]string
	if exe := os.Getenv("BROWSER"); exe != "" {
		cmds = append(cmds, []string{exe})
	}
	switch runtime.GOOS {
	case "darwin":
		cmds = append(cmds, []string{"/usr/bin/open"})
	case "windows":
		cmds = append(cmds, []string{"cmd", "/c", "start"})
	default:
		if os.Getenv("DISPLAY") != "" {
			// xdg-open is only for use in a desktop environment.
			cmds = append(cmds, []string{"xdg-open"})
		}
	}
	cmds = append(cmds, []string{"chrome"}, []string{"google-chrome"}, []string{"chromium"}, []string{"firefox"})
	for _, args := range cmds {
		cmd := exec.Command(args[0], append(args[1:], url)...)
		if cmd.Start() == nil && WaitTimeout(cmd, 3*time.Second) {
			return true
		}
	}
	return false
}

// WaitTimeout reports whether the command appears to have run successfully.
// If the command runs longer than the timeout, it's deemed successful.
// If the command runs within the timeout, it's deemed successful if it exited cleanly.
func WaitTimeout(cmd *exec.Cmd, timeout time.Duration) bool {
	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		return true
	case err := <-ch:
		return err == nil
	}
}

func SleepContext(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}
	return
}
