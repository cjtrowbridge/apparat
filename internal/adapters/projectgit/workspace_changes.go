// Produces a structured changed-file list and the exact base reviewed in a diff.
// The same representation travels through native and Android document workspaces.
package projectgit

import (
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"os"
	"sort"
)

func workspaceChanges(tree *git.Worktree) (string, error) {
	states, err := tree.Status()
	if err != nil {
		return "", err
	}
	var changes []domain.WorkspaceChange
	for path, s := range states {
		if !WorkspacePath(path) {
			continue
		}
		changes = append(changes, domain.WorkspaceChange{Path: path, Staged: s.Staging != git.Unmodified && s.Staging != git.Untracked, Unstaged: s.Worktree != git.Unmodified, Conflict: s.Staging == git.UpdatedButUnmerged || s.Worktree == git.UpdatedButUnmerged})
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
	data, err := json.Marshal(domain.WorkspaceView{Changes: changes})
	return string(data), err
}
func workspaceReview(repo *git.Repository, work billy.Filesystem, path string, read func(string) (string, error)) (string, error) {
	if !WorkspacePath(path) || path == "." {
		return "", errors.New("choose a changed file")
	}
	value, err := read(path)
	base := Digest(value)
	if err != nil {
		_, statErr := work.Lstat(path)
		if !os.IsNotExist(statErr) {
			return "", err
		}
		base = "deleted"
	}
	report, err := gitDiff(repo, path, value)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(domain.WorkspaceView{Base: base, Report: report})
	return string(data), err
}
