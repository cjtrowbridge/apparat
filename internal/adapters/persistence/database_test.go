// Verifies opens and initializes SQLite storage.
// Uses package-local contracts and Go standard-library types.
package persistence

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

func TestModelScopeAndSessionIsolation(t *testing.T) {
	ctx := t.Context()
	db, err := Open(ctx, t.TempDir()+"/model.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a := domain.ModelRef{Provider: "ollama", Name: "a", Digest: "a"}
	b := domain.ModelRef{Provider: "ollama", Name: "b", Digest: "b"}
	first, err := db.BindChatModel(ctx, "", "project-a", "old", a)
	if err != nil || first.Session != a || first.Default.Model.Valid() {
		t.Fatal(first, err)
	}
	id, err := domain.NewID(time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	edit := domain.ModelEdit{Project: "project-a", Session: "current", Default: true, Operation: id, Model: b}
	saved, err := db.SaveChatModel(ctx, "", edit)
	if err != nil || saved.Default.Model != b || saved.Session != b {
		t.Fatal(saved, err)
	}
	retry, err := db.SaveChatModel(ctx, "", edit)
	if err != nil || retry != saved {
		t.Fatal("lost acknowledgement", retry, err)
	}
	for _, tc := range []struct {
		scope, project, session string
		want                    domain.ModelRef
	}{{"", "project-a", "new", b}, {"", "project-a", "old", a}, {"", "project-b", "new", a}, {"cluster", "", "new", a}} {
		got, err := db.BindChatModel(ctx, tc.scope, tc.project, tc.session, a)
		if err != nil || got.Session != tc.want {
			t.Fatal(tc, got, err)
		}
	}
	edit.Operation, err = domain.NewID(time.Now(), nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.SaveChatModel(ctx, "", edit)
	if !errors.Is(err, domain.ErrModelConflict) {
		t.Fatal("stale default overwrite", err)
	}
}

func TestCPUBaselineSurvivesConcurrentStartupAndRestart(t *testing.T) {
	ctx := t.Context()
	path := t.TempDir() + "/node.db"
	open := func() *DB {
		db, err := Open(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		return db
	}
	a, b := open(), open()
	err := a.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	const want = 12345678.987654321
	measure := func() float64 {
		calls.Add(1)
		return want
	}
	read := func(db *DB) {
		rate, err := db.CPUBaseline(ctx, measure)
		if err != nil || rate != want {
			t.Errorf("baseline %g: %v", rate, err)
		}
	}
	var wg sync.WaitGroup
	for _, db := range []*DB{a, b} {
		wg.Go(func() {
			read(db)
		})
	}
	wg.Wait()
	a.Close()
	b.Close()
	c := open()
	defer c.Close()
	read(c)
	if calls.Load() != 1 {
		t.Fatal("benchmark runs", calls.Load())
	}
}
