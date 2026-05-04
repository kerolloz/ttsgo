//go:build !windows
// +build !windows

package process

import (
	"context"
	"os/exec"
	"syscall"
	"time"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killTree kills a process and all its children by targeting its process group.
func killTree(pid int) {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		// Fallback to just the PID if PGID lookup fails
		_ = syscall.Kill(pid, syscall.SIGTERM)
		return
	}

	// SIGTERM the entire group
	_ = syscall.Kill(-pgid, syscall.SIGTERM)

	// Wait up to 2s for graceful exit
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Timeout reached, force kill
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			return
		case <-ticker.C:
			// Check if any process in the group still exists
			if err := syscall.Kill(-pgid, 0); err != nil {
				return // Group is dead
			}
		}
	}
}
