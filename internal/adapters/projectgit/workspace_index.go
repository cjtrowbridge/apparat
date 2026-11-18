// Stages exactly the bytes reviewed by the user and publishes a locked Git index.
// A storage failure must never turn an existing file into a staged deletion.
package projectgit

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/go-git/go-billy/v5"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func updateFileIndex(repo *git.Repository, meta, work billy.Filesystem, q domain.WorkspaceRequest, content string) (string, error) {
	lock, err := meta.OpenFile("index.lock", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", errors.New("index locked or unavailable")
	}
	pending := true
	defer func() {
		if pending {
			_ = meta.Remove("index.lock")
		}
	}()
	idx, err := repo.Storer.Index()
	if err == nil {
		err = editFileIndex(repo, work, idx, q, content)
	}
	if err == nil {
		err = index.NewEncoder(lock).Encode(idx)
	}
	closeErr := lock.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	err = meta.Rename("index.lock", "index")
	pending = err != nil
	if err != nil {
		return "", err
	}
	return "Updated index for " + q.Path, nil
}
func editFileIndex(repo *git.Repository, work billy.Filesystem, idx *index.Index, q domain.WorkspaceRequest, content string) error {
	var entry *index.Entry
	for _, it := range idx.Entries {
		if it.Name != q.Path {
			continue
		}
		if it.Stage != 0 && q.Action != "resolve" {
			return errors.New("file has unresolved merge stages; inspect conflict and explicitly resolve")
		}
		if it.Stage == 0 {
			copy := *it
			entry = &copy
		}
	}
	kept := idx.Entries[:0]
	for _, it := range idx.Entries {
		if it.Name != q.Path {
			kept = append(kept, it)
		}
	}
	idx.Entries = kept
	idx.Cache = nil
	if q.Action == "stage-deletion" {
		if entry == nil {
			return errors.New("file has no index entry to delete")
		}
		return nil
	}
	if q.Action == "unstage" {
		head, err := repo.Head()
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		commit, err := repo.CommitObject(head.Hash())
		if err != nil {
			return err
		}
		base, err := commit.File(q.Path)
		if errors.Is(err, object.ErrFileNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		idx.Entries = append(idx.Entries, &index.Entry{Name: q.Path, Hash: base.Hash, Mode: base.Mode})
		return nil
	}
	if Digest(content) != q.Base {
		return errors.New("file changed; reopen before staging")
	}
	stat, err := work.Lstat(q.Path)
	if err != nil || !stat.Mode().IsRegular() {
		return errors.New("stage requires a regular text file")
	}
	obj := &plumbing.MemoryObject{}
	obj.SetType(plumbing.BlobObject)
	w, err := obj.Writer()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, content)
	closeErr := w.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	hash, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("could not store file; index preserved: %w", err)
	}
	if entry == nil {
		entry = &index.Entry{Name: q.Path, Mode: filemode.Regular}
	}
	if stat.Mode().Perm()&0111 != 0 {
		entry.Mode = filemode.Executable
	}
	entry.Hash = hash
	entry.Size = uint32(len(content))
	entry.ModifiedAt = stat.ModTime()
	idx.Entries = append(idx.Entries, entry)
	return nil
}

func conflictReport(repo *git.Repository, path string) (string, error) {
	idx, err := repo.Storer.Index()
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, it := range idx.Entries {
		if it.Stage == 0 || it.Name != path {
			continue
		}
		fmt.Fprintf(&out, "Stage %d · %s\n", it.Stage, it.Hash)
		b, err := repo.BlobObject(it.Hash)
		if err != nil || b.Size > domain.WorkspaceTextLimit {
			continue
		}
		r, err := b.Reader()
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(io.LimitReader(r, domain.WorkspaceTextLimit+1))
		_ = r.Close()
		if err == nil && ValidText(string(data)) {
			out.Write(data)
			out.WriteByte('\n')
		}
	}
	if out.Len() == 0 {
		return "No unresolved merge stages for this file", nil
	}
	return out.String(), nil
}
