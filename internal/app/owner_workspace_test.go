// Verifies scoped remote operations on an actual SQLite registration and working file.
// Includes conflict handling, lost save acknowledgments, read-cache validation and grant revocation.
package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOwnerWorkspaceLifecycle(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "node.db")
	db, err := persistence.Open(ctx, dbPath)
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	check(err)
	defer db.Close()
	check(db.Migrate(ctx))
	worker, err := identity.Generate()
	check(err)
	device := identity.NewSession(t.TempDir())
	check(device.Install(nil, "", "plain", "", "", true))
	person := identity.NewPersonal(t.TempDir())
	check(person.Create("Test", []byte("test secret")))
	pin := identity.Fingerprint(worker.Public)
	check(db.ConfirmOwner(ctx, pin, person.Fingerprint()))
	proof, err := person.SignLogin(device, pin, identity.ClaimDigest("fixture"))
	check(err)
	caller, err := db.Login(ctx, proof)
	check(err)
	dir := t.TempDir()
	const project = "remote-test"
	check(db.RegisterProject(ctx, domain.Project{ID: project, Location: dir, Name: "workspace"}))
	file := filepath.Join(dir, "file")
	check(os.WriteFile(file, []byte("first"), 0600))
	n := Runtime{database: db}
	defer n.pages.Close()
	q := domain.WorkspaceRequest{Action: "read", ProjectID: project, Path: "file"}
	view, err := n.serveWorkspace(ctx, caller, q)
	check(err)
	if view.Text != "first" {
		t.Fatal("admin could not read without a grant")
	}
	fixture, err := sql.Open("sqlite", dbPath)
	check(err)
	defer fixture.Close()
	_, err = fixture.Exec("UPDATE cluster_users SET admin=0 WHERE user_id=?", caller.User)
	check(err)
	// Legacy checkout write grants must never override the current cluster role.
	_, err = fixture.Exec("INSERT INTO workspace_grants VALUES(?,?,?)", project, caller.User, "write")
	check(err)
	caller, err = db.CurrentAccess(ctx, caller.Cluster, caller.User, caller.Device)
	check(err)
	v, err := n.serveWorkspace(ctx, caller, q)
	check(err)
	if !v.ReadOnly {
		t.Fatal("member received editable file")
	}
	q.Known = v.Base
	v, err = n.serveWorkspace(ctx, caller, q)
	check(err)
	if !v.ReadOnly || !v.Unchanged || v.Text != "" {
		t.Fatal("warm read sent redundant content")
	}
	q.Known = ""
	q.Action = "save"
	q.Base = projectgit.Digest("first")
	q.Text = "replacement"
	q.ID, err = domain.NewID(time.Now(), nil)
	check(err)
	_, err = n.serveWorkspace(ctx, caller, q)
	if !errors.Is(err, httpapi.ErrDenied) {
		t.Fatal("read grant allowed write")
	}
	scopes, err := db.WorkspaceScopes(ctx, project, caller.User)
	check(err)
	if len(scopes) != 1 || scopes[0] != "read" {
		t.Fatal(scopes)
	}
	check(db.WorkspaceGrant(ctx, project, caller.User, "write", false))
	_, err = fixture.Exec("UPDATE cluster_users SET admin=1 WHERE user_id=?", caller.User)
	check(err)
	check(os.WriteFile(file, []byte("external"), 0600))
	_, err = n.serveWorkspace(ctx, caller, q)
	if !errors.Is(err, httpapi.ErrWorkspaceConflict) {
		t.Fatal("expected conflict", err)
	}
	q.Base = projectgit.Digest("external")
	q.ID, err = domain.NewID(time.Now(), nil)
	check(err)
	_, err = n.serveWorkspace(ctx, caller, q)
	check(err)
	check(os.WriteFile(file, []byte("later"), 0600))
	_, err = n.serveWorkspace(ctx, caller, q)
	check(err)
	body, err := os.ReadFile(file)
	check(err)
	if string(body) != "later" {
		t.Fatal("retry overwrote a later edit")
	}
	_, err = fixture.Exec("UPDATE cluster_users SET admin=0,admission='blocked' WHERE user_id=?", caller.User)
	check(err)
	_, err = n.serveWorkspace(ctx, caller, q)
	if !errors.Is(err, httpapi.ErrDenied) {
		t.Fatal("revoked reader retained write")
	}
}

func TestRegistryPagesAndReaderVisibility(t *testing.T) {
	ctx := context.Background()
	db, err := persistence.Open(ctx, filepath.Join(t.TempDir(), "node.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 70; i++ {
		id := fmt.Sprintf("remote-%03d", i)
		err = db.RegisterProject(ctx, domain.Project{ID: id, Name: id, Location: id})
		if err != nil {
			t.Fatal(err)
		}
	}
	before := ""
	seen := map[string]bool{}
	for {
		page, err := db.RegistryPage(ctx, before)
		if err != nil || len(page) == 0 || len(page) > 32 {
			t.Fatal(len(page), err)
		}
		for _, p := range page {
			if seen[p.ID] {
				t.Fatal("duplicate cursor result")
			}
			seen[p.ID] = true
		}
		before = page[len(page)-1].NextCursor
		if before == "" {
			break
		}
	}
	if len(seen) != 70 {
		t.Fatal("incomplete catalog")
	}
}
