// Opens and initializes SQLite storage.
// Dependencies: modernc.org/sqlite.
package persistence

import (
	"context"

	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	_ "modernc.org/sqlite"
	"net/url"
)

type DB struct {
	sql         *sql.DB
	queueMu     sync.RWMutex
	queueHeads  map[queueView]int64
	policyMu    sync.RWMutex
	policyHeads map[string]PolicyCursor
}

type Migration struct {
	Version int
	Name    string
	SQL     string
}

type Diagnostic struct {
	UserVersion    int
	ForeignKeys    bool
	MigrationCount int
}

func Open(ctx context.Context, path string) (*DB, error) {
	db, err := sql.Open("sqlite", "file:"+url.PathEscape(path)+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(1500)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	{
		_, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
		if err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	{
		err := db.PingContext(ctx)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return &DB{sql: db}, nil
}

func (db *DB) Close() error {
	return db.sql.Close()
}

func (db *DB) ApplyMigrations(ctx context.Context, migrations []Migration) error {
	{
		_, err := db.sql.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL, checksum TEXT NOT NULL, applied_at_ms INTEGER NOT NULL)`)
		if err != nil {
			return err
		}
	}
	names := map[string]bool{}
	previous := 0
	for _, migration := range migrations {
		if migration.Version <= previous || migration.Name == "" || names[migration.Name] {
			return errors.New("migration versions must increase and names must be unique")
		}
		previous = migration.Version
		names[migration.Name] = true
	}
	for _, migration := range migrations {
		checksum := checksum(migration.SQL)
		var existing string
		var existingName string
		err := db.sql.QueryRowContext(ctx, `SELECT checksum, name FROM schema_migrations WHERE version = ?`, migration.Version).Scan(&existing, &existingName)
		if err == nil {
			if existing != checksum || existingName != migration.Name {
				return fmt.Errorf("migration %d checksum mismatch", migration.Version)
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		tx, err := db.sql.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		{
			_, err := tx.ExecContext(ctx, migration.SQL)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		{
			_, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, name, checksum, applied_at_ms) VALUES (?, ?, ?, ?)`, migration.Version, migration.Name, checksum, UTCMillis(time.Now()))
			if err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		{
			err := tx.Commit()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *DB) Diagnostics(ctx context.Context) (Diagnostic, error) {
	var diag Diagnostic
	{
		err := db.sql.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&diag.UserVersion)
		if err != nil {
			return diag, err
		}
	}
	var foreignKeys int
	{
		err := db.sql.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys)
		if err != nil {
			return diag, err
		}
	}
	diag.ForeignKeys = foreignKeys == 1
	_ = db.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&diag.MigrationCount)
	return diag, nil
}

func NewID() (string, error) {
	return domain.NewID(time.Now(), nil)
}

func UTCMillis(t time.Time) int64 {
	return t.UTC().UnixNano() / int64(time.Millisecond)
}

func checksum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
