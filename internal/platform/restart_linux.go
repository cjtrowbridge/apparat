//go:build linux && !android

// Replaces the current process after its GUI and node have shut down.
// Preserves Steam's process tracking, command arguments and launch environment.
package platform

import (
	"os"
	"syscall"
)

func RestartProcess() error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	return syscall.Exec(path, os.Args, os.Environ())
}
