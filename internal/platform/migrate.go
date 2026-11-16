// Migrates one unambiguous legacy runtime root without merging data.
// A parent migration lock serializes inspection and atomic rename.
package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func MigrateRoot(root string, legacy []string) error {
	if len(legacy) == 0 {
		return nil
	}
	lock, err := Acquire(filepath.Join(filepath.Dir(root), ".migration"), "migration", "migration")
	if err != nil {
		return err
	}
	defer func() {
		_ = lock.Close()
	}()
	candidates := []string{}
	for _, path := range legacy {
		entries, err := os.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			candidates = append(candidates, path)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) > 1 {
		return errors.New("both legacy roots contain state; select one explicitly with --runtime-dir or archive each before migration; nothing was merged")
	}
	_, err = os.Stat(root)
	if err == nil {
		return errors.New("new and legacy node roots both exist; use --runtime-dir to select explicitly; nothing was overwritten")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// Hold legacy ownership too: never rename a running older node using this lock contract.
	owner, err := Acquire(candidates[0], "migration", "migration")
	if err != nil {
		return err
	}
	defer func() {
		_ = owner.Close()
	}()
	err = os.Rename(candidates[0], root)
	if err != nil {
		return fmt.Errorf("automatic node migration unavailable; stop all Apparat processes, back up and move the selected legacy root to %s, then retry: %w", root, err)
	}
	return nil
}
