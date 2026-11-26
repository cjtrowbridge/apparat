// Tracks commands in the selected interactive shell through a private readline binding.
// Shell integration runs only on an empty editing line and records exit independently of the prompt.
package terminal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
func prepareShell(dir string, personal bool) error {
	control := filepath.Join(dir, "integration")
	err := os.Mkdir(control, 0700)
	if err != nil {
		return err
	}
	mount := "/run/apparat"
	if personal {
		mount = control
	}
	rc := "APPARAT_CONTROL=" + shellQuote(mount) + "\n" + `
PS1='$ '
PROMPT_COMMAND='builtin printf "%s" "$((++APPARAT_PROMPT))" > "$APPARAT_CONTROL/ready"'
_apparat_dispatch() {
    [[ -r "$APPARAT_CONTROL/request" ]] || return
    builtin source "$APPARAT_CONTROL/request"
    [[ "$APPARAT_DONE" != "$APPARAT_JOB" ]] || return
    APPARAT_DONE=$APPARAT_JOB
    if [[ -n "$READLINE_LINE" || "${EPOCHREALTIME/[.,]/}" -ge "$APPARAT_LIMIT" ]]; then
        builtin printf '%s busy' "$APPARAT_JOB" > "$APPARAT_CONTROL/status"
        return
    fi
    builtin printf '\n$ %s\n' "$APPARAT_COMMAND"
    builtin eval -- "$APPARAT_COMMAND"
    builtin printf '%s %s' "$APPARAT_DONE" "$?" > "$APPARAT_CONTROL/status"
}
bind -x '"\e[99~":_apparat_dispatch'
`
	for _, name := range []string{"status", "ready"} {
		err = os.WriteFile(filepath.Join(control, name), nil, 0600)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(control, "bashrc"), []byte(rc), 0600)
}

func (h *Host) sharedCommand(chat string, q domain.WorkspaceRequest, device string, submit bool) (domain.WorkspaceView, error) {
	var v domain.WorkspaceView
	if !domain.ValidID(q.RelatedRun) {
		return v, errors.New("select a terminal for this command")
	}
	if !canAccessShell(filepath.Join(chat, q.RelatedRun), device) {
		return v, os.ErrPermission
	}
	jobs := filepath.Join(chat, "jobs")
	job := filepath.Join(jobs, q.ID)
	shell := h.sessions[filepath.Join(chat, q.RelatedRun)]
	_, err := os.Stat(job)
	if err == nil {
		result, err := commandStatus(job, q)
		if err != nil || shell == nil || shell.job != job {
			return result, err
		}
		return h.refreshShared(shell), nil
	}
	if !os.IsNotExist(err) || !submit {
		return v, err
	}
	if shell == nil || shell.control != device {
		return v, errors.New("the selected terminal must be running and controlled by this device")
	}
	if shell.job != "" {
		h.refreshShared(shell)
		if shell.job != "" {
			return v, errors.New("terminal is already running an agent command")
		}
	}
	control := filepath.Join(shell.dir, "integration")
	ready, _ := os.ReadFile(filepath.Join(control, "ready"))
	if len(ready) == 0 || string(ready) == shell.dirtyPrompt {
		return v, errors.New("terminal is busy or has unfinished input")
	}
	err = os.MkdirAll(jobs, 0700)
	if err == nil {
		err = os.Mkdir(job, 0700)
	}
	if err == nil {
		err = saveCommand(job, q)
	}
	if err != nil {
		return v, err
	}
	request := fmt.Sprintf("APPARAT_JOB=%s\nAPPARAT_LIMIT=%d\nAPPARAT_COMMAND=%s\n", shellQuote(q.ID), q.Deadline.UnixMicro(), shellQuote(q.Argument))
	err = os.WriteFile(filepath.Join(control, "next"), []byte(request), 0600)
	if err == nil {
		err = os.Rename(filepath.Join(control, "next"), filepath.Join(control, "request"))
	}
	if err != nil {
		return v, err
	}
	shell.job = job
	_ = shell.file.SetWriteDeadline(time.Now().Add(time.Second))
	_, err = io.WriteString(shell.file, "\x1b[99~")
	return shell.screen(), err
}

func (h *Host) refreshShared(s *Session) domain.WorkspaceView {
	v := s.screen()
	data, _ := os.ReadFile(filepath.Join(s.dir, "integration", "status"))
	fields := strings.Fields(string(data))
	if len(fields) != 2 || fields[0] != filepath.Base(s.job) {
		return v
	}
	code, err := strconv.Atoi(fields[1])
	v.Detail = "Command declined by busy or expired terminal"
	if err == nil && code >= 0 && code <= 255 {
		v.Detail = fmt.Sprintf("Exited %d", code)
	}
	if os.WriteFile(filepath.Join(s.job, "screen"), []byte(v.Report), 0600) == nil && os.WriteFile(filepath.Join(s.job, "exit"), []byte(v.Detail), 0600) == nil {
		s.job = ""
	}
	return v
}

func canAccessShell(dir, device string) bool {
	if device == "local" {
		return true
	}
	_, err := os.Stat(filepath.Join(dir, "host-account"))
	return os.IsNotExist(err)
}
