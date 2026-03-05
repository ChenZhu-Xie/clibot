//go:build !linux

package cli

import "syscall"

// setPdeathsig is a no-op on non-Linux platforms
// macOS and Windows handle process termination differently
// For macOS: Session processes are cleaned up during graceful shutdown
// For Windows: Not officially supported
func setPdeathsig(attrs *syscall.SysProcAttr) {
	// No-op on non-Linux platforms
}

// setSetpgid is a no-op on non-Linux platforms
func setSetpgid(attrs *syscall.SysProcAttr) {
	// No-op: Setpgid is not available on all platforms
}
