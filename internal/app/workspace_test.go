// Verifies editable working trees, selection isolation and durable workspace records.
// Uses real Git checkouts under disposable roots without network access.
package app_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func gitFixtureCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@localhost", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@localhost")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fixture Git: %s %v", out, err)
	}
	return strings.TrimSpace(string(out))
}
func TestLocalWorkspaceOffline(t *testing.T) {
	ctx := context.Background()
	store := t.TempDir()
	first := t.TempDir()
	second := filepath.Join(t.TempDir(), "second")
	gitFixtureCommand(t, first, "init", "-b", "main")
	requireSuccess(t, os.WriteFile(filepath.Join(first, "example.txt"), []byte("first\n"), 0600))
	requireSuccess(t, os.Mkdir(filepath.Join(first, ".apparat"), 0700))
	spec := `{"version":1,"tasks":[{"id":"snapshot","label":"Save example","enabled":true,"action":"snapshot-text","path":"example.txt"}]}`
	requireSuccess(t, os.WriteFile(filepath.Join(first, ".apparat/tasks.json"), []byte(spec), 0600))
	gitFixtureCommand(t, first, "add", ".")
	gitFixtureCommand(t, first, "commit", "-m", "base")
	gitFixtureCommand(t, first, "clone", first, second)
	prior := gitFixtureCommand(t, second, "rev-parse", "HEAD")
	requireSuccess(t, os.WriteFile(filepath.Join(second, "example.txt"), []byte("other index\n"), 0600))
	gitFixtureCommand(t, second, "add", "example.txt")
	node := nodeFor(t, store, false)
	requireSuccess(t, node.Initialize(ctx))
	defer func() {
		_ = node.Close()
	}()
	requireSuccess(t, node.StoreProject(ctx, "", first, ""))
	requireSuccess(t, node.StoreProject(ctx, "", second, ""))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	q := domain.WorkspaceRequest{Path: "example.txt"}
	for _, p := range rows {
		if p.Location == first {
			q.ProjectID = p.ID
		}
	}
	do := func(action string) domain.WorkspaceView {
		t.Helper()
		q.Action = action
		result, err := node.Workspace(ctx, q)
		requireSuccess(t, err)
		return result
	}
	q.Base = do("read").Base
	q.Text = "draft text\n"
	do("draft")
	requireSuccess(t, os.WriteFile(filepath.Join(first, "example.txt"), []byte("external edit\n"), 0600))
	q.ID, err = domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	q.Action = "save"
	_, err = node.Workspace(ctx, q)
	if err == nil {
		t.Fatal("overwrote competing edit")
	}
	requireSuccess(t, node.Close())
	node = nodeFor(t, store, false)
	requireSuccess(t, node.Initialize(ctx))
	got := do("read")
	if got.Text != q.Text || !strings.Contains(got.Report, "conflict") {
		t.Fatal("draft conflict lost")
	}
	q.Base = do("rebase").Base
	q.ID, err = domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	saved := do("save")
	do("save")
	q.Base = saved.Base
	if !strings.Contains(do("diff").Report, "+draft text") {
		t.Fatal("missing working diff")
	}
	do("stage")
	q.Argument = "commit selected file"
	do("commit")
	if gitFixtureCommand(t, first, "show", "HEAD:example.txt") != "draft text" {
		t.Fatal("wrong committed contents")
	}
	if gitFixtureCommand(t, second, "rev-parse", "HEAD") != prior || gitFixtureCommand(t, second, "show", ":example.txt") != "other index" {
		t.Fatal("other checkout was changed")
	}
	q.ID, err = domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	q.Argument = "snapshot"
	do("run")
	expected := "Verified example.txt\ndraft text\n"
	if do("artifact").Report != expected {
		t.Fatal("snapshot readback failed")
	}
	do("run")
	q.ID, err = domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	q.Argument = "pending message"
	do("post")
	if !strings.Contains(do("chat").Report, q.Argument) {
		t.Fatal("message missing")
	}
	if !strings.Contains(do("history").Report, "commit selected file") {
		t.Fatal("commit missing")
	}
	q.Path = ".git/config"
	q.Action = "read"
	_, err = node.Workspace(ctx, q)
	if err == nil {
		t.Fatal("exposed Git internals")
	}
}
