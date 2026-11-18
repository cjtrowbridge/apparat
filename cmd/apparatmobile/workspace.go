//go:build android && gui

// Operates in the selected document tree without importing the checkout.
// Text replacement preserves the old document until the replacement is published.
package apparatmobile

import (
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (f treeFS) text(name string) (string, error) {
	if !projectgit.WorkspacePath(name) {
		return "", os.ErrPermission
	}
	info, err := f.Stat(name)
	if err != nil || info.IsDir() || info.Size() > domain.WorkspaceTextLimit {
		return "", errors.New("choose readable UTF-8 text up to 64 KiB")
	}
	file, err := f.Open(name)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()
	b, err := io.ReadAll(io.LimitReader(file, domain.WorkspaceTextLimit+1))
	if err != nil || !projectgit.ValidText(string(b)) {
		return "", errors.New("unsupported document text")
	}
	return string(b), nil
}
func (d projectDocuments) WorkspaceFile(uri string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	host := currentDocuments()
	if host == nil {
		return v, errors.New("reselect the folder to restore access")
	}
	f := treeFS{host: host, uri: uri, base: "."}
	if !projectgit.WorkspacePath(q.Path) {
		return v, os.ErrPermission
	}
	switch q.Action {
	case "list":
		entries, err := f.ReadDir(q.Path)
		if err != nil {
			return v, err
		}
		for _, e := range entries {
			if projectgit.WorkspacePath(e.Name()) {
				v.Entries = append(v.Entries, domain.WorkspaceEntry{Name: e.Name(), Directory: e.IsDir()})
			}
		}
		return v, nil
	case "mkdir":
		_, err = f.call("newdir", q.Path, "")
		return v, err
	case "create":
		_, err = f.call("create", q.Path, "")
		if err != nil {
			return v, err
		}
	case "read", "save", "delete", "move":
	default:
		meta, err := f.Chroot(".git")
		if err != nil {
			return v, err
		}
		v.Report, err = projectgit.GitOnFilesystems(meta, f, q, f.text)
		return v, err
	}
	text, err := f.text(q.Path)
	if err != nil {
		return v, err
	}
	v = domain.WorkspaceView{Text: text, Editing: true, Base: projectgit.Digest(text)}
	if q.Action == "read" || q.Action == "create" || q.Action == "save" && text == q.Text {
		return v, nil
	}
	if v.Base != q.Base {
		return v, errors.New("document changed; recover and compare the draft")
	}
	switch q.Action {
	case "delete":
		return domain.WorkspaceView{}, f.Remove(q.Path)
	case "move":
		if !projectgit.WorkspacePath(q.Argument) {
			return v, os.ErrPermission
		}
		_, err = f.call("rename", q.Path, q.Argument)
		return domain.WorkspaceView{}, err
	}
	temp := path.Join(path.Dir(q.Path), ".apparat-write-"+rand.Text())
	file, err := f.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return v, err
	}
	defer func() {
		_ = f.Remove(temp)
	}()
	_, err = io.WriteString(file, q.Text)
	closeErr := file.Close()
	if err != nil {
		return v, err
	}
	if closeErr != nil {
		return v, closeErr
	}
	again, err := f.text(q.Path)
	if err != nil || again != text {
		return v, errors.New("document changed during save")
	}
	err = f.replace(temp, q.Path)
	if err != nil {
		return v, err
	}
	v.Text = q.Text
	v.Base = projectgit.Digest(q.Text)
	return v, err
}
