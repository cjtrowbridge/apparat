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

func checkedPath(root *os.Root, path string) error {
	if !WorkspacePath(path) {
		return errors.New("choose a path inside this working tree")
	}
	prefix := "."
	for _, part := range strings.Split(path, "/") {
		prefix = filepath.Join(prefix, part)
		info, err := root.Lstat(prefix)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("workspace links cannot be followed")
		}
	}
	return nil
}
func ReadText(root *os.Root, path string) (string, error) {
	err := checkedPath(root, path)
	if err != nil {
		return "", err
	}
	info, err := root.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > domain.WorkspaceTextLimit {
		return "", errors.New("choose a regular text file of at most 64 KiB")
	}
	file, err := root.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()
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
		err = checkedPath(root, q.Path)
		if err != nil {
			return v, err
		}
		dir, err := root.Open(q.Path)
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
	if q.Action == "create" {
		err = checkedPath(root, filepath.Dir(q.Path))
		if err != nil {
			return v, err
		}
		f, err := root.OpenFile(q.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return v, err
		}
		err = f.Close()
		if err != nil {
			return v, err
		}
	}
	content, err := ReadText(root, q.Path)
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
		info, err := root.Stat(q.Path)
		if err != nil {
			return v, err
		}
		temp := filepath.Join(filepath.Dir(q.Path), ".apparat-write-"+rand.Text())
		file, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return v, err
		}
		defer func() {
			_ = root.Remove(temp)
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
		current, err := ReadText(root, q.Path)
		if err != nil || current != content {
			return v, errors.New("checkout changed during save")
		}
		err = root.Rename(temp, q.Path)
		v.Text = q.Text
		v.Base = Digest(q.Text)
		return v, err
	case "move":
		if !WorkspacePath(q.Argument) {
			return v, errors.New("invalid destination")
		}
		err = checkedPath(root, filepath.Dir(q.Argument))
		if err != nil {
			return v, err
		}
		// Link first so an occupied destination can never be overwritten.
		err = root.Link(q.Path, q.Argument)
		if err != nil {
			return v, err
		}
		return domain.WorkspaceView{}, root.Remove(q.Path)
	case "delete":
		return domain.WorkspaceView{}, root.Remove(q.Path)
	}
	return v, errors.New("unknown file operation")
}
