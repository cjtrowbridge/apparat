//go:build linux && !android

// Exercises the isolated PTY through the same session actions used by the GUI.
// Verifies process output, viewer/controller separation and persisted output after termination.
package terminal

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsolatedSession(t *testing.T) {
	if prerequisite() != "" {
		t.Skip(prerequisite())
	}
	probe := exec.Command("/usr/bin/bwrap", "--unshare-all", "--unshare-user", "--disable-userns", "--ro-bind", "/", "/", "/bin/true")
	if probe.Run() != nil {
		t.Skip("host disallows isolated user namespaces")
	}
	h := Host{}
	defer h.Close()
	root := t.TempDir()
	project := t.TempDir()
	id, _ := domain.NewID(time.Now(), nil)
	q := domain.WorkspaceRequest{ProjectID: id, ID: id, Action: "terminal-start"}
	call := func(device string) (domain.WorkspaceView, error) {
		return h.Do(context.Background(), root, project, "owner", device, q, func() bool {
			return true
		})
	}
	_, err := call("desktop")
	if err != nil {
		t.Fatal(err)
	}
	q.Action = "terminal-input"
	q.Argument = "printf 'marker-%d' 9184\r"
	_, err = call("phone")
	if err == nil {
		t.Fatal("viewer sent input")
	}
	q.Action = "terminal-take"
	_, err = call("phone")
	if err != nil {
		t.Fatal(err)
	}
	q.Action = "terminal-input"
	_, err = call("phone")
	if err != nil {
		t.Fatal(err)
	}
	q.Action = "terminal-screen"
	deadline := time.Now().Add(5 * time.Second)
	for {
		view, err := call("desktop")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(view.Report, "marker-9184") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal(view.Report)
		}
		time.Sleep(20 * time.Millisecond)
	}
	q.Action = "terminal-resize"
	q.Argument = "100x28"
	_, err = call("phone")
	if err != nil {
		t.Fatal(err)
	}
	q.Action = "terminal-stop"
	_, err = call("phone")
	if err != nil {
		t.Fatal(err)
	}
	q.Action = "terminal-screen"
	for time.Now().Before(deadline) {
		v, err := call("desktop")
		if err == nil && v.Detail == "Ended" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal(os.ErrDeadlineExceeded)
}

func TestChatTerminalIsolation(t *testing.T) {
	root := t.TempDir()
	project, _ := domain.NewID(time.Now(), nil)
	chat, _ := domain.NewID(time.Now(), nil)
	other, _ := domain.NewID(time.Now(), nil)
	id, _ := domain.NewID(time.Now(), nil)
	writeScreen := func(dir, contents string) {
		t.Helper()
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(filepath.Join(dir, "screen"), []byte(contents), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	base := filepath.Join(root, project, "owner")
	writeScreen(filepath.Join(base, id), "legacy")
	writeScreen(filepath.Join(base, "chats", chat, id), "conversation")
	h := Host{}
	defer h.Close()
	call := func(action, thread string) (domain.WorkspaceView, error) {
		q := domain.WorkspaceRequest{ProjectID: project, ID: id, Action: action, Thread: thread}
		return h.Do(context.Background(), root, "", "owner", "desktop", q, func() bool {
			return true
		})
	}
	for _, tc := range []struct{ action, thread, want string }{
		{"terminal-screen", "", "legacy"}, {"chat-terminal-screen", chat, "conversation"},
	} {
		v, err := call(tc.action, tc.thread)
		if err != nil || v.Report != tc.want {
			t.Fatalf("%s: %q, %v", tc.action, v.Report, err)
		}
	}
	for _, thread := range []string{other, "", "../" + chat} {
		_, err := call("chat-terminal-screen", thread)
		if err == nil {
			t.Fatal("read screen outside chat")
		}
	}
	v, err := call("chat-terminal-list", chat)
	if err != nil || len(v.Choices) != 1 || v.Choices[0].ID != id {
		t.Fatal(v, err)
	}
	v, err = call("chat-terminal-list", other)
	if err != nil || len(v.Choices) != 0 {
		t.Fatal(v, err)
	}
}

func TestApprovedChatCommand(t *testing.T) {
	if prerequisite() != "" {
		t.Skip(prerequisite())
	}
	probe := exec.Command("/usr/bin/bwrap", "--unshare-all", "--unshare-user", "--disable-userns", "--ro-bind", "/", "/", "/bin/true")
	if probe.Run() != nil {
		t.Skip("host disallows isolated user namespaces")
	}
	root, project := t.TempDir(), t.TempDir()
	h := Host{}
	defer h.Close()
	thread, _ := domain.NewID(time.Now(), nil)
	job, _ := domain.NewID(time.Now(), nil)
	q := domain.WorkspaceRequest{ProjectID: thread, Thread: thread, ID: job, Deadline: time.Now().Add(time.Minute), Approved: true}
	command := "printf 'verified terminal result\\n'; exit 7"
	call := func(action, arg string) (domain.WorkspaceView, error) {
		q.Action = "chat-terminal-" + action
		q.Argument = arg
		return h.Do(context.Background(), root, project, "user", "device", q, func() bool {
			return true
		})
	}
	v, err := call("permissions", "")
	if err != nil || v.Detail != "read-only" {
		t.Fatal(v, err)
	}
	_, err = call("approved", command)
	if err == nil {
		t.Fatal("read-only accepted a command")
	}
	_, err = call("permission-set", "always-ask")
	if err != nil {
		t.Fatal(err)
	}
	q.Approved = false
	_, err = call("approved", command)
	if err == nil {
		t.Fatal("accepted missing approval")
	}
	q.Approved = true
	_, err = call("approved", command)
	if err != nil {
		t.Fatal(err)
	}
	_, err = call("approved", "echo changed")
	if err == nil {
		t.Fatal("accepted changed command for same invocation")
	}
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); {
		v, err = call("result", command)
		if err != nil {
			t.Fatal(err)
		}
		if v.Detail == "Exited 7" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if v.Detail != "Exited 7" || !strings.Contains(v.Report, "verified terminal result") {
		t.Fatal(v)
	}
	v, err = call("approved", command)
	if err != nil || v.Detail != "Exited 7" {
		t.Fatal("did not return the existing receipt", v, err)
	}
	_, err = call("permission-set", "read-only")
	if err != nil {
		t.Fatal(err)
	}
	_, err = call("approved", command)
	if err == nil {
		t.Fatal("accepted revoked permission")
	}
	q.Approved = false
	q.ID, _ = domain.NewID(time.Now(), nil)
	_, err = call("approved", command)
	if err == nil {
		t.Fatal("read-only accepted automatic dispatch")
	}
	_, err = call("permission-set", "yolo")
	if err != nil {
		t.Fatal(err)
	}
	_, err = call("approved", command)
	if err != nil {
		t.Fatal("yolo failed to start", err)
	}
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); {
		v, err = call("result", command)
		if err != nil {
			t.Fatal(err)
		}
		if v.Detail == "Exited 7" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if v.Detail != "Exited 7" || !strings.Contains(v.Report, "verified terminal result") {
		t.Fatal(v)
	}
	_, err = call("permission-set", "always-ask")
	if err != nil {
		t.Fatal(err)
	}
	_, err = call("approved", command)
	if err == nil {
		t.Fatal("accepted stale automatic dispatch after switching to Always Ask")
	}

}
