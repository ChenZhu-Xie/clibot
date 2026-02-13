package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// expandHome expands ~ to user's home directory
// This is a shared utility function used across multiple CLI adapters
// Returns an error if the home directory cannot be determined
func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// buildShellCommand creates a cross-platform shell command
// For Linux/macOS (including WSL2): sh -c "command"
// For Windows (native, though not officially supported): cmd /c "command"
func buildShellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		// Windows native: use cmd /c (not officially supported)
		return exec.Command("cmd", "/c", command)
	}
	// Linux/macOS (including WSL2): use sh -c
	return exec.Command("sh", "-c", command)
}
