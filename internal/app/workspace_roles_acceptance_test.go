// Exercises the project permission matrix against two real pinned worker endpoints.
// Registers isolated repositories and checks API replies against each worker's files.
package app

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func requireWorkspaceOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type workspaceActor struct {
	person *identity.Personal
	device *identity.Session
}

func fixtureDevice(t *testing.T) *identity.Session {
	d := identity.NewSession(t.TempDir())
	requireWorkspaceOK(t, d.Install(nil, "", "plain", "", "", true))
	return d
}
func fixtureActor(t *testing.T, name string) workspaceActor {
	p := identity.NewPersonal(t.TempDir())
	requireWorkspaceOK(t, p.Create(name, []byte("acceptance fixture password")))
	return workspaceActor{p, fixtureDevice(t)}
}
func TestClusterWorkspaceRoleMatrix(t *testing.T) {
	owner := fixtureActor(t, "Owner")
	member := fixtureActor(t, "Member")
	for _, name := range []string{"first-worker", "second-worker"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			exerciseRoleWorker(t, owner, member)
		})
	}
}
func exerciseRoleWorker(t *testing.T, owner, member workspaceActor) {
	ctx := t.Context()
	root := t.TempDir()
	path := filepath.Join(root, "worker.db")
	db, err := persistence.Open(ctx, path)
	requireWorkspaceOK(t, err)
	defer db.Close()
	requireWorkspaceOK(t, db.Migrate(ctx))
	key := fixtureDevice(t)
	pin := identity.Fingerprint(key.Public())
	requireWorkspaceOK(t, db.ConfirmOwner(ctx, pin, owner.person.Fingerprint()))
	node := Runtime{database: db}
	defer node.pages.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	requireWorkspaceOK(t, err)
	server := httpapi.ServeWorkerListener(listener, root, db, key, httpapi.WorkerSurface{Workspace: node.serveWorkspace})
	defer server.Close()
	login := func(a workspaceActor) *httpapi.WorkerClient {
		c, err := httpapi.OpenWorker(ctx, listener.Addr().String(), pin, a.person, a.device)
		requireWorkspaceOK(t, err)
		return c
	}
	admin := login(owner)
	pending := login(member)
	_, err = pending.CatalogPage(ctx, "")
	if !errors.Is(err, httpapi.ErrDenied) {
		t.Fatal("pending member could access catalog", err)
	}
	head := persistence.PolicyCursor{}
	change := func(edit identity.PolicyEdit) {
		edit.Cluster = admin.Access.Cluster
		edit.Base = head.Revision
		edit.ID = fmt.Sprintf("acceptance-%d", edit.Base)
		receipt, err := admin.ChangePolicy(ctx, owner.person, owner.device, edit, head.Hash)
		requireWorkspaceOK(t, err)
		head = persistence.PolicyCursor{Revision: edit.Base + 1, Hash: receipt.Hash()}
	}
	change(identity.PolicyEdit{Kind: "order_on_worker", Target: pin, Value: admin.Access.User, Service: admin.Access.Device})
	role := func(value string) {
		change(identity.PolicyEdit{Kind: "user_role", Target: pending.Access.User, Value: value})
	}
	role("approved")
	reader := login(member)
	paths := registerMatrixProjects(t, db, root)
	checkMatrixCatalog(t, reader, paths, false)
	checkMatrixCatalog(t, admin, paths, true)
	id := "remote-fixture"
	file := filepath.Join(paths[id], "nested", "note")
	fixture, err := sql.Open("sqlite", path)
	requireWorkspaceOK(t, err)
	defer fixture.Close()
	_, err = fixture.Exec("INSERT INTO workspace_grants VALUES(?,?,?)", id, reader.Access.User, "write")
	requireWorkspaceOK(t, err)
	read := domain.WorkspaceRequest{ProjectID: id, Path: "nested/note", Action: "read"}
	checkMatrixSave(t, reader, read, file, false)
	checkMatrixSave(t, admin, read, file, true)
	role("admin")
	checkMatrixSave(t, reader, read, file, true)
	view, err := reader.Workspace(ctx, read)
	requireWorkspaceOK(t, err)
	role("approved")
	checkMatrixWrite(t, reader, read, file, view, false)
	read.Known = view.Base
	cached, err := reader.Workspace(ctx, read)
	requireWorkspaceOK(t, err)
	if !cached.Unchanged || !cached.ReadOnly {
		t.Fatal("demoted session regained editing on cached read")
	}
	watchCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	notices := make(chan httpapi.Notice, 4)
	done := make(chan error, 1)
	go func() {
		done <- reader.Watch(watchCtx, []httpapi.Watch{{Project: id, Directory: "nested"}}, func(n httpapi.Notice) {
			notices <- n
		})
	}()
	select {
	case n := <-notices:
		if !n.Reset {
			t.Fatal("watch never received initial snapshot")
		}
	case err := <-done:
		t.Fatal("member could not watch", err)
	case <-watchCtx.Done():
		t.Fatal("watch setup timed out")
	}
	role("blocked")
	_, err = reader.Workspace(ctx, read)
	if !errors.Is(err, httpapi.ErrDenied) {
		t.Fatal("blocked session could refresh cached file", err)
	}
	_, err = reader.CatalogPage(ctx, "")
	if !errors.Is(err, httpapi.ErrDenied) {
		t.Fatal("blocked session could list projects", err)
	}
	requireWorkspaceOK(t, os.WriteFile(file, []byte("after block"), 0600))
	select {
	case <-done:
		if watchCtx.Err() != nil || len(notices) != 0 {
			t.Fatal("blocked watch timed out or disclosed an update")
		}
	case <-notices:
		t.Fatal("blocked session received an update")
	case <-watchCtx.Done():
		t.Fatal("blocked watch stayed open")
	}
}
func registerMatrixProjects(t *testing.T, db *persistence.DB, root string) map[string]string {
	paths := map[string]string{}
	origin := filepath.Join(root, "origin.git")
	requireWorkspaceOK(t, exec.Command("git", "init", "--bare", "-b", "main", origin).Run())
	for i := 0; i < 33; i++ {
		id := fmt.Sprintf("local-%02d", i)
		if i == 0 {
			id = "remote-fixture"
		}
		path := filepath.Join(root, id)
		if i == 0 {
			requireWorkspaceOK(t, exec.Command("git", "clone", "--local", origin, path).Run())
		} else {
			requireWorkspaceOK(t, exec.Command("git", "init", "-b", "main", path).Run())
		}
		requireWorkspaceOK(t, os.Mkdir(filepath.Join(path, "nested"), 0700))
		requireWorkspaceOK(t, os.WriteFile(filepath.Join(path, "nested", "note"), []byte(id), 0600))
		requireWorkspaceOK(t, db.RegisterProject(t.Context(), domain.Project{ID: id, Location: path, Name: id}))
		paths[id] = path
	}
	return paths
}
func checkMatrixCatalog(t *testing.T, c *httpapi.WorkerClient, paths map[string]string, admin bool) {
	t.Helper()
	seen := map[string]bool{}
	cursor := ""
	for {
		page, err := c.CatalogPage(t.Context(), cursor)
		requireWorkspaceOK(t, err)
		if len(page) == 0 || len(page) > 32 {
			t.Fatalf("admin=%t catalog page size %d", admin, len(page))
		}
		for _, project := range page {
			id := project.ID
			if paths[id] == "" || seen[id] {
				t.Fatal("unexpected catalog entry", id)
			}
			seen[id] = true
			q := domain.WorkspaceRequest{ProjectID: id, Path: "nested", Action: "list"}
			listing, err := c.Workspace(t.Context(), q)
			requireWorkspaceOK(t, err)
			if !slices.Contains(listing.Entries, domain.WorkspaceEntry{Name: "note"}) {
				t.Fatal("missing nested file", id)
			}
			q.Path = "nested/note"
			q.Action = "read"
			view, err := c.Workspace(t.Context(), q)
			requireWorkspaceOK(t, err)
			if view.Text != id || view.ReadOnly == admin {
				t.Fatalf("admin=%t incorrect read for %s", admin, id)
			}
			checkMatrixWrite(t, c, q, filepath.Join(paths[id], "nested", "note"), view, admin)
		}
		cursor = page[len(page)-1].NextCursor
		if cursor == "" {
			break
		}
	}
	if len(seen) != len(paths) {
		t.Fatalf("admin=%t catalog contained %d of %d projects", admin, len(seen), len(paths))
	}
}
func checkMatrixSave(t *testing.T, c *httpapi.WorkerClient, q domain.WorkspaceRequest, path string, allowed bool) {
	t.Helper()
	view, err := c.Workspace(t.Context(), q)
	requireWorkspaceOK(t, err)
	checkMatrixWrite(t, c, q, path, view, allowed)
}
func checkMatrixWrite(t *testing.T, c *httpapi.WorkerClient, read domain.WorkspaceRequest, path string, view domain.WorkspaceView, allowed bool) {
	t.Helper()
	before, err := os.ReadFile(path)
	requireWorkspaceOK(t, err)
	if projectgit.Digest(string(before)) != view.Base {
		t.Fatal("fixture attempted save with stale revision")
	}
	q := read
	q.Action = "save"
	q.Text = string(before) + " edited"
	q.Base = view.Base
	q.ID, err = domain.NewID(time.Now(), rand.Reader)
	requireWorkspaceOK(t, err)
	_, err = c.Workspace(t.Context(), q)
	expected := string(before)
	if allowed {
		requireWorkspaceOK(t, err)
		expected = q.Text
	} else if !errors.Is(err, httpapi.ErrDenied) || !strings.Contains(err.Error(), "Only a current cluster administrator") {
		t.Fatalf("member save did not return workspace write denial: %v", err)
	}
	after, err := os.ReadFile(path)
	requireWorkspaceOK(t, err)
	current, err := c.Workspace(t.Context(), read)
	requireWorkspaceOK(t, err)
	if string(after) != expected || current.Text != expected {
		t.Fatalf("save allowed=%t: disk or API bytes incorrect", allowed)
	}
}
