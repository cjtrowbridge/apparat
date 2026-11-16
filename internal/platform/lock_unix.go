//go:build linux || darwin || android

// Implements advisory exclusive locks on Unix and Android.
// Closing the owning descriptor releases the lock after process failure.
package platform

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func lockFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}
func unlockFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}
func lockContended(err error) bool {
	return errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)
}
