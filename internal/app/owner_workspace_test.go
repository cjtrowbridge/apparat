// Verifies scoped remote operations on an actual SQLite registration and working file.
// Includes conflict handling, lost save acknowledgments, read-cache validation and grant revocation.
package app

import (
	"context"
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
	db, err := persistence.Open(ctx, filepath.Join(t.TempDir(), "node.db"))
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
	_, err = n.serveWorkspace(ctx, caller, q)
	var missing httpapi.WorkspaceGrantNeeded
	if !errors.Is(err, httpapi.ErrDenied) || !errors.As(err, &missing) || missing.Scope != "read" {
		t.Fatal("admin obtained file access without grant")
	}
	check(db.WorkspaceGrant(ctx, project, caller.User, "read", true))
	v, err := n.serveWorkspace(ctx, caller, q)
	check(err)
	q.Known = v.Base
	v, err = n.serveWorkspace(ctx, caller, q)
	check(err)
	if !v.Unchanged || v.Text != "" {
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
	check(db.WorkspaceGrant(ctx, project, caller.User, "write", true))
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
	check(db.WorkspaceGrant(ctx, project, caller.User, "read", false))
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
		page, err := db.RegistryPage(ctx, before, true, "owner", true)
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
	hidden, err := db.RegistryPage(ctx, "", true, "other", false)
	if err != nil || len(hidden) != 0 || len(seen) != 70 {
		t.Fatal("incorrect catalog isolation or continuation")
	}
}
