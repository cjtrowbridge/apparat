// Implements local Git actions without launching repository commands or hooks.
// Mutations are scoped to the selected file or a clean branch transition.
package projectgit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/utils/diff"
)

func GitAction(location string, q domain.WorkspaceRequest) (string, error) {
	meta := filepath.Join(location, ".git")
	info, err := os.Lstat(meta)
	if err != nil || !info.IsDir() {
		return "", errors.New("workspace Git requires a checkout with a local .git directory")
	}
	root, err := os.OpenRoot(location)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = root.Close()
	}()
	return GitOnFilesystems(osfs.New(meta, osfs.WithBoundOS()), osfs.New(location, osfs.WithBoundOS()), q, func(p string) (string, error) {
		return ReadText(root, p)
	})
}
func GitOnFilesystems(meta, work billy.Filesystem, q domain.WorkspaceRequest, read func(string) (string, error)) (string, error) {
	storage := filesystem.NewStorage(meta, cache.NewObjectLRUDefault())
	repo, err := git.Open(storage, work)
	if err != nil {
		return "", err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	switch q.Action {
	case "history", "commit-detail":
		return gitHistory(repo, q)
	case "branches":
		refs, err := repo.Branches()
		if err != nil {
			return "", err
		}
		var out strings.Builder
		err = refs.ForEach(func(ref *plumbing.Reference) error {
			if out.Len() > 8192 {
				return errors.New("branch list exceeds display limit")
			}
			fmt.Fprintln(&out, ref.Name().Short())
			return nil
		})
		return out.String(), err
	case "switch":
		ref := plumbing.NewBranchReferenceName(q.Argument)
		err = ref.Validate()
		if err != nil {
			return "", err
		}
		status, err := tree.Status()
		if err != nil || !status.IsClean() {
			return "", errors.New("switch requires a clean working tree and index")
		}
		err = tree.Checkout(&git.CheckoutOptions{Branch: ref})
		return "Switched to " + q.Argument, err
	case "status":
		status, err := tree.Status()
		if err != nil {
			return "", err
		}
		if status.IsClean() {
			return "Clean working tree", nil
		}
		return status.String(), nil
	}
	if !WorkspacePath(q.Path) || q.Path == "." {
		return "", errors.New("select a file first")
	}
	current, err := read(q.Path)
	if err != nil {
		return "", err
	}
	switch q.Action {
	case "diff":
		return gitDiff(repo, q.Path, current)
	case "stage":
		if Digest(current) != q.Base {
			return "", errors.New("file changed; reopen before staging")
		}
		_, err = tree.Add(q.Path)
		return "Staged " + q.Path, err
	case "unstage":
		err = tree.Restore(&git.RestoreOptions{Staged: true, Files: []string{q.Path}})
		return "Unstaged " + q.Path, err
	case "commit":
		status, err := tree.Status()
		if err != nil {
			return "", err
		}
		staged := false
		for path, state := range status {
			if state.Staging == git.Unmodified || state.Staging == git.Untracked {
				continue
			}
			if path != q.Path {
				return "", errors.New("another file is staged; commit it separately or unstage it first")
			}
			staged = true
		}
		if !staged || strings.TrimSpace(q.Argument) == "" {
			return "", errors.New("stage the selected file and enter a commit message")
		}
		hash, err := tree.Commit(q.Argument, &git.CommitOptions{Author: &object.Signature{Name: "Apparat", Email: "apparat@localhost", When: time.Now()}})
		return "Commit " + hash.String(), err
	}
	return "", errors.New("unknown Git operation")
}
func gitHistory(repo *git.Repository, q domain.WorkspaceRequest) (string, error) {
	if q.Action == "commit-detail" {
		if len(q.Argument) != 40 || !plumbing.IsHash(q.Argument) {
			return "", errors.New("enter a full commit ID from history")
		}
		c, err := repo.CommitObject(plumbing.NewHash(q.Argument))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s\n%s\n%s", c.Hash, c.Author.Name, c.Message), nil
	}
	history, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return "", err
	}
	defer history.Close()
	var out strings.Builder
	for i := 0; i < 30; i++ {
		c, err := history.Next()
		if err != nil {
			break
		}
		title, _, _ := strings.Cut(c.Message, "\n")
		fmt.Fprintf(&out, "%s %s\n", c.Hash, title)
	}
	return out.String(), nil
}
func gitDiff(repo *git.Repository, path, current string) (string, error) {
	var original string
	head, err := repo.Head()
	if err == nil {
		c, err := repo.CommitObject(head.Hash())
		if err != nil {
			return "", err
		}
		file, err := c.File(path)
		if err == nil {
			if file.Size > domain.WorkspaceTextLimit {
				return "", errors.New("base file exceeds text limit")
			}
			original, err = file.Contents()
			if err != nil || !ValidText(original) {
				return "", errors.New("base file is not supported text")
			}
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return "", err
		}
	} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
		return "", err
	}
	var out strings.Builder
	for _, change := range diff.DoWithTimeout(original, current, time.Second) {
		prefix := " "
		switch change.Type {
		case -1:
			prefix = "-"
		case 1:
			prefix = "+"
		}
		for line := range strings.SplitAfterSeq(change.Text, "\n") {
			out.WriteString(prefix)
			out.WriteString(line)
			if !strings.HasSuffix(line, "\n") {
				out.WriteByte('\n')
			}
		}
	}
	return out.String(), nil
}
