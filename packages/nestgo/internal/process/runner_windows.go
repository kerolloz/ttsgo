//go:build windows
// +build windows

package process

import (
	"os/exec"
	"strconv"
)

func setSysProcAttr(cmd *exec.Cmd) {
	// No Setpgid on Windows
}

// killTree kills a process and all its children on Windows using taskkill.
func killTree(pid int) {
	// taskkill /F /T /PID <pid>
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	_ = cmd.Run()
}
