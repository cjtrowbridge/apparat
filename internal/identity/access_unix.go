//go:build !windows

// Restricts the private identity directory to the owner before writing keys.
// Rejects a replaced directory or symlink instead of following it.
package identity

import (
	"errors"
	"os"
)

func protectDirectory(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		return errors.New("identity directory is unavailable")
	}
	return os.Chmod(dir, 0700)
}
