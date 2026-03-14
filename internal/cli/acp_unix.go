//go:build !windows

package cli

import (
	"os"
	"syscall"

	"github.com/keepmind9/clibot/internal/logger"
)

// killProcess terminates the ACP server process on Unix/Linux/macOS
func (a *ACPAdapter) killProcess(sess *acpSession) error {
	if !sess.isRemote && sess.cmd != nil && sess.cmd.Process != nil {
		logger.WithField("session", sess.workDir).Info("killing-acp-process")

		var killErr error
		// Unix/Linux/macOS: Kill entire process group using negative PID
		// The Setpgid: true in buildShellCommand ensures the process
		// is the process group leader, so -pid kills the entire group
		killErr = syscall.Kill(-sess.cmd.Process.Pid, syscall.SIGKILL)

		if killErr != nil {
			logger.WithField("error", killErr).Warn("failed-to-kill-acp-process")
		}

		// Wait for process to exit
		sess.cmd.Wait()
		sess.cmd = nil
	}

	return nil
}
// isSessionActive checks if the underlying process or connection for a session is still alive.
func (a *ACPAdapter) isSessionActive(sess *acpSession) bool {
	if sess.isRemote {
		if sess.conn == nil {
			return false
		}
		select {
		case <-sess.conn.Done():
			return false
		default:
			return true
		}
	} else {
		if sess.cmd == nil || sess.cmd.Process == nil {
			return false
		}
		return sess.cmd.Process.Signal(os.Signal(syscall.Signal(0))) == nil
	}
}
