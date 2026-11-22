// Verifies persists cluster device profiles, change records and cursors in SQLite.
// Dependencies: github.com/cjtrowbridge/apparat/internal/database.
package persistence

import (
	"context"
	"database/sql"
	"testing"
)

func TestDirectoryPersistsProfilesAndChanges(t *testing.T) {
	db, err := Open(context.Background(), t.TempDir()+"/apparat.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	dir := NewDirectory(db.sql)
	{
		err := dir.Init(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	profile := DeviceProfile{ID: "device-1", Signature: "sig", Roles: []string{"service-host"}, Capabilities: []Capability{{WorkloadClass: "text_generation", Runtime: "mock", Healthy: true}}, Reachable: true}
	{
		err := dir.PutDevice(context.Background(), profile)
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := dir.GetDevice(context.Background(), "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Capabilities[0].WorkloadClass != "text_generation" {
		t.Fatalf("profile = %+v", got)
	}
	changes, err := dir.ChangesAfter(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].DeviceID != "device-1" {
		t.Fatalf("changes = %+v", changes)
	}
}

func TestProfileAndChangeRollbackTogether(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, t.TempDir()+"/node.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	directory := NewDirectory(db.sql)
	err = directory.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.sql.ExecContext(ctx, "CREATE TRIGGER reject_change BEFORE INSERT ON cluster_changes BEGIN SELECT RAISE(ABORT,'full'); END")
	if err != nil {
		t.Fatal(err)
	}
	err = directory.PutDevice(ctx, DeviceProfile{ID: "one"})
	if err == nil {
		t.Fatal("expected failure recording change")
	}
	_, err = directory.GetDevice(ctx, "one")
	if err != sql.ErrNoRows {
		t.Fatalf("unexpected device after rollback: %v", err)
	}
}
