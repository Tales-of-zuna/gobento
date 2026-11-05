//go:build windows

package generate

import (
	"os/exec"
	"strconv"
)

// killTree makes cancelling cmd's context end cmd and every process it
// started, with taskkill /T. npm runs scaffolders as child processes, which
// would otherwise keep writing into a directory that cleanup is removing.
func killTree(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)) //nolint:gosec // G204: the PID of the npm process gobento started
		if err := kill.Run(); err != nil {
			return cmd.Process.Kill()
		}
		return nil
	}
}
