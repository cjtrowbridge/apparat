// Exercises real notifications over the pinned TLS worker with two subscriptions.
// Blocking an active member closes the socket before another directory event can be disclosed.
package httpapi

import (
	"context"
	"database/sql"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func exerciseWorkspaceWatch(t *testing.T, c *WorkerClient, db *persistence.DB, dbPath string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	dir := t.TempDir()
	check := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	check(os.Mkdir(filepath.Join(dir, "sub"), 0700))
	const p = "live-test"
	check(db.RegisterProject(ctx, domain.Project{ID: p, Name: "fixture", Location: dir}))
	// An approved administrator watches without an explicit checkout grant.
	events := make(chan Notice, 8)
	done := make(chan error, 1)
	go func() {
		done <- c.Watch(ctx, []Watch{{Project: p, Directory: "."}, {Project: p, Directory: "sub"}}, func(n Notice) {
			events <- n
		})
	}()
	next := func() Notice {
		select {
		case n := <-events:
			return n
		case err := <-done:
			t.Fatal(err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		return Notice{}
	}
	if !next().Reset || !next().Reset {
		t.Fatal("new watch missing snapshot signal")
	}
	check(os.WriteFile(filepath.Join(dir, "sub", "new"), []byte("value"), 0600))
	n := next()
	if n.Reset || n.Resource.Directory != "sub" {
		t.Fatal(n)
	}
	fixture, err := sql.Open("sqlite", dbPath)
	check(err)
	defer fixture.Close()
	_, err = fixture.Exec("UPDATE cluster_users SET admin=0 WHERE user_id=?", c.Access.User)
	check(err)
	defer func() {
		_, restoreErr := fixture.Exec("UPDATE cluster_users SET admin=1,admission='approved' WHERE user_id=?", c.Access.User)
		check(restoreErr)
	}()

	page, err := c.CatalogPage(ctx, "")
	check(err)
	visible := false
	for _, item := range page {
		if item.ID == p {
			visible = true
		}
	}
	if !visible {
		t.Fatal("approved member missing registered checkout")
	}
	check(os.WriteFile(filepath.Join(dir, "sub", "new"), []byte("member can watch"), 0600))
	_ = next()
	_, err = fixture.Exec("UPDATE cluster_users SET admission='blocked' WHERE user_id=?", c.Access.User)
	check(err)
	_, err = c.CatalogPage(ctx, "")
	if err == nil {
		t.Fatal("blocked member retained catalog access")
	}

	check(os.WriteFile(filepath.Join(dir, "sub", "new"), []byte("changed"), 0600))
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("watch did not close after revoke")
	}
}
