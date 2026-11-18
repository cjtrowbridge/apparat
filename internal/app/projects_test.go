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
