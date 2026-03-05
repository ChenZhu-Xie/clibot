//go:build linux

package cli

import "syscall"

// setPdeathsig sets the parent death signal on Linux
// When the parent process dies, the kernel sends SIGTERM to the child
func setPdeathsig(attrs *syscall.SysProcAttr) {
	attrs.Pdeathsig = syscall.SIGTERM
}

// setSetpgid sets the process group ID on Linux
func setSetpgid(attrs *syscall.SysProcAttr) {
	attrs.Setpgid = true
}
