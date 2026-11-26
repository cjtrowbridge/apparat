//go:build linux && !android

// Exercises local account access, shell selection, yolo auto-dispatch and receipt replay with a disposable checkout.
package terminal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func TestLocalHostShellGrantsAndRecovery(t *testing.T) {
	root, checkout := t.TempDir(), t.TempDir()
	h := Host{}
	defer h.Close()
	newID := func() string {
		id, err := domain.NewID(time.Now(), nil)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	thread, shell, job := newID(), newID(), newID()
	q := domain.WorkspaceRequest{ProjectID: "global", Thread: thread, ID: shell}
	call := func(device, action, arg string) (domain.WorkspaceView, error) {
		q.Action = "chat-terminal-" + action
		q.Argument = arg
		return h.Do(context.Background(), root, checkout, "user", device, q, func() bool {
			return true
		})
	}
	_, err := call("peer", "start", "host-account")
	if err == nil {
		t.Fatal("worker grant opened a host-account shell")
	}
	_, err = call("local", "start", "host-account")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"take", "input", "screen", "stop", "history"} {
		_, err = call("peer", action, "")
		if err == nil {
			t.Fatal("peer accessed host-account terminal", action)
		}
	}
	v, err := call("local", "permissions", "")
	if err != nil || v.Detail != "read-only" || v.Text != shell {
		t.Fatal(v, err)
	}
	ready := filepath.Join(root, "global", "user", "chats", thread, shell, "integration", "ready")
	until := time.Now().Add(5 * time.Second)
	for time.Now().Before(until) {
		b, _ := os.ReadFile(ready)
		if len(b) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_, err = call("local", "permission-set", "yolo")
	if err != nil {
		t.Fatal(err)
	}
	q.RelatedRun = shell
	q.ID = job
	q.Deadline = time.Now().Add(time.Minute)
	command := "printf x >> receipt-test; sleep 0.1; pwd"
	v, err = call("local", "shared-command", command)
	if err != nil || v.Detail != "Running" {
		t.Fatal("yolo automatic command did not run on the host account", v, err)
	}
	for time.Now().Before(until) {
		v, err = call("local", "result", command)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(v.Detail, "Exited") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if v.Detail != "Exited 0" {
		t.Fatal(v)
	}
	_, err = call("local", "permission-set", "always-ask")
	if err != nil {
		t.Fatal(err)
	}
	q.Approved = false
	_, err = call("local", "shared-command", command)
	if err == nil {
		t.Fatal("always-ask accepted a command without approval")
	}
	q.Approved = true
	v, err = call("local", "shared-command", command)
	if err != nil || v.Detail != "Exited 0" {
		t.Fatal(v, err)
	}
	for time.Now().Before(until) {
		v, err = call("local", "result", command)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(v.Detail, "Exited") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if v.Detail != "Exited 0" {
		t.Fatal(v)
	}
	v, err = call("local", "shared-command", command)
	if err != nil || v.Detail != "Exited 0" {
		t.Fatal(v, err)
	}
	data, err := os.ReadFile(filepath.Join(checkout, "receipt-test"))
	if err != nil || string(data) != "x" {
		t.Fatal("command execution count", string(data), err)
	}
	q.RelatedRun = newID()
	_, err = call("local", "shared-command", command)
	if err == nil {
		t.Fatal("replayed receipt against a different terminal")
	}
}
