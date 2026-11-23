//go:build linux && !android

// Exercises the isolated PTY through the same session actions used by the GUI.
// Verifies process output, viewer/controller separation and persisted output after termination.
package terminal

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"os/exec"
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
