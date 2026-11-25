//go:build !linux || android

// Offers remote terminal clients on targets without qualified local PTY isolation.
// Local startup fails before any shell is launched.
package terminal

import (
	"errors"
	"os"
	"os/exec"
)

func prerequisite() string {
	return "Select a Linux host for terminal sessions."
}
func start(string, string) (*os.File, *exec.Cmd, error) {
	return nil, nil, errors.New(prerequisite())
}
func setSize(*os.File, int, int) error {
	return errors.New(prerequisite())
}
