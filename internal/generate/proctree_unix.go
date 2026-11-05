//go:build !windows

package generate

import (
	"os/exec"
	"syscall"
)

// killTree makes cancelling cmd's context kill cmd and every process it
// started: cmd runs in a process group of its own, and the whole group is
// killed. npm runs scaffolders as child processes, which would otherwise
// keep writing into a directory that cleanup is removing.
func killTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
