// Compares one working directory with the index and its matching HEAD tree.
// Folder browsing never opens the contents of sibling or descendant directories.
package projectgit

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type directoryState struct {
	head          object.TreeEntry
	index         *index.Entry
	working       os.FileInfo
	indexedFolder bool
}

func headDirectory(repo *git.Repository, dir string) (*object.Tree, error) {
	head, err := repo.Head()
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil || dir == "." {
		return tree, err
	}
	tree, err = tree.Tree(dir)
	if errors.Is(err, object.ErrDirectoryNotFound) {
		return nil, nil
	}
	return tree, err
}

func directoryStatus(ctx context.Context, repo *git.Repository, meta, work billy.Filesystem, dir string) (string, error) {
	if !WorkspacePath(dir) {
		return "", os.ErrPermission
	}
	children := map[string]*directoryState{}
	child := func(name string) *directoryState {
		if children[name] == nil {
			children[name] = &directoryState{}
		}
		return children[name]
	}
	tree, err := headDirectory(repo, dir)
	if err != nil {
		return "", err
	}
	if tree != nil {
		for _, e := range tree.Entries {
			child(e.Name).head = e
		}
	}
	idx, err := repo.Storer.Index()
	if err != nil {
		return "", err
	}
	prefix := ""
	if dir != "." {
		prefix = dir + "/"
	}
	for _, entry := range idx.Entries {
		relative, ok := strings.CutPrefix(entry.Name, prefix)
		if !ok {
			continue
		}
		name, _, nested := strings.Cut(relative, "/")
		st := child(name)
		if nested {
			st.indexedFolder = true
		} else {
			st.index = entry
		}
	}
	entries, err := work.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	for _, e := range entries {
		child(e.Name()).working = e
	}
	matcher, err := directoryIgnores(meta, work, dir)
	if err != nil {
		return "", err
	}
	view := domain.WorkspaceView{}
	for name, st := range children {
		err := ctx.Err()
		if err != nil {
			return "", err
		}
		full := path.Join(dir, name)
		if !WorkspacePath(full) {
			continue
		}
		change, err := st.change(ctx, work, full, matcher)
		if err != nil {
			return "", err
		}
		if change.Marker != "" {
			view.Changes = append(view.Changes, change)
		}
	}
	sort.Slice(view.Changes, func(i, j int) bool {
		return view.Changes[i].Path < view.Changes[j].Path
	})
	data, err := json.Marshal(view)
	return string(data), err
}

func (s directoryState) change(ctx context.Context, work billy.Filesystem, name string, ignores gitignore.Matcher) (domain.WorkspaceChange, error) {
	ch := domain.WorkspaceChange{Path: name}
	if s.working != nil && !s.working.Mode().IsRegular() {
		return ch, nil
	}
	if s.working == nil {
		ch.Marker = "D"
		ch.Directory = s.indexedFolder || s.head.Mode == filemode.Dir
		ch.Staged = s.index == nil
		ch.Unstaged = s.index != nil
		return ch, nil
	}
	if s.index == nil {
		if s.head.Mode != filemode.Empty && s.head.Mode != filemode.Dir {
			ch.Marker = "D"
			ch.Staged = true
		} else if !ignores.Match(strings.Split(name, "/"), false) {
			ch.Marker = "U"
		}
		ch.Unstaged = true
		return ch, nil
	}
	ch.Conflict = s.index.Stage != 0
	ch.Staged = s.index.Hash != s.head.Hash || s.index.Mode != s.head.Mode || ch.Conflict
	if s.head.Mode == filemode.Empty {
		ch.Marker = "U"
		return ch, nil
	}
	if ch.Staged {
		ch.Marker = "C"
		return ch, nil
	}
	mode, err := filemode.NewFromOSFileMode(s.working.Mode())
	if err != nil {
		return ch, err
	}
	changed := mode != s.index.Mode || (s.index.Size != 0 && int64(s.index.Size) != s.working.Size())
	if !changed {
		f, err := work.Open(name)
		if err != nil {
			return ch, err
		}
		defer f.Close()
		hash := plumbing.NewHasher(plumbing.BlobObject, s.working.Size())
		_, err = io.Copy(hash, directoryReader{ctx, f})
		if err != nil {
			return ch, err
		}
		changed = hash.Sum() != s.index.Hash
	}
	if changed {
		ch.Marker = "C"
		ch.Unstaged = true
	}
	return ch, nil
}

func directoryIgnores(meta, work billy.Filesystem, dir string) (gitignore.Matcher, error) {
	var patterns []gitignore.Pattern
	read := func(fsys billy.Filesystem, name string, scope []string) error {
		info, err := fsys.Lstat(name)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		f, err := fsys.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()
		lines := bufio.NewScanner(f)
		for lines.Scan() {
			line := lines.Text()
			if line != "" && !strings.HasPrefix(line, "#") {
				patterns = append(patterns, gitignore.ParsePattern(line, scope))
			}
		}
		return lines.Err()
	}
	err := read(meta, "info/exclude", nil)
	if err != nil {
		return nil, err
	}
	var parts []string
	if dir != "." {
		parts = strings.Split(dir, "/")
	}
	for i := 0; i <= len(parts); i++ {
		scope := parts[:i]
		err = read(work, path.Join(strings.Join(scope, "/"), ".gitignore"), scope)
		if err != nil {
			return nil, err
		}
	}
	return gitignore.NewMatcher(patterns), nil
}

type directoryReader struct {
	ctx    context.Context
	source io.Reader
}

func (r directoryReader) Read(b []byte) (int, error) {
	err := r.ctx.Err()
	if err != nil {
		return 0, err
	}
	return r.source.Read(b)
}
