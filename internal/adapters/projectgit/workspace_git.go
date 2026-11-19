// Implements local Git actions without launching repository commands or hooks.
// Mutations are scoped to the selected file or a clean branch transition.
package projectgit

import (
	"context"
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
	"github.com/pmezard/go-difflib/difflib"
)

func GitAction(location string, q domain.WorkspaceRequest) (string, error) {
	return GitActionContext(context.Background(), location, q)
}
func GitActionContext(ctx context.Context, location string, q domain.WorkspaceRequest) (string, error) {
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
	return GitOnFilesystems(ctx, osfs.New(meta, osfs.WithBoundOS()), osfs.New(location, osfs.WithBoundOS()), q, func(p string) (string, error) {
		return ReadText(root, p)
	})
}
func GitOnFilesystems(ctx context.Context, meta, work billy.Filesystem, q domain.WorkspaceRequest, read func(string) (string, error)) (string, error) {
	storage := &workspaceStorage{Storage: filesystem.NewStorage(meta, cache.NewObjectLRUDefault()), meta: meta}
	repo, err := git.Open(storage, work)
	if err != nil {
		return "", err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	switch q.Action {
	case "directory-status":
		return directoryStatus(ctx, repo, meta, work, q.Path)
	case "changes":
		return workspaceChanges(tree)
	case "review":
		return workspaceReview(repo, work, q.Path, read)
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
	var current string
	deletion := q.Action == "stage-deletion" || q.Action == "commit" && q.Base == "deleted"
	if deletion {
		_, err = work.Lstat(q.Path)
		if !os.IsNotExist(err) {
			return "", errors.New("file must be absent before staging or committing its deletion")
		}
	} else if q.Action != "unstage" {
		current, err = read(q.Path)
		if err != nil {
			return "", err
		}
	}
	switch q.Action {
	case "diff":
		return gitDiff(repo, q.Path, current)
	case "conflict":
		return conflictReport(repo, q.Path)
	case "stage", "stage-deletion", "unstage", "resolve":
		return updateFileIndex(repo, meta, work, q, current)
	case "commit":
		lock, err := meta.OpenFile("index.lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return "", errors.New("index busy; retry when the other Git operation finishes")
		}
		defer func() {
			_ = lock.Close()
			_ = meta.Remove("index.lock")
		}()
		idx, err := repo.Storer.Index()
		if err != nil {
			return "", err
		}
		for _, entry := range idx.Entries {
			if entry.Stage != 0 {
				return "", errors.New("resolve remaining merge conflicts before committing")
			}
		}
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
				return "", fmt.Errorf("%s is also staged; review Changes and unstage it before committing %s", path, q.Path)
			}
			staged = true
		}
		if !staged || strings.TrimSpace(q.Argument) == "" {
			return "", errors.New("stage the selected file and enter a commit message")
		}
		expected, err := repo.Head()
		if err != nil && !errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", err
		}
		storage.expected = expected
		indexEntry, err := idx.Entry(q.Path)
		if deletion && indexEntry != nil {
			return "", errors.New("stage the deletion first")
		}
		if !deletion && (err != nil || Digest(current) != q.Base || indexEntry.Hash != plumbing.ComputeHash(plumbing.BlobObject, []byte(current))) {
			return "", errors.New("working file differs from the reviewed stage; reopen and stage the intended contents")
		}
		parents, err := workspaceMergeParents(repo, meta, expected)
		if err != nil {
			return "", err
		}
		hash, err := tree.Commit(q.Argument, &git.CommitOptions{Parents: parents, Author: &object.Signature{Name: "Apparat", Email: "apparat@localhost", When: time.Now()}})
		if err == nil && len(parents) > 1 {
			for _, name := range []string{"MERGE_HEAD", "MERGE_MSG", "MERGE_MODE", "AUTO_MERGE"} {
				clean := meta.Remove(name)
				if clean != nil && !os.IsNotExist(clean) {
					return "Commit " + hash.String() + " created; inspect remaining merge metadata", nil
				}
			}
		}
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
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return "No commits yet", nil
	}
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
	return difflib.GetUnifiedDiffString(difflib.UnifiedDiff{A: difflib.SplitLines(original), B: difflib.SplitLines(current), FromFile: path, ToFile: path, Context: 3})
}
