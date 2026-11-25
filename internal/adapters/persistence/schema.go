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

var migrations = []Migration{
	{Version: 1, Name: "phase3_core", SQL: phase3Schema},
	{Version: 2, Name: "transitional_repositories", SQL: legacySchema},
	{Version: 3, Name: "local_services", SQL: serviceSchema},
	{Version: 4, Name: "local_preferences", SQL: preferencesSchema},
	{Version: 5, Name: "tab_placement", SQL: "ALTER TABLE preferences ADD COLUMN placement TEXT NOT NULL DEFAULT 'top'"},
	{Version: 6, Name: "local_project_registry", SQL: projectsSchema},
	{Version: 7, Name: "local_workspaces", SQL: workspaceSchema},
	{Version: 8, Name: "spoken_choices", SQL: "ALTER TABLE preferences ADD COLUMN spoken_choices INTEGER NOT NULL DEFAULT 0"},
	{Version: 9, Name: "personal_device_bindings", SQL: personalSchema},
	{Version: 10, Name: "worker_claims", SQL: claimSchema},
	{Version: 11, Name: "cluster_policy", SQL: policySchema},
	{Version: 12, Name: "claim_receipts", SQL: claimReceiptSchema},
	{Version: 13, Name: "device_revocations", SQL: admissionSchema},
	{Version: 14, Name: "pool_task_queue", SQL: queueSchema},
	{Version: 15, Name: "device_hostnames", SQL: "ALTER TABLE cluster_devices_authorized ADD COLUMN hostname TEXT NOT NULL DEFAULT ''"},
	{Version: 16, Name: "worker_replicas", SQL: replicaSchema},
	{Version: 17, Name: "pool_device_directory", SQL: directorySchema},
	{Version: 18, Name: "cluster_device_names", SQL: "CREATE TABLE cluster_device_names(cluster TEXT NOT NULL REFERENCES worker_cluster(cluster),device TEXT NOT NULL,name TEXT NOT NULL,PRIMARY KEY(cluster,device))"},
	{Version: 19, Name: "worker_only_pool_directory", SQL: "UPDATE view_revision SET revision=revision+1"},
	{Version: 20, Name: "remote_workspace_access", SQL: remoteWorkspaceSchema},
	{Version: 21, Name: "chat_model_choices", SQL: modelSchema},
	{Version: 22, Name: "retire_personal_pool_view", SQL: "UPDATE view_revision SET revision=revision+1"},
	{Version: 23, Name: "distributed_inference_jobs", SQL: inferenceSchema},
	{Version: 24, Name: "chat_inference_outbox", SQL: chatWorkSchema},
	{Version: 25, Name: "global_session_artifacts", SQL: globalSchema},
	{Version: 26, Name: "planning_outbox", SQL: planningOutboxSchema},
	{Version: 27, Name: "inference_publication_cursors", SQL: "ALTER TABLE chat_packets ADD COLUMN admitted INTEGER NOT NULL DEFAULT 0; ALTER TABLE global_chats ADD COLUMN markdown_revision INTEGER NOT NULL DEFAULT 0;"},
	{Version: 28, Name: "automatic_inference_membership", SQL: inferenceMembershipSchema},
}

func (db *DB) Migrate(ctx context.Context) error {
	err := db.ApplyMigrations(ctx, migrations)
	if err != nil {
		return err
	}
	err = db.loadPolicyHeads(ctx)
	if err != nil {
		return err
	}
	err = db.loadQueueHeads(ctx)
	if err != nil {
		return err
	}
	db.policyMu.Lock()
	defer db.policyMu.Unlock()
	return db.refreshViewRevision(ctx)
}
