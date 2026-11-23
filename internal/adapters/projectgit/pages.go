// Streams one directory through short-lived, caller-bound page cursors.
// At most 64 pinned directory handles are retained and idle cursors close automatically.
package projectgit

import (
	"crypto/rand"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"io"
	"os"
	"sync"
	"time"
)

type directoryCursor struct {
	owner, path string
	file        *os.File
	timer       *time.Timer
}
type DirectoryPages struct {
	mu      sync.Mutex
	cursors map[string]*directoryCursor
}

func (p *DirectoryPages) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for key, c := range p.cursors {
		c.timer.Stop()
		c.file.Close()
		delete(p.cursors, key)
	}
}
func (p *DirectoryPages) Page(owner, location, path, token string) (v domain.WorkspaceView, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	c := p.cursors[token]
	if token != "" {
		if c == nil || c.owner != owner || c.path != path {
			return v, errors.New("directory cursor expired; reopen folder")
		}
		c.timer.Stop()
		delete(p.cursors, token)
	} else {
		if len(p.cursors) >= 64 {
			return v, errors.New("directory cursor limit reached")
		}
		root, err := os.OpenRoot(location)
		if err != nil {
			return v, err
		}
		defer root.Close()
		dir, err := workspaceDirectory(root, path)
		if err != nil {
			return v, err
		}
		defer dir.Close()
		file, err := dir.Open(".")
		if err != nil {
			return v, err
		}
		c = &directoryCursor{owner: owner, path: path, file: file}
	}
	entries, err := c.file.ReadDir(64)
	for _, e := range entries {
		if WorkspacePath(e.Name()) && (e.IsDir() || e.Type().IsRegular()) {
			v.Entries = append(v.Entries, domain.WorkspaceEntry{Name: e.Name(), Directory: e.IsDir()})
		}
	}
	if err == io.EOF || len(entries) < 64 || err != nil {
		c.file.Close()
		if err == io.EOF {
			err = nil
		}
		return v, err
	}
	if p.cursors == nil {
		p.cursors = map[string]*directoryCursor{}
	}
	key := rand.Text()
	c.timer = time.AfterFunc(time.Minute, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.cursors[key] == c {
			delete(p.cursors, key)
			c.file.Close()
		}
	})
	p.cursors[key] = c
	v.Cursor = key
	return v, nil
}
