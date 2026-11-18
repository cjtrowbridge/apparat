// Owns ordered forward migrations including transitional development tables.
// The original Phase 3 SQL bytes remain unchanged to preserve its checksum.
package persistence

import "context"

const phase3Schema = `
CREATE TABLE IF NOT EXISTS local_runtime (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at_ms INTEGER NOT NULL
);
`
const legacySchema = `
CREATE TABLE IF NOT EXISTS cluster_devices (id TEXT PRIMARY KEY, profile_json TEXT NOT NULL, signature TEXT NOT NULL, updated_at_ms INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS cluster_changes (sequence INTEGER PRIMARY KEY AUTOINCREMENT, device_id TEXT NOT NULL, kind TEXT NOT NULL, created_at_ms INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS cluster_cursors (name TEXT PRIMARY KEY, sequence INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS messages (id TEXT PRIMARY KEY, direction TEXT NOT NULL, payload TEXT NOT NULL, attempts INTEGER NOT NULL DEFAULT 0, next_attempt_ms INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS replay_seen (id TEXT PRIMARY KEY, seen_at_ms INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS event_cursors (name TEXT PRIMARY KEY, cursor TEXT NOT NULL);
`
const serviceSchema = `
CREATE TABLE services (id TEXT PRIMARY KEY, name TEXT NOT NULL, driver TEXT NOT NULL, enabled INTEGER NOT NULL, fail INTEGER NOT NULL, revision INTEGER NOT NULL);
CREATE TABLE service_observations (service_id TEXT PRIMARY KEY REFERENCES services(id) ON DELETE CASCADE, revision INTEGER NOT NULL, at_ms INTEGER NOT NULL, healthy INTEGER NOT NULL, safe_error TEXT NOT NULL);
CREATE TABLE service_capabilities (service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE, id TEXT NOT NULL, workload TEXT NOT NULL, PRIMARY KEY(service_id,id));
`

func (db *DB) Migrate(ctx context.Context) error {
	return db.ApplyMigrations(ctx, []Migration{
		{Version: 1, Name: "phase3_core", SQL: phase3Schema},
		{Version: 2, Name: "transitional_repositories", SQL: legacySchema},
		{Version: 3, Name: "local_services", SQL: serviceSchema},
		{Version: 4, Name: "local_preferences", SQL: preferencesSchema},
		{Version: 5, Name: "tab_placement", SQL: "ALTER TABLE preferences ADD COLUMN placement TEXT NOT NULL DEFAULT 'top'"},
		{Version: 6, Name: "local_project_registry", SQL: projectsSchema},
	})
}
