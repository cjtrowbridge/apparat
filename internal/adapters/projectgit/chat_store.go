// Persists whole transcript revisions under the registered checkout.
// Pins each directory and stages a complete sibling before publishing a change.
package projectgit

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"os"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

// ChatFile returns the transcript basename using the session's UTC timestamp.
func ChatFile(id string) (string, error) {
	t, err := domain.ChatTime(id)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02-15-04-05") + "-" + id + ".md", nil
}

func chatDirectory(location string, create bool) (*os.Root, error) {
	parent, err := os.OpenRoot(location)
	if err != nil {
		return nil, err
	}
	for _, part := range []string{".apparat", "chats"} {
		if create {
			err = parent.Mkdir(part, 0700)
			if err != nil && !errors.Is(err, os.ErrExist) {
				_ = parent.Close()
				return nil, err
			}
		}
		if !create {
			_, err = parent.Lstat(part)
			if err != nil {
				_ = parent.Close()
				return nil, err
			}
		}
		child, err := workspaceDirectory(parent, part)
		_ = parent.Close()
		if err != nil {
			return nil, err
		}
		parent = child
	}
	return parent, nil
}

func openChatLeaf(dir *os.Root, name string) (*os.File, error) {
	info, err := dir.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("expected a regular transcript file")
	}
	f, err := dir.Open(name)
	if err != nil {
		return nil, err
	}

	current, err := f.Stat()
	if err != nil || !os.SameFile(current, info) {
		_ = f.Close()
		return nil, errors.New("transcript changed while opening")
	}
	return f, nil
}

func chatRead(dir *os.Root, name string) ([]byte, error) {
	f, err := openChatLeaf(dir, name)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return io.ReadAll(f)
}

// WithChat reads one revision. The owner serializes local writers.
// An absent file is accepted only when an initial header is supplied.
func WithChat(location, id string, initial *domain.ChatHeader, apply func(*domain.ChatSession) (bool, error)) (s domain.ChatSession, err error) {
	name, err := ChatFile(id)
	if err != nil {
		return s, err
	}
	dir, err := chatDirectory(location, initial != nil)
	if err != nil {
		return s, err
	}
	defer func() {
		_ = dir.Close()
	}()
	raw, err := chatRead(dir, name)
	fresh := errors.Is(err, os.ErrNotExist) && initial != nil
	if fresh {
		s.ChatHeader = *initial
		if s.ID != id {
			return s, errors.New("session identity mismatch")
		}
		err = s.Validate()
		if err != nil {
			return s, err
		}
	} else {
		if err != nil {
			return s, err
		}
		s, err = parseChat(raw, id)
		if err != nil {
			return s, err
		}
		if initial != nil && (initial.Scope != s.Scope || initial.Description != s.Description) {
			return s, errors.New("session ID already has different metadata")
		}
	}
	changed := false
	if apply != nil {
		changed, err = apply(&s)
		if err != nil {
			return domain.ChatSession{}, err
		}
	}
	if !fresh && !changed {
		return s, nil
	}
	data, err := chatBytes(s)
	if err != nil {
		return s, err
	}
	stage := ".apparat-write-" + rand.Text()
	f, err := dir.OpenFile(stage, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return s, err
	}
	defer func() {
		_ = dir.Remove(stage)
	}()
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return s, err
	}
	if fresh {
		// Publish without replacing an existing transcript.
		err = dir.Link(stage, name)
		if err != nil && !errors.Is(err, os.ErrExist) {
			err = moveWorkspaceFile(dir, stage, dir, name)
		}
	} else {
		current, readErr := chatRead(dir, name)
		if readErr != nil || !bytes.Equal(current, raw) {
			return s, errors.New("transcript changed during update")
		}
		err = dir.Rename(stage, name)
	}
	return s, err
}
