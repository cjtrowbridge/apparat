//go:build !linux

// Retains a conservative fallback on unqualified platforms.
// Existing working trees remain accessible through the registry.
package projectgit

import (
	"errors"
	"os"
)

func moveWorkspaceFile(*os.Root, string, *os.Root, string) error {
	return errors.New("safe move unavailable on this platform")
}
