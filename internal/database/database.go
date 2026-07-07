package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ SQL *sql.DB }

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
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &DB{SQL: db}, nil
}

func (db *DB) Close() error { return db.SQL.Close() }

func (db *DB) ApplyMigrations(ctx context.Context, migrations []Migration) error {
	if _, err := db.SQL.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL, checksum TEXT NOT NULL, applied_at_ms INTEGER NOT NULL)`); err != nil {
		return err
	}
	for _, migration := range migrations {
		checksum := checksum(migration.SQL)
		var existing string
		err := db.SQL.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = ?`, migration.Version).Scan(&existing)
		if err == nil {
			if existing != checksum {
				return fmt.Errorf("migration %d checksum mismatch", migration.Version)
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		tx, err := db.SQL.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, name, checksum, applied_at_ms) VALUES (?, ?, ?, ?)`, migration.Version, migration.Name, checksum, UTCMillis(time.Now())); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) Diagnostics(ctx context.Context) (Diagnostic, error) {
	var diag Diagnostic
	if err := db.SQL.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&diag.UserVersion); err != nil {
		return diag, err
	}
	var foreignKeys int
	if err := db.SQL.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		return diag, err
	}
	diag.ForeignKeys = foreignKeys == 1
	_ = db.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&diag.MigrationCount)
	return diag, nil
}

func NewID() (string, error) {
	var random [10]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%013x%s", UTCMillis(time.Now()), hex.EncodeToString(random[:])), nil
}

func UTCMillis(t time.Time) int64 { return t.UTC().UnixNano() / int64(time.Millisecond) }

func checksum(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
