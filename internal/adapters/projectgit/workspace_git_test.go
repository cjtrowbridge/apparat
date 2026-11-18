// Reproduces the external-provider ref-publication failure using disposable Git storage.
// Checks that staging and ref changes preserve unrelated data on failure.
package projectgit

import (
	"errors"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"os"
	"path/filepath"
	"testing"
)

type failingRename struct{ billy.Filesystem }

func (f failingRename) Rename(string, string) error {
	return errors.New("simulated provider failure")
}
func TestReferenceReplacementFailurePreservesOriginal(t *testing.T) {
	dir := t.TempDir()
	fs := failingRename{osfs.New(dir, osfs.WithBoundOS())}
	s := workspaceStorage{Storage: filesystem.NewStorage(fs, cache.NewObjectLRUDefault()), meta: fs}
	name := plumbing.NewBranchReferenceName("main")
	old := plumbing.NewHashReference(name, plumbing.NewHash("1111111111111111111111111111111111111111"))
	requirePath := filepath.Join(dir, "refs", "heads", "main")
	err := os.MkdirAll(filepath.Dir(requirePath), 0700)
	if err != nil {
		t.Fatal(err)
	}
	expected := old.Hash().String() + "\n"
	err = os.WriteFile(requirePath, []byte(expected), 0600)
	if err != nil {
		t.Fatal(err)
	}
	fresh := plumbing.NewHashReference(name, plumbing.NewHash("2222222222222222222222222222222222222222"))
	err = s.CheckAndSetReference(fresh, old)
	if err == nil {
		t.Fatal("expected provider failure")
	}
	b, err := os.ReadFile(requirePath)
	if err != nil || string(b) != expected {
		t.Fatal("damaged previous ref", err)
	}
	_, err = os.Stat(requirePath + ".lock")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("leaked lock")
	}
}
