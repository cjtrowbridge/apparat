// Exercises real notifications over the pinned TLS worker with two subscriptions.
// A revoked project grant closes the socket before another directory event can be disclosed.
package httpapi

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func exerciseWorkspaceWatch(t *testing.T, c *WorkerClient, db *persistence.DB) {
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
	check(db.WorkspaceGrant(ctx, p, c.Access.User, "read", true))
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
	check(db.WorkspaceGrant(ctx, p, c.Access.User, "read", false))
	check(os.WriteFile(filepath.Join(dir, "sub", "new"), []byte("changed"), 0600))
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("watch did not close after revoke")
	}
}
