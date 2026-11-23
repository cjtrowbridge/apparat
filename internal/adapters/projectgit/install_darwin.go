//go:build darwin

// Publishes completed checkouts through a native exclusive rename.
// Concurrent checkout attempts cannot overwrite an existing destination.
package projectgit

import "golang.org/x/sys/unix"

func installClone(source, dest string) error {
	return unix.RenamexNp(source, dest, unix.RENAME_EXCL)
}
