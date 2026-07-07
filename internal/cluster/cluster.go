package cluster

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Capability struct {
	WorkloadClass string            `json:"workload_class"`
	Runtime       string            `json:"runtime"`
	Models        []string          `json:"models,omitempty"`
	Limits        map[string]string `json:"limits,omitempty"`
	Healthy       bool              `json:"healthy"`
}
type DeviceProfile struct {
	ID                     string       `json:"id"`
	Signature              string       `json:"signature"`
	Roles                  []string     `json:"roles"`
	Permissions            []string     `json:"permissions"`
	Endpoints              []string     `json:"endpoints"`
	CertificateFingerprint string       `json:"certificate_fingerprint"`
	WireGuardKey           string       `json:"wireguard_key"`
	Capabilities           []Capability `json:"capabilities"`
	LastSeenMs             int64        `json:"last_seen_ms"`
	Reachable              bool         `json:"reachable"`
}
type Change struct {
	Sequence    int64
	DeviceID    string
	Kind        string
	CreatedAtMs int64
}

type Directory struct{ db *sql.DB }

func New(db *sql.DB) Directory { return Directory{db: db} }

func (dir Directory) Init(ctx context.Context) error {
	_, err := dir.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS cluster_devices (id TEXT PRIMARY KEY, profile_json TEXT NOT NULL, signature TEXT NOT NULL, updated_at_ms INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS cluster_changes (sequence INTEGER PRIMARY KEY AUTOINCREMENT, device_id TEXT NOT NULL, kind TEXT NOT NULL, created_at_ms INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS cluster_cursors (name TEXT PRIMARY KEY, sequence INTEGER NOT NULL);
`)
	return err
}

func (dir Directory) PutDevice(ctx context.Context, profile DeviceProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	now := time.Now().UTC().UnixMilli()
	_, err = dir.db.ExecContext(ctx, `INSERT INTO cluster_devices(id, profile_json, signature, updated_at_ms) VALUES (?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET profile_json=excluded.profile_json, signature=excluded.signature, updated_at_ms=excluded.updated_at_ms`, profile.ID, string(data), profile.Signature, now)
	if err != nil {
		return err
	}
	_, err = dir.db.ExecContext(ctx, `INSERT INTO cluster_changes(device_id, kind, created_at_ms) VALUES (?, ?, ?)`, profile.ID, "device_profile", now)
	return err
}

func (dir Directory) GetDevice(ctx context.Context, id string) (DeviceProfile, error) {
	var raw string
	if err := dir.db.QueryRowContext(ctx, `SELECT profile_json FROM cluster_devices WHERE id = ?`, id).Scan(&raw); err != nil {
		return DeviceProfile{}, err
	}
	var profile DeviceProfile
	return profile, json.Unmarshal([]byte(raw), &profile)
}

func (dir Directory) ChangesAfter(ctx context.Context, after int64) ([]Change, error) {
	rows, err := dir.db.QueryContext(ctx, `SELECT sequence, device_id, kind, created_at_ms FROM cluster_changes WHERE sequence > ? ORDER BY sequence`, after)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var changes []Change
	for rows.Next() {
		var change Change
		if err := rows.Scan(&change.Sequence, &change.DeviceID, &change.Kind, &change.CreatedAtMs); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}
