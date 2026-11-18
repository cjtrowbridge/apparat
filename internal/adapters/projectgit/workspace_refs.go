// Publishes a completed reference via a Git lockfile.
// Provider lock or write failures must leave the prior branch tip intact.
package projectgit

import (
	"errors"
	git "github.com/go-git/go-git/v5"
	"io"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type workspaceStorage struct {
	*filesystem.Storage
	meta     billy.Filesystem
	expected *plumbing.Reference
}

func (s *workspaceStorage) SetReference(ref *plumbing.Reference) error {
	old := s.expected
	if old != nil && old.Name() != ref.Name() {
		return errors.New("branch changed during operation")
	}
	return s.CheckAndSetReference(ref, old)
}
func (s *workspaceStorage) CheckAndSetReference(ref, old *plumbing.Reference) error {
	name := ref.Name()
	if name != plumbing.HEAD && !name.IsBranch() {
		return errors.New("workspace can only update local branches or HEAD")
	}
	if name != plumbing.HEAD {
		err := name.Validate()
		if err != nil {
			return err
		}
	}
	target := name.String()
	lock := target + ".lock"
	file, err := s.meta.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return errors.New("branch lock unavailable")
	}
	pending := true
	defer func() {
		if pending {
			_ = s.meta.Remove(lock)
		}
	}()
	if old != nil {
		current, err := s.Reference(name)
		if err != nil || current.String() != old.String() {
			_ = file.Close()
			return errors.New("branch moved during operation")
		}
	}
	line := ref.Strings()[1] + "\n"
	_, err = io.WriteString(file, line)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	err = s.meta.Rename(lock, target)
	pending = err != nil
	return err
}

// Preserve both lines of ancestry when completing a resolved merge.
func workspaceMergeParents(repo *git.Repository, meta billy.Filesystem, head *plumbing.Reference) ([]plumbing.Hash, error) {
	for _, marker := range []string{"rebase-apply", "rebase-merge", "CHERRY_PICK_HEAD", "REVERT_HEAD", "sequencer"} {
		_, err := meta.Stat(marker)
		if err == nil {
			return nil, errors.New("finish the current Git sequence before a workspace commit")
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	f, err := meta.Open("MERGE_HEAD")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	data, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(data) > 4096 || head == nil {
		return nil, errors.New("invalid merge state")
	}
	parents := []plumbing.Hash{head.Hash()}
	for _, value := range strings.Fields(string(data)) {
		if !plumbing.IsHash(value) {
			return nil, errors.New("invalid merge parent")
		}
		hash := plumbing.NewHash(value)
		_, err = repo.CommitObject(hash)
		if err != nil {
			return nil, err
		}
		parents = append(parents, hash)
	}
	if len(parents) < 2 {
		return nil, errors.New("missing merge parent")
	}
	return parents, nil
}
