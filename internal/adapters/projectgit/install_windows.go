//go:build windows

// Publishes completed checkouts through a native exclusive rename.
// Concurrent checkout attempts cannot overwrite an existing destination.
package projectgit

import "golang.org/x/sys/windows"

func installClone(source, dest string) error {
	from, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(dest)
	if err != nil {
		return err
	}
	return windows.MoveFile(from, to)
}
