//go:build linux

// Publishes a completed checkout atomically, including on Android.
// RENAME_NOREPLACE leaves other concurrent clones and user files intact.
package projectgit

import "golang.org/x/sys/unix"

func installClone(source, dest string) error {
	return unix.Renameat2(unix.AT_FDCWD, source, unix.AT_FDCWD, dest, unix.RENAME_NOREPLACE)
}
