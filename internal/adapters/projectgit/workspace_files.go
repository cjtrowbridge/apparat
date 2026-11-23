// Exposes bounded working-tree text and directory operations.
// Writes use staged sibling files and compare against the editor's base digest.
package projectgit

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func WorkspacePath(path string) bool {
	if !fs.ValidPath(path) || strings.ContainsAny(path, "\\\x00\r\n") {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if strings.EqualFold(part, ".git") || strings.HasPrefix(part, ".apparat-write-") {
			return false
		}
	}
	return true
}

func Digest(text string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
}

func ValidText(text string) bool {
	return len(text) <= domain.WorkspaceTextLimit && utf8.ValidString(text) && !strings.ContainsRune(text, 0)
}

func ReadText(root *os.Root, path string) (string, error) {
	if !WorkspacePath(path) {
		return "", os.ErrPermission
	}
	parent, err := workspaceDirectory(root, filepath.Dir(path))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = parent.Close()
	}()
	return readLeaf(parent, filepath.Base(path))
}
func readLeaf(root *os.Root, path string) (string, error) {
	info, err := root.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > domain.WorkspaceTextLimit {
		return "", errors.New("choose a regular text file of at most 64 KiB")
	}
	file, err := root.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return "", errors.New("file changed while opening")
	}
	data, err := io.ReadAll(io.LimitReader(file, domain.WorkspaceTextLimit+1))
	if err != nil || !ValidText(string(data)) {
		return "", errors.New("file exceeds text limits or contains binary data")
	}
	return string(data), nil
}
func Files(location string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	if !WorkspacePath(q.Path) {
		return v, errors.New("invalid workspace path")
	}
	root, err := os.OpenRoot(location)
	if err != nil {
		return v, err
	}
	defer func() {
		_ = root.Close()
	}()
	if q.Action == "list" {
		folder, err := workspaceDirectory(root, q.Path)
		if err != nil {
			return v, err
		}
		defer func() {
			_ = folder.Close()
		}()
		dir, err := folder.Open(".")
		if err != nil {
			return v, err
		}
		defer func() {
			_ = dir.Close()
		}()
		items, err := dir.ReadDir(2049)
		if err != nil && err != io.EOF {
			return v, err
		}
		if len(items) > 2048 {
			return v, errors.New("directory exceeds the 2048 item limit")
		}
		for _, item := range items {
			if WorkspacePath(item.Name()) && item.Type()&os.ModeSymlink == 0 && (item.IsDir() || item.Type().IsRegular()) {
				v.Entries = append(v.Entries, domain.WorkspaceEntry{Name: item.Name(), Directory: item.IsDir()})
			}
		}
		sort.Slice(v.Entries, func(i, j int) bool {
			return v.Entries[i].Name < v.Entries[j].Name
		})
		return v, nil
	}
	parent, err := workspaceDirectory(root, filepath.Dir(q.Path))
	if err != nil {
		return v, err
	}
	defer func() {
		_ = parent.Close()
	}()
	name := filepath.Base(q.Path)
	if q.Action == "mkdir" {
		return v, parent.Mkdir(name, 0700)
	}
	if q.Action == "create" {
		f, err := parent.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return v, err
		}
		err = f.Close()
		if err != nil {
			return v, err
		}
	}
	content, err := readLeaf(parent, name)
	if err != nil {
		return v, err
	}
	v.Editing = true
	v.Text = content
	v.Base = Digest(content)
	if q.Action == "read" || q.Action == "create" {
		return v, nil
	}
	if q.Base == "" || v.Base != q.Base {
		if q.Action == "save" && content == q.Text {
			return v, nil
		}
		return v, errors.New("file changed since opening; your draft is retained")
	}
	switch q.Action {
	case "save":
		if !ValidText(q.Text) {
			return v, errors.New("draft exceeds text limits")
		}
		info, err := parent.Lstat(name)
		if err != nil {
			return v, err
		}
		temp := ".apparat-write-" + rand.Text()
		file, err := parent.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return v, err
		}
		defer func() {
			_ = parent.Remove(temp)
		}()
		_, err = file.WriteString(q.Text)
		if err == nil {
			err = file.Chmod(info.Mode().Perm())
		}
		if err == nil {
			err = file.Sync()
		}
		closeErr := file.Close()
		if err != nil {
			return v, err
		}
		if closeErr != nil {
			return v, closeErr
		}
		current, err := readLeaf(parent, name)
		if err != nil || current != content {
			return v, errors.New("checkout changed during save")
		}
		err = parent.Rename(temp, name)
		v.Text = q.Text
		v.Base = Digest(q.Text)
		return v, err
	case "move":
		if !WorkspacePath(q.Argument) {
			return v, errors.New("invalid destination")
		}
		dest, err := workspaceDirectory(root, filepath.Dir(q.Argument))
		if err != nil {
			return v, err
		}
		defer func() {
			_ = dest.Close()
		}()
		return domain.WorkspaceView{}, moveWorkspaceFile(parent, name, dest, filepath.Base(q.Argument))

	case "delete":
		return domain.WorkspaceView{}, parent.Remove(name)
	}
	return v, errors.New("unknown file operation")
}
