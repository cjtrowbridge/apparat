// Produces consistent SQLite snapshots and validates disposable restores.
// Backup metadata binds app/schema identity to the exact snapshot digest.
package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type BackupMetadata struct {
	App    string `json:"app"`
	Schema int    `json:"schema"`
	SHA256 string `json:"sha256"`
}

func (db *DB) Integrity(ctx context.Context) error {
	var result string
	err := db.sql.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return err
	}
	if result != "ok" {
		return errors.New("SQLite integrity check failed; preserve the database and restore a verified backup to a new root")
	}
	return nil
}
func (db *DB) Backup(ctx context.Context, destination string) error {
	err := db.Integrity(ctx)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_ = file.Close()
	metadataFile, err := os.OpenFile(destination+".json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		_ = os.Remove(destination)
		return err
	}
	defer func() {
		_ = metadataFile.Close()
	}()
	completed := false
	defer func() {
		if !completed {
			_ = os.Remove(destination)
			_ = os.Remove(destination + ".json")
		}
	}()
	_, err = db.sql.ExecContext(ctx, `VACUUM INTO ?`, destination)
	if err != nil {
		return err
	}
	snapshot, err := os.OpenFile(destination, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	err = snapshot.Sync()
	closeErr := snapshot.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	var schema int
	err = db.sql.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&schema)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(BackupMetadata{App: "apparat-phase8", Schema: schema, SHA256: hex.EncodeToString(sum[:])})
	if err != nil {
		return err
	}
	_, err = metadataFile.Write(metadata)
	if err == nil {
		err = metadataFile.Sync()
	}
	if err != nil {
		return err
	}
	completed = true
	return nil
}
func RestoreBackup(ctx context.Context, source, destination string) error {
	metadataBytes, err := os.ReadFile(source + ".json")
	if err != nil {
		return err
	}
	var metadata BackupMetadata
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return err
	}
	if metadata.App != "apparat-phase8" || (metadata.Schema < 3 || metadata.Schema > 7) {
		return errors.New("unsupported backup app/schema")
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != metadata.SHA256 {
		return errors.New("backup checksum mismatch")
	}
	// Never overwrite an authoritative destination; validate a new disposable copy.
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err == nil {
		err = syncErr
	}
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(destination)
		return err
	}
	db, err := Open(ctx, destination)
	if err != nil {
		_ = os.Remove(destination)
		return err
	}
	err = db.Integrity(ctx)
	if err == nil {
		var schema int
		err = db.sql.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&schema)
		if err == nil && schema != metadata.Schema {
			err = fmt.Errorf("backup schema mismatch")
		}
	}
	closeErr = db.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(destination)
		return err
	}
	directory, err := os.Open(filepath.Dir(destination))
	if err == nil {
		_ = directory.Sync()
		_ = directory.Close()
	}
	return nil
}
