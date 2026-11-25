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
	personal                         bool
	job, dirtyPrompt                 string
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
		selected, _ := os.ReadFile(selectionFile(dir, device))
		return domain.WorkspaceView{Detail: mode, Text: string(selected)}, err
	}
	if chat && q.Action == "terminal-permission-set" {
		err := setChatPermission(dir, q.Argument)
		return domain.WorkspaceView{Detail: q.Argument}, err
	}
	tracked := chat && q.Action == "terminal-shared-command"
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
			if !item.Directory || !domain.ValidID(item.Name) || !canAccessShell(filepath.Join(dir, item.Name), device) {
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
	if !canAccessShell(dir, device) {
		return v, os.ErrPermission
	}
	s := h.sessions[dir]
	if s != nil && (s.owner != user || s.project != q.ProjectID) {
		return v, os.ErrPermission
	}
	if tracked || chat && q.Action == "terminal-result" {
		if !tracked && q.RelatedRun == "" {
			return commandStatus(dir, q)
		}
		return h.sharedCommand(filepath.Dir(dir), q, device, tracked)
	}
	if q.Action == "terminal-start" {
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
		personal := q.Argument == "host-account" && chat && device == "local"
		if q.Argument == "host-account" && !personal {
			return v, os.ErrPermission
		}
		reason := prerequisite()
		if reason != "" && !personal {
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
		if personal {
			err = os.WriteFile(filepath.Join(dir, "host-account"), nil, 0600)
			if err != nil {
				return v, err
			}
		}
		err = prepareShell(dir, personal)
		if err != nil {
			return v, err
		}
		f, cmd, err := start(location, dir, personal)
		if err != nil {
			_ = os.RemoveAll(dir)
			return v, err
		}
		s = &Session{personal: personal, id: q.ID, owner: user, project: q.ProjectID, control: device, dir: dir, file: f, cmd: cmd, vt: vt10x.New(vt10x.WithSize(80, 24), vt10x.WithWriter(f))}
		if h.sessions == nil {
			h.sessions = map[string]*Session{}
		}
		h.sessions[dir] = s
		if chat {
			_ = os.WriteFile(selectionFile(filepath.Dir(dir), device), []byte(q.ID), 0600)
		}
		h.workers.Add(1)
		go h.supervise(s, alive)
		return s.screen(), nil
	}
	if q.Action == "terminal-history" {
		return history(dir, q.Argument)
	}
	if chat && q.Action == "terminal-screen" && q.Argument == "select" {
		_, err = os.Stat(dir)
		if err != nil {
			return v, err
		}
		err = os.WriteFile(selectionFile(filepath.Dir(dir), device), []byte(q.ID), 0600)
		if err != nil {
			return v, err
		}
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
		ready, _ := os.ReadFile(filepath.Join(s.dir, "integration", "ready"))
		s.dirtyPrompt = string(ready)
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
	environment := "Isolated shell · read-only checkout · offline"
	if s.personal {
		environment = "Local host account"
	}
	s.vt.Lock()
	cursor := s.vt.Cursor()
	caret := ""
	if s.vt.CursorVisible() {
		caret = fmt.Sprintf("%d,%d", cursor.X, cursor.Y)
	}
	s.vt.Unlock()
	lines := strings.Split(s.vt.String(), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	report := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	return domain.WorkspaceView{Text: caret, ComparedBase: environment, Report: report, Detail: detail, Cursor: strconv.FormatUint(s.revision, 10)}
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
			_ = s.cmd.Wait()
			_ = s.file.Close()
			s.mu.Lock()
			s.ended = true
			_ = os.WriteFile(filepath.Join(s.dir, "screen"), []byte(s.vt.String()), 0600)
			s.mu.Unlock()
			h.mu.Lock()
			if s.job != "" {
				h.refreshShared(s)
				if s.job != "" {
					_ = os.WriteFile(filepath.Join(s.job, "screen"), []byte(s.vt.String()), 0600)
					_ = os.WriteFile(filepath.Join(s.job, "exit"), []byte("Shell ended before command completion"), 0600)
				}
			}
			delete(h.sessions, s.dir)
			h.mu.Unlock()
			return
		case <-ticker.C:
			h.mu.Lock()
			if s.job != "" {
				h.refreshShared(s)
			}
			h.mu.Unlock()
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
