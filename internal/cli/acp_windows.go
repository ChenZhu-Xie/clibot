//go:build windows

package cli

import (
	"github.com/keepmind9/clibot/internal/logger"
)

// killProcess terminates the ACP server process on Windows
func (a *ACPAdapter) killProcess(sessionName string) error {
	if !a.isRemote && a.cmd != nil && a.cmd.Process != nil {
		logger.WithField("session", sessionName).Info("killing-acp-process")

		killErr := a.cmd.Process.Kill()

		if killErr != nil {
			logger.WithField("error", killErr).Warn("failed-to-kill-acp-process")
		}

		// Wait for process to exit
		a.cmd.Wait()
		a.cmd = nil
	}

	return nil
}
