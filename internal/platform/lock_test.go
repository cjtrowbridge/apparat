// Exercises real node ownership and non-destructive legacy-root migration.
// Uses isolated temporary roots and never inspects user runtime data.
package platform

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLockContentionAndRelease(t *testing.T) {
	root := t.TempDir()
	first, err := Acquire(root, "test", "headless")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Acquire(root, "other", "gui")
	var contention *ContentionError
	if !errors.As(err, &contention) || contention.Owner.PID != os.Getpid() {
		t.Fatalf("contention: %v", err)
	}
	{
		err = first.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err = first.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	next, err := Acquire(root, "next", "headless")
	if err != nil {
		t.Fatal(err)
	}
	_ = next.Close()
}
func TestMigrationConflictAndAtomicMove(t *testing.T) {
	base := t.TempDir()
	gui := filepath.Join(base, "apparat")
	headless := filepath.Join(base, "apparatd")
	root := filepath.Join(base, "node")
	for _, path := range []string{gui, headless} {
		_ = os.Mkdir(path, 0700)
		_ = os.WriteFile(filepath.Join(path, "state"), []byte(path), 0600)
	}
	if MigrateRoot(root, []string{gui, headless}) == nil {
		t.Fatal("merged two populated roots")
	}
	_, err := os.Stat(filepath.Join(gui, "state"))
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Rename(headless, filepath.Join(base, "archived"))
	{
		err = MigrateRoot(root, []string{gui, headless})
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(root, "state"))
	if err != nil || string(data) != gui {
		t.Fatal("state not preserved", err)
	}
	{
		err = MigrateRoot(root, []string{gui, headless})
		if err != nil {
			t.Fatal("second migration", err)
		}
	}
}
func TestMigrationRefusesActiveLegacyOwner(t *testing.T) {
	base := t.TempDir()
	old := filepath.Join(base, "apparat")
	lock, err := Acquire(old, "legacy", "gui")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = lock.Close()
	}()
	if MigrateRoot(filepath.Join(base, "node"), []string{old}) == nil {
		t.Fatal("renamed an owned root")
	}
}

func TestHeadlessLegacyAndIndependentRoots(t *testing.T) {
	base := t.TempDir()
	old := filepath.Join(base, "apparatd")
	root := filepath.Join(base, "node")
	{
		err := os.Mkdir(old, 0700)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err := os.WriteFile(filepath.Join(old, "state"), []byte("preserved"), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err := MigrateRoot(root, []string{filepath.Join(base, "apparat"), old})
		if err != nil {
			t.Fatal(err)
		}
	}
	first, err := Acquire(root, "first", "gui")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = first.Close()
	}()
	second, err := Acquire(filepath.Join(base, "independent"), "second", "headless")
	if err != nil {
		t.Fatal(err)
	}
	_ = second.Close()
	// An interrupted/pre-existing destination must not overwrite the legacy source.
	{
		err = os.Mkdir(old, 0700)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err = os.WriteFile(filepath.Join(old, "state"), []byte("other"), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	if MigrateRoot(root, []string{old}) == nil {
		t.Fatal("overwrote occupied destination")
	}
	data, err := os.ReadFile(filepath.Join(old, "state"))
	if err != nil || string(data) != "other" {
		t.Fatal("lost legacy state", err)
	}
}
