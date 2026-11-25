// Owns restricted interactive processes and their independent controller leases.
// PTY output updates a VT screen; detached clients can resume from the current screen.
package terminal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/hinshun/vt10x"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"sync"
	"time"
)

type Session struct {
	mu                               sync.Mutex
	id, owner, project, control, dir string
	file                             *os.File
	cmd                              *exec.Cmd
	vt                               vt10x.Terminal
	tracked                          bool
	ended                            bool
	revision                         uint64
	chunk                            int64
	bytes                            int64
}
type Host struct {
	pages    projectgit.DirectoryPages
	mu       sync.Mutex
	sessions map[string]*Session
	workers  sync.WaitGroup
	closing  bool
}

func (h *Host) Close() {
	h.mu.Lock()
	h.closing = true
	h.pages.Close()
	for _, s := range h.sessions {
		s.stop()
	}
	h.mu.Unlock()
	h.workers.Wait()
}
func (s *Session) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ended {
		_ = s.cmd.Process.Kill()
	}
}
func (h *Host) Do(ctx context.Context, dir, location, user, device string, q domain.WorkspaceRequest, alive func() bool) (v domain.WorkspaceView, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closing {
		return v, errors.New("terminal host closed")
	}
	dir = filepath.Join(dir, q.ProjectID, user)
	chat := strings.HasPrefix(q.Action, "chat-terminal-")
	if chat {
		if !domain.ValidID(q.Thread) {
			return v, errors.New("select a chat session")
		}
		dir = filepath.Join(dir, "chats", q.Thread)
		q.Action = strings.TrimPrefix(q.Action, "chat-")
	}
	if chat && q.Action == "terminal-observe" {
		q.Action = "terminal-screen"
	}
	if chat && q.Action == "terminal-permissions" {
		mode, err := chatPermission(dir)
		return domain.WorkspaceView{Detail: mode}, err
	}
	if chat && q.Action == "terminal-permission-set" {
		err := setChatPermission(dir, q.Argument)
		return domain.WorkspaceView{Detail: q.Argument}, err
	}
	tracked := chat && q.Action == "terminal-approved"
	if tracked {
		mode, err := chatPermission(dir)
		allowed := mode == "yolo" && !q.Approved || mode == "always-ask" && q.Approved
		if err != nil || !allowed || q.Argument == "" || len(q.Argument) > 4096 || strings.ContainsRune(q.Argument, 0) || !q.Deadline.After(time.Now()) || time.Until(q.Deadline) > 30*time.Minute {
			return v, errors.New("command denied by session permissions or deadline")
		}
	}
	if q.Action == "terminal-list" {
		v.Report = prerequisite()
		page, e := h.pages.Page(dir+"/"+device, dir, ".", q.Argument)
		if os.IsNotExist(e) {
			return v, nil
		}
		if e != nil {
			return v, e
		}
		v.Cursor = page.Cursor
		for _, item := range page.Entries {
			if !item.Directory || !domain.ValidID(item.Name) {
				continue
			}
			state := "Ended"
			if h.sessions[filepath.Join(dir, item.Name)] != nil {
				state = "Running"
			}
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: item.Name, Label: item.Name + " · " + state, State: state})
		}
		return v, nil
	}
	if !domain.ValidID(q.ID) {
		return v, errors.New("select a terminal session")
	}
	dir = filepath.Join(dir, q.ID)
	s := h.sessions[dir]
	if s != nil && (s.owner != user || s.project != q.ProjectID) {
		return v, os.ErrPermission
	}
	if tracked || chat && q.Action == "terminal-result" {
		_, err := os.Stat(dir)
		if err == nil {
			prior, e := commandStatus(dir, q)
			if e != nil {
				return v, e
			}
			if s != nil {
				live := s.screen()
				if live.Detail == "Running" {
					return live, nil
				}
			}
			return prior, nil
		}
		if !os.IsNotExist(err) || !tracked {
			return v, err
		}
	}
	if q.Action == "terminal-start" || tracked {
		if s == nil && chat && q.ID == q.Thread {
			_, err := os.Stat(dir)
			if err == nil {
				data, readErr := os.ReadFile(filepath.Join(dir, "screen"))
				return domain.WorkspaceView{Report: string(data), Detail: "Ended"}, readErr
			}
			if !os.IsNotExist(err) {
				return v, err
			}
		}
		if s != nil {
			return s.screen(), nil
		}
		reason := prerequisite()
		if reason != "" {
			return v, errors.New(reason)
		}
		if len(h.sessions) >= 8 {
			return v, errors.New("active terminal limit reached")
		}
		err = os.MkdirAll(filepath.Dir(dir), 0700)
		if err == nil {
			err = os.Mkdir(dir, 0700)
		}
		if err != nil {
			return v, err
		}
		command := ""
		if tracked {
			err = saveCommand(dir, q)
			if err != nil {
				return v, err
			}
			command = q.Argument
		}
		f, cmd, err := start(location, command)
		if err != nil {
			_ = os.Remove(dir)
			return v, err
		}
		s = &Session{tracked: tracked, id: q.ID, owner: user, project: q.ProjectID, control: device, dir: dir, file: f, cmd: cmd, vt: vt10x.New(vt10x.WithSize(80, 24))}
		if h.sessions == nil {
			h.sessions = map[string]*Session{}
		}
		h.sessions[dir] = s
		h.workers.Add(1)
		go h.supervise(s, alive)
		return s.screen(), nil
	}
	if q.Action == "terminal-history" {
		return history(dir, q.Argument)
	}
	if s == nil {
		if q.Action != "terminal-screen" {
			return v, errors.New("session has ended")
		}
		data, err := os.ReadFile(filepath.Join(dir, "screen"))
		return domain.WorkspaceView{Report: string(data), Detail: "Ended"}, err
	}
	if q.Action == "terminal-screen" {
		return s.screen(), nil
	}
	if q.Action == "terminal-take" {
		s.control = device
		return s.screen(), nil
	}
	if s.control != device {
		return v, errors.New("take terminal control first")
	}
	switch q.Action {
	case "terminal-stop":
		s.stop()
	case "terminal-input":
		if len(q.Argument) > 4096 {
			return v, errors.New("terminal input too large")
		}
		_ = s.file.SetWriteDeadline(time.Now().Add(time.Second))
		_, err = io.WriteString(s.file, q.Argument)
	case "terminal-resize":
		err = s.resize(q.Argument)
	default:
		return v, errors.New("unknown terminal operation")
	}
	return s.screen(), err
}
func (s *Session) screen() domain.WorkspaceView {
	s.mu.Lock()
	defer s.mu.Unlock()
	detail := "Running"
	if s.ended {
		detail = "Ended"
	}
	return domain.WorkspaceView{Report: s.vt.String(), Detail: detail, Cursor: strconv.FormatUint(s.revision, 10)}
}
func (s *Session) resize(value string) error {
	var cols, rows int
	_, err := fmt.Sscanf(value, "%dx%d", &cols, &rows)
	if err != nil || cols < 20 || cols > 200 || rows < 5 || rows > 80 {
		return errors.New("unsupported terminal size")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err = setSize(s.file, cols, rows)
	if err == nil {
		s.vt.Resize(cols, rows)
		s.revision++
	}
	return err
}
func (h *Host) supervise(s *Session, alive func() bool) {
	defer h.workers.Done()
	done := make(chan struct{})
	go func() {
		defer close(done)
		reader := bufio.NewReader(io.TeeReader(s.file, &scrollback{dir: s.dir}))
		for {
			err := s.vt.Parse(reader)
			if err != nil {
				break
			}
			s.mu.Lock()
			s.revision++
			now := time.Now()
			// Persist at most one visual checkpoint per second while output is changing.
			chunk := now.Unix()
			if s.chunk != chunk {
				s.chunk = chunk
				screen := s.vt.String()
				_ = os.WriteFile(filepath.Join(s.dir, "screen"), []byte(screen), 0600)
			}
			s.mu.Unlock()
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			waitErr := s.cmd.Wait()
			_ = s.file.Close()
			s.mu.Lock()
			s.ended = true
			_ = os.WriteFile(filepath.Join(s.dir, "screen"), []byte(s.vt.String()), 0600)
			if s.tracked {
				status := "Exited 0"
				if waitErr != nil {
					status = fmt.Sprintf("Exited %d", s.cmd.ProcessState.ExitCode())
				}
				_ = os.WriteFile(filepath.Join(s.dir, "exit"), []byte(status), 0600)
			}
			s.mu.Unlock()
			h.mu.Lock()
			delete(h.sessions, s.dir)
			h.mu.Unlock()
			return
		case <-ticker.C:
			if !alive() {
				s.stop()
			}
		}
	}
}
func history(dir, cursor string) (domain.WorkspaceView, error) {
	text, next, err := readOutput(dir, cursor)
	return domain.WorkspaceView{Report: text, Cursor: next}, err
}
