//go:build !linux && !windows && !darwin

// Fails safely on targets without a qualified exclusive checkout publication primitive.
// Existing working directories remain untouched.
package projectgit

import "errors"

func installClone(source, dest string) error {
	return errors.New("clone publication unsupported on this platform")
}
