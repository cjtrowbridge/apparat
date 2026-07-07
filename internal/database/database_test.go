package database

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
	defer func() { _ = db.Close() }()
	migration := Migration{Version: 1, Name: "devices", SQL: `CREATE TABLE devices(id TEXT PRIMARY KEY);`}
	if err := db.ApplyMigrations(context.Background(), []Migration{migration}); err != nil {
		t.Fatal(err)
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
	defer func() { _ = db.Close() }()
	if err := db.ApplyMigrations(context.Background(), []Migration{{Version: 1, Name: "one", SQL: `CREATE TABLE one(id TEXT);`}}); err != nil {
		t.Fatal(err)
	}
	err = db.ApplyMigrations(context.Background(), []Migration{{Version: 1, Name: "one", SQL: `CREATE TABLE two(id TEXT);`}})
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("expected checksum error, got %v", err)
	}
}
