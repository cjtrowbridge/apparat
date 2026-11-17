// Verifies transactional services, stale-probe fencing and backup recovery.
// All SQLite integration state is confined to disposable test roots.
package persistence

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"path/filepath"
	"testing"
)

func serviceDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	{
		err = db.Migrate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	return db
}
func TestServiceTransactionsFenceStaleResults(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	id := "00000000010000000000000000"
	service := domain.Service{ID: id, Name: "first", Driver: "mock-text", Enabled: true, Revision: 1}
	{
		err := db.SaveService(ctx, service, 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	observation := domain.Observation{ServiceID: id, Revision: 1, At: 2, Healthy: true, Capabilities: []domain.Capability{{ID: id, ServiceID: id, Workload: "text_generation"}}}
	{
		err := db.Observe(ctx, observation)
		if err != nil {
			t.Fatal(err)
		}
	}
	service.Enabled = false
	service.Revision = 2
	{
		err := db.SaveService(ctx, service, 1)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err := db.Observe(ctx, observation)
		if !errors.Is(err, domain.ErrConflict) {
			t.Fatal("stale accepted", err)
		}
	}
	list, err := db.Services(ctx)
	if err != nil || len(list) != 1 || list[0].Health != "disabled" || len(list[0].Capabilities) != 0 {
		t.Fatal(list, err)
	}
	{
		err = db.RemoveService(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err = db.Observe(ctx, observation)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatal(err)
		}
	}
}
func TestBackupRestoreAndCorruption(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	source := filepath.Join(t.TempDir(), "backup.db")
	{
		err := db.Backup(ctx, source)
		if err != nil {
			t.Fatal(err)
		}
	}
	restored := filepath.Join(t.TempDir(), "restore.db")
	{
		err := RestoreBackup(ctx, source, restored)
		if err != nil {
			t.Fatal(err)
		}
	}
	if RestoreBackup(ctx, source, restored) == nil {
		t.Fatal("overwrote existing destination")
	}
	{
		err := os.WriteFile(source, []byte("corrupt"), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	if RestoreBackup(ctx, source, filepath.Join(t.TempDir(), "restore.db")) == nil {
		t.Fatal("corruption accepted")
	}
}
func TestMigrationRollbackAndReplayErrors(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	err := db.ApplyMigrations(ctx, []Migration{{Version: 5, Name: "broken", SQL: "CREATE TABLE rollback_probe(id TEXT); INVALID SQL;"}})
	if err == nil {
		t.Fatal("invalid migration accepted")
	}
	var count int
	err = db.sql.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE name='rollback_probe'").Scan(&count)
	if err != nil || count != 0 {
		t.Fatal("migration not rolled back", count, err)
	}
	store := NewMessages(db.sql, 5)
	seen, err := store.Seen(ctx, "message")
	if seen || err != nil {
		t.Fatal(seen, err)
	}
	seen, err = store.Seen(ctx, "message")
	if !seen || err != nil {
		t.Fatal(seen, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	seen, err = store.Seen(cancelled, "another")
	if seen || !errors.Is(err, context.Canceled) {
		t.Fatal("storage error mistaken for replay", seen, err)
	}
}

func TestBackupPreservesExistingMetadata(t *testing.T) {
	db := serviceDB(t)
	path := filepath.Join(t.TempDir(), "backup.db")
	{
		err := os.WriteFile(path+".json", []byte("existing"), 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	if db.Backup(context.Background(), path) == nil {
		t.Fatal("overwrote metadata")
	}
	data, err := os.ReadFile(path + ".json")
	if err != nil || string(data) != "existing" {
		t.Fatal(string(data), err)
	}
	{
		_, err = os.Stat(path)
		if !os.IsNotExist(err) {
			t.Fatal("partial snapshot remains", err)
		}
	}
}

func TestForeignKeysAndBusyClassification(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	_, err := db.sql.ExecContext(ctx, `INSERT INTO service_capabilities VALUES('missing','id','text_generation')`)
	if err == nil || Retryable(err) {
		t.Fatal("foreign key failure misclassified", err)
	}
	if Retryable(context.Canceled) {
		t.Fatal("cancellation retryable")
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = db.sql.ExecContext(cancelled, "SELECT 1")
	_ = tx.Rollback()
	if !errors.Is(err, context.Canceled) {
		t.Fatal("connection wait not cancellable", err)
	}
}

func TestDemoSeedIsAtomic(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	service := domain.Service{ID: "00000000000000000000000000", Name: "one", Driver: "mock-text", Enabled: true, Revision: 1}
	if db.SeedServices(ctx, []domain.Service{service, service}) == nil {
		t.Fatal("duplicate seed accepted")
	}
	list, err := db.Services(ctx)
	if err != nil || len(list) != 0 {
		t.Fatal("partial seed survived", list, err)
	}
}

func BenchmarkThreeServiceQuery(b *testing.B) {
	db, err := Open(context.Background(), filepath.Join(b.TempDir(), "bench.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()
	{
		err = db.Migrate(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
	for index := 0; index < 3; index++ {
		id, _ := NewID()
		service := domain.Service{ID: id, Name: "benchmark", Driver: "mock-text", Enabled: true, Revision: 1}
		{
			err = db.SaveService(context.Background(), service, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		{
			_, err = db.Services(context.Background())
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
