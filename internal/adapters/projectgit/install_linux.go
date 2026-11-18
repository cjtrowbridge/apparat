//go:build linux

// Publishes a completed checkout atomically, including on Android.
// RENAME_NOREPLACE leaves other concurrent clones and user files intact.
package projectgit

import (
	"golang.org/x/sys/unix"
	"os"
)

func installClone(source, dest string) error {
	return unix.Renameat2(unix.AT_FDCWD, source, unix.AT_FDCWD, dest, unix.RENAME_NOREPLACE)
}

func moveWorkspaceFile(from *os.Root, old string, to *os.Root, name string) error {
	src, err := from.Open(".")
	if err != nil {
		return err
	}
	defer func() {
		_ = src.Close()
	}()
	dest, err := to.Open(".")
	if err != nil {
		return err
	}
	defer func() {
		_ = dest.Close()
	}()
	return unix.Renameat2(int(src.Fd()), old, int(dest.Fd()), name, unix.RENAME_NOREPLACE)
}
