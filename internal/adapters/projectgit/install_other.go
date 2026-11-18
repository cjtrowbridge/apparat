//go:build !linux

// Retains a conservative fallback on unqualified platforms.
// Existing working trees remain accessible through the registry.
package projectgit

import "errors"

func installClone(source, dest string) error {
	return errors.New("clone publication unsupported on this platform")
}
