//go:build android && gui

// Implements go-git's filesystem interface through one Android tree grant.
// All document lookups descend from that grant, including Git metadata reads.
package apparatmobile

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"golang.org/x/sys/unix"
)

type treeFS struct {
	host      Documents
	uri, base string
}
type docInfo struct {
	Label     string `json:"name"`
	Directory bool
	Bytes     int64 `json:"size"`
	Modified  int64
}

func (i docInfo) Name() string {
	return i.Label
}
func (i docInfo) Size() int64 {
	return i.Bytes
}
func (i docInfo) Mode() os.FileMode {
	if i.Directory {
		return os.ModeDir | 0755
	}
	return 0644
}
func (i docInfo) ModTime() time.Time {
	return time.UnixMilli(i.Modified)
}
func (i docInfo) IsDir() bool {
	return i.Directory
}
func (i docInfo) Sys() any {
	return nil
}

type docFile struct{ *os.File }

func (f docFile) Lock() error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX)
}
func (f docFile) Unlock() error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func (f treeFS) name(p string) (string, error) {
	if p == "" {
		p = "."
	}
	if !fs.ValidPath(p) || strings.Contains(p, "\\") {
		return "", os.ErrPermission
	}
	return path.Join(f.base, p), nil
}
func (f treeFS) call(op, p, arg string) (string, error) {
	name, err := f.name(p)
	if err != nil {
		return "", err
	}
	result := f.host.Access(f.uri, op, name, arg)
	switch result {
	case "!missing":
		return "", os.ErrNotExist
	case "!exists":
		return "", os.ErrExist
	case "!provider", "":
		return "", errors.New("folder provider cannot complete " + op)
	}
	return result, nil
}
func (f treeFS) Root() string {
	return f.base
}
func (f treeFS) Join(parts ...string) string {
	return path.Join(parts...)
}
func (f treeFS) Chroot(p string) (billy.Filesystem, error) {
	base, err := f.name(p)
	f.base = base
	return f, err
}
func (f treeFS) ReadDir(p string) ([]os.FileInfo, error) {
	raw, err := f.call("list", p, "")
	if err != nil {
		return nil, err
	}
	var data []docInfo
	err = json.Unmarshal([]byte(raw), &data)
	if err != nil {
		return nil, err
	}
	var entries []os.FileInfo
	for _, info := range data {
		if !fs.ValidPath(info.Label) || strings.Contains(info.Label, "/") || info.Label == "." {
			return nil, errors.New("invalid provider file name")
		}
		entries = append(entries, info)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	return entries, nil
}
func (f treeFS) Stat(p string) (os.FileInfo, error) {
	raw, err := f.call("stat", p, "")
	if err != nil {
		return nil, err
	}
	var info docInfo
	err = json.Unmarshal([]byte(raw), &info)
	return info, err
}
func (f treeFS) Lstat(p string) (os.FileInfo, error) {
	return f.Stat(p)
}
func (f treeFS) MkdirAll(p string, _ os.FileMode) error {
	_, err := f.call("mkdir", p, "")
	return err
}
func (f treeFS) OpenFile(p string, flag int, perm os.FileMode) (billy.File, error) {
	mode := "r"
	if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
		mode = "rw"
	}
	if flag&os.O_CREATE != 0 {
		err := f.MkdirAll(path.Dir(p), 0755)
		if err != nil {
			return nil, err
		}
		_, err = f.call("create", p, "")
		if err != nil && (!errors.Is(err, os.ErrExist) || flag&os.O_EXCL != 0) {
			return nil, err
		}
	}
	raw, err := f.call("open", p, mode)
	if err != nil {
		return nil, err
	}
	fd, err := strconv.Atoi(raw)
	if err != nil || fd < 0 {
		return nil, errors.New("provider returned an invalid descriptor")
	}
	file := os.NewFile(uintptr(fd), p)
	if flag&os.O_TRUNC != 0 {
		err = file.Truncate(0)
	}
	if err == nil && flag&os.O_APPEND != 0 {
		_, err = file.Seek(0, 2)
	}
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return docFile{file}, nil
}
func (f treeFS) Create(p string) (billy.File, error) {
	return f.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
}
func (f treeFS) Open(p string) (billy.File, error) {
	return f.OpenFile(p, os.O_RDONLY, 0)
}
func (f treeFS) TempFile(dir, prefix string) (billy.File, error) {
	return f.OpenFile(path.Join(dir, prefix+rand.Text()), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
}
func (f treeFS) Remove(p string) error {
	_, err := f.call("remove", p, "")
	return err
}
func (f treeFS) Rename(from, to string) error {
	dest, err := f.name(to)
	if err != nil {
		return err
	}
	err = f.MkdirAll(path.Dir(to), 0755)
	if err != nil {
		return err
	}
	_, err = f.Stat(to)
	if err == nil {
		return f.replace(from, to)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err = f.call("rename", from, dest)
	return err
}
func (f treeFS) Symlink(string, string) error {
	return errors.New("document folders do not provide symbolic links")
}
func (f treeFS) Readlink(string) (string, error) {
	return "", os.ErrInvalid
}

func (f treeFS) replace(from, to string) error {
	dest, err := f.name(to)
	if err != nil {
		return err
	}
	_, err = f.call("replace", from, dest)
	return err
}
