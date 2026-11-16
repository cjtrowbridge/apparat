// Verifies opens and initializes SQLite storage.
// Uses package-local contracts and Go standard-library types.
package persistence

import (
	"context"
	"strings"
	"testing"
)

func TestOpenMigrateDiagnostics(t *testing.T) {
	db, err := Open(context.Background(), t.TempDir()+"/apparat.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	migration := Migration{Version: 1, Name: "devices", SQL: `CREATE TABLE devices(id TEXT PRIMARY KEY);`}
	{
		err := db.ApplyMigrations(context.Background(), []Migration{migration})
		if err != nil {
			t.Fatal(err)
		}
	}
	diag, err := db.Diagnostics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !diag.ForeignKeys || diag.MigrationCount != 1 {
		t.Fatalf("diag = %+v", diag)
	}
}

func TestMigrationChecksumMismatch(t *testing.T) {
	db, err := Open(context.Background(), t.TempDir()+"/apparat.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	{
		err := db.ApplyMigrations(context.Background(), []Migration{{Version: 1, Name: "one", SQL: `CREATE TABLE one(id TEXT);`}})
		if err != nil {
			t.Fatal(err)
		}
	}
	err = db.ApplyMigrations(context.Background(), []Migration{{Version: 1, Name: "one", SQL: `CREATE TABLE two(id TEXT);`}})
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("expected checksum error, got %v", err)
	}
}
