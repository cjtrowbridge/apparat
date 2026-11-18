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
	q := domain.WorkspaceRequest{Path: "example.txt", Approved: true}
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
	q.Argument = do("compare").ComparedBase
	requireSuccess(t, os.WriteFile(filepath.Join(first, "example.txt"), []byte("newer change\n"), 0600))
	q.Action = "rebase"
	_, err = node.Workspace(ctx, q)
	if err == nil {
		t.Fatal("rebased onto unreviewed contents")
	}
	q.Argument = do("compare").ComparedBase
	q.Base = do("rebase").Base
	q.ID, err = domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	saved := do("save")
	do("save")
	q.Base = saved.Base
	if !strings.Contains(do("diff").Report, "+draft text") {
		t.Fatal("missing working diff")
	}
	delta := do("changes").Changes
	if len(delta) != 1 || delta[0].Path != q.Path || delta[0].Marker != "C" || !delta[0].Unstaged {
		t.Fatal("changed path missing from Changes")
	}
	reviewed := do("review")
	if reviewed.Base != q.Base || !strings.Contains(reviewed.Report, "@@") {
		t.Fatal("review does not describe current file")
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
	requireSuccess(t, node.Close())
	node = nodeFor(t, store, false)
	requireSuccess(t, node.Initialize(ctx))
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

func TestWorkspaceFileAndArtifactProtection(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := t.TempDir()
	gitFixtureCommand(t, repo, "init")
	f := filepath.Join(repo, "file")
	requireSuccess(t, os.WriteFile(f, []byte("hello"), 0644))
	requireSuccess(t, os.Symlink(".git", filepath.Join(repo, "alias")))
	requireSuccess(t, os.Mkdir(filepath.Join(repo, ".apparat"), 0700))
	config := `{"version":1,"tasks":[{"id":"one","enabled":true,"action":"snapshot-text","path":"file"},{"id":"missing","enabled":true,"action":"snapshot-text","path":"absent"}]}`
	requireSuccess(t, os.WriteFile(filepath.Join(repo, ".apparat/tasks.json"), []byte(config), 0600))
	node := nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	defer func() {
		_ = node.Close()
	}()
	requireSuccess(t, node.StoreProject(ctx, "", repo, ""))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	q := domain.WorkspaceRequest{Approved: true, ProjectID: rows[0].ID, Path: "file", Action: "read"}
	call := func(action string) domain.WorkspaceView {
		t.Helper()
		q.Action = action
		v, err := node.Workspace(ctx, q)
		requireSuccess(t, err)
		return v
	}
	fail := func(action string) {
		t.Helper()
		q.Action = action
		_, err := node.Workspace(ctx, q)
		if err == nil {
			t.Fatal("unexpected success", action, q.Path)
		}
	}
	q.Base = call("read").Base
	q.Argument = "renamed"
	call("move")
	q.Path = "renamed"
	q.Argument = ".apparat/tasks.json"
	fail("move")
	q.Argument = "file"
	call("move")
	q.Path = "alias/config"
	fail("read")
	q.Path = "../file"
	fail("read")
	q.Path = "binary"
	requireSuccess(t, os.WriteFile(filepath.Join(repo, q.Path), []byte{0}, 0600))
	fail("read")
	q.Path = "file"
	newID := func() string {
		id, err := domain.NewID(time.Now(), nil)
		requireSuccess(t, err)
		return id
	}
	q.ID = newID()
	q.Argument = "missing"
	fail("run")
	history := call("runs")
	if len(history.Choices) != 1 || history.Choices[0].State != "failed" {
		t.Fatal("failure absent from run history")
	}
	q.ID = newID()
	q.Argument = "one"
	call("run")
	one := q.ID
	q.ID = newID()
	call("run")
	two := q.ID
	requireSuccess(t, node.Close())
	node = nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	q.ID = one
	call("delete-artifact")
	fail("artifact")
	q.ID = two
	if call("artifact").Report != "Verified file\nhello" {
		t.Fatal("shared artifact lost after deletion")
	}
	q.ID = newID()
	q.Argument = "test message"
	q.Thread = "Draft discussion"
	call("post")
	requireSuccess(t, node.Close())
	node = nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	if len(call("threads").Choices) != 1 || !strings.Contains(call("chat").Report, "test message") {
		t.Fatal("chat did not survive restart")
	}
}

func TestWorkspaceGitActions(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	gitFixtureCommand(t, repo, "init")
	path := filepath.Join(repo, "note")
	requireSuccess(t, os.WriteFile(path, []byte("base\n"), 0600))
	gitFixtureCommand(t, repo, "add", "note")
	gitFixtureCommand(t, repo, "commit", "-m", "base")
	gitFixtureCommand(t, repo, "checkout", "--detach")
	node := nodeFor(t, t.TempDir(), false)
	requireSuccess(t, node.Initialize(ctx))
	defer func() {
		_ = node.Close()
	}()
	requireSuccess(t, node.StoreProject(ctx, "", repo, ""))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	q := domain.WorkspaceRequest{ProjectID: rows[0].ID, Path: "note"}
	act := func(op string) domain.WorkspaceView {
		t.Helper()
		q.Action = op
		v, err := node.Workspace(ctx, q)
		requireSuccess(t, err)
		return v
	}
	q.Argument = strings.TrimSpace(act("branches").Report)
	act("switch")
	q.Argument = gitFixtureCommand(t, repo, "rev-parse", "HEAD")
	if !strings.Contains(act("commit-detail").Report, "base") {
		t.Fatal("commit details missing")
	}
	blob := gitFixtureCommand(t, repo, "rev-parse", "HEAD:note")
	command := exec.Command("git", "-C", repo, "update-index", "--index-info")
	command.Stdin = strings.NewReader("0 0000000000000000000000000000000000000000\tnote\n100644 " + blob + " 1\tnote\n100644 " + blob + " 2\tnote\n100644 " + blob + " 3\tnote\n")
	requireSuccess(t, command.Run())
	if !strings.Contains(act("conflict").Report, "Stage 3") {
		t.Fatal("merge stages unavailable")
	}
	requireSuccess(t, os.WriteFile(path, []byte("resolved\n"), 0600))
	q.Base = act("read").Base
	act("resolve")
	if gitFixtureCommand(t, repo, "ls-files", "--unmerged") != "" {
		t.Fatal("unresolved index")
	}
	act("unstage")
	if gitFixtureCommand(t, repo, "show", ":note") != "base" {
		t.Fatal("unstage altered base")
	}
	act("delete")
	q.Base = "deleted"
	act("stage-deletion")
	q.Argument = "remove note"
	act("commit")
	if gitFixtureCommand(t, repo, "ls-tree", "HEAD", "note") != "" {
		t.Fatal("deletion lost")
	}
}

func TestDirectoryStatusScope(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	gitFixtureCommand(t, dir, "init")
	requireSuccess(t, os.MkdirAll(filepath.Join(dir, "nested", "gone"), 0700))
	write := func(name, text string) {
		t.Helper()
		requireSuccess(t, os.WriteFile(filepath.Join(dir, name), []byte(text), 0600))
	}
	write("tracked", "abc")
	write(".gitignore", "*.tmp\n")
	write("nested/gone/file", "old")
	gitFixtureCommand(t, dir, "add", ".")
	gitFixtureCommand(t, dir, "commit", "-m", "seed")
	write("tracked", "xyz")
	write("new", "new")
	write("ignored.tmp", "ignored")
	write("nested/another", "nested")
	requireSuccess(t, os.RemoveAll(filepath.Join(dir, "nested/gone")))
	node := nodeFor(t, t.TempDir(), false)
	requireSuccess(t, node.Initialize(ctx))
	defer func() {
		requireSuccess(t, node.Close())
	}()
	requireSuccess(t, node.StoreProject(ctx, "", dir, ""))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	check := func(path string, want map[string]string) {
		t.Helper()
		v, err := node.Workspace(ctx, domain.WorkspaceRequest{ProjectID: rows[0].ID, Action: "directory-status", Path: path})
		requireSuccess(t, err)
		if len(v.Changes) != len(want) {
			t.Fatalf("%s: unexpected changes %+v", path, v.Changes)
		}
		for _, c := range v.Changes {
			if want[c.Path] != c.Marker {
				t.Fatalf("unexpected change %+v", c)
			}
		}
	}
	check(".", map[string]string{"tracked": "C", "new": "U"})
	check("nested", map[string]string{"nested/another": "U", "nested/gone": "D"})
	check("nested/gone", map[string]string{"nested/gone/file": "D"})
}
