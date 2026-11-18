// Pins each directory before operating on a checkout-relative leaf.
// Verifies the identity of every opened directory to reject link replacement.
package projectgit

import (
	"errors"
	"os"
	"strings"
)

func workspaceDirectory(root *os.Root, path string) (*os.Root, error) {
	if !WorkspacePath(path) {
		return nil, os.ErrPermission
	}
	parent, err := root.OpenRoot(".")
	if err != nil {
		return nil, err
	}
	if path == "." {
		return parent, nil
	}
	for _, part := range strings.Split(path, "/") {
		before, err := parent.Lstat(part)
		if err != nil || !before.IsDir() {
			_ = parent.Close()
			return nil, errors.New("folder missing or not a regular directory")
		}
		child, err := parent.OpenRoot(part)
		_ = parent.Close()
		if err != nil {
			return nil, err
		}
		opened, err := child.Stat(".")
		if err != nil || !os.SameFile(opened, before) {
			_ = child.Close()
			return nil, errors.New("directory changed while opening")
		}
		parent = child
	}
	return parent, nil
}
