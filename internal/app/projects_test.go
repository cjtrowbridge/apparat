// Exercises registration across restart, location changes and removal.
// An opt-in integration run uses real GitHub transfers with host-owned credentials.
package app_test

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	git "github.com/go-git/go-git/v5"
	"os"
	"path/filepath"
	"testing"
)

func requireSuccess(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func TestProjectRegistryLifecycle(t *testing.T) {
	ctx := context.Background()
	nodeRoot := t.TempDir()
	checkout := filepath.Join(t.TempDir(), "work")
	_, err := git.PlainInit(checkout, false)
	requireSuccess(t, err)
	requireSuccess(t, os.Mkdir(filepath.Join(checkout, ".vscode"), 0700))
	tasks := filepath.Join(checkout, ".vscode", "tasks.json")
	requireSuccess(t, os.WriteFile(tasks, []byte(`{"tasks":[{"label":"Rebuild stale CAD designs"}]}`), 0600))
	node := nodeFor(t, nodeRoot, false)
	requireSuccess(t, node.Initialize(ctx))
	requireSuccess(t, node.StoreProject(ctx, "", checkout, "work"))
	requireSuccess(t, node.StoreProject(ctx, "", checkout, ""))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	if len(rows) != 1 || !domain.ValidID(rows[0].ID) || len(rows[0].Actions) != 1 || rows[0].Owner == "" {
		t.Fatal("incomplete registration")
	}
	id := rows[0].ID
	requireSuccess(t, node.Close())
	node = nodeFor(t, nodeRoot, false)
	requireSuccess(t, node.Initialize(ctx))
	rows, err = node.Projects(ctx)
	requireSuccess(t, err)
	if len(rows) != 1 || rows[0].ID != id || rows[0].Account != "work" {
		t.Fatal("lost registration after restart")
	}
	requireSuccess(t, os.WriteFile(tasks, []byte("{}"), 0600))
	rows, err = node.Projects(ctx)
	requireSuccess(t, err)
	if len(rows[0].Actions) != 0 {
		t.Fatal("stale action list")
	}
	moved := checkout + "-moved"
	requireSuccess(t, os.Rename(checkout, moved))
	rows, err = node.Projects(ctx)
	requireSuccess(t, err)
	if rows[0].Available {
		t.Fatal("missing checkout reported available")
	}
	requireSuccess(t, node.StoreProject(ctx, id, moved, "personal"))
	rows, err = node.Projects(ctx)
	requireSuccess(t, err)
	if rows[0].ID != id || !rows[0].Available || rows[0].Location != moved {
		t.Fatal("reconnection lost identity")
	}
	requireSuccess(t, node.RemoveProject(ctx, id))
	_, err = os.Stat(filepath.Join(moved, ".vscode", "tasks.json"))
	requireSuccess(t, err)
}
func TestPhase9AuthorizedGitHubClones(t *testing.T) {
	if os.Getenv("APPARAT_TEST_PHASE9_GITHUB") != "1" {
		t.Skip("requires authorized host GitHub fixtures")
	}
	ctx := context.Background()
	nodeRoot := t.TempDir()
	node := nodeFor(t, nodeRoot, false)
	requireSuccess(t, node.Initialize(ctx))
	privateParent := t.TempDir()
	results := make(chan error, 2)
	go func() {
		results <- node.CloneProject(ctx, "", "cjtrowbridge/vibe-modeling", "")
	}()
	go func() {
		results <- node.CloneProject(ctx, privateParent, "cjtrowbridge/apparat", "cjtrowbridge")
	}()
	first, second := <-results, <-results
	requireSuccess(t, first)
	requireSuccess(t, second)
	requireSuccess(t, node.Close())
	node = nodeFor(t, nodeRoot, false)
	requireSuccess(t, node.Initialize(ctx))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	if len(rows) != 2 {
		t.Fatal("expected two real checkouts")
	}
	for _, project := range rows {
		if !project.Available || project.Branch != "main" || len(project.Actions) == 0 {
			t.Fatal("missing Git or task metadata", project.Name)
		}
	}
	_, err = projectgit.Inspect(filepath.Join(privateParent, "apparat"))
	requireSuccess(t, err)
}

func TestCreateLocalProject(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	node := nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	for _, parent := range []string{"", t.TempDir()} {
		path, err := node.CreateProject(ctx, parent, "My project")
		requireSuccess(t, err)
		expected := parent
		if expected == "" {
			expected = filepath.Join(root, "data", "repos")
		}
		if path != filepath.Join(expected, "My project") {
			t.Fatal("wrong repository location", path)
		}
		_, err = node.CreateProject(ctx, parent, "My project")
		if err == nil {
			t.Fatal("duplicate creation succeeded")
		}
		for _, bad := range []string{".git", "../outside", "nested/name", ""} {
			_, err = node.CreateProject(ctx, parent, bad)
			if err == nil {
				t.Fatal("accepted invalid name", bad)
			}
		}
		rows, err := node.Projects(ctx)
		requireSuccess(t, err)
		var id string
		for _, p := range rows {
			if p.Location == path && p.Branch == "main" && p.Revision == "" && p.Available {
				id = p.ID
			}
		}
		q := domain.WorkspaceRequest{ProjectID: id, Path: ".", Action: "list"}
		_, err = node.Workspace(ctx, q)
		requireSuccess(t, err)
		q.Action = "history"
		_, err = node.Workspace(ctx, q)
		requireSuccess(t, err)
		requireSuccess(t, os.WriteFile(filepath.Join(path, "notes.txt"), []byte("hello"), 0600))
		q.Path = "notes.txt"
		q.Action = "stage"
		q.Base = projectgit.Digest("hello")
		_, err = node.Workspace(ctx, q)
		requireSuccess(t, err)
		q.Action = "commit"
		q.Argument = "First commit"
		_, err = node.Workspace(ctx, q)
		requireSuccess(t, err)
		info, err := projectgit.Inspect(path)
		requireSuccess(t, err)
		if info.Revision == "" || info.Branch != "main" {
			t.Fatal("missing first commit")
		}
	}
	requireSuccess(t, node.Close())
	node = nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	reopened, err := node.Projects(ctx)
	requireSuccess(t, err)
	if len(reopened) != 2 || !reopened[0].Available || !reopened[1].Available {
		t.Fatal("new projects did not reopen")
	}
	requireSuccess(t, node.Close())
}
