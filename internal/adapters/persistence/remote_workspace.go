// Stores per-project host permissions, client drafts and bounded save receipts.
// The worker checks a grant alongside the current authenticated device session.
package persistence

import (
	"context"
	"database/sql"
	"errors"
)

const remoteWorkspaceSchema = `
CREATE INDEX workspace_entry_page ON project_entries(project,kind,id DESC);
CREATE TABLE workspace_grants(project TEXT REFERENCES project_registry(id) ON DELETE CASCADE, user TEXT NOT NULL, scope TEXT NOT NULL, PRIMARY KEY(project,user,scope));
CREATE TABLE remote_drafts(user TEXT, project TEXT, path TEXT, base TEXT NOT NULL, body TEXT NOT NULL, operation TEXT NOT NULL, PRIMARY KEY(user,project,path));
CREATE TABLE remote_saves(project TEXT REFERENCES project_registry(id) ON DELETE CASCADE,user TEXT,op TEXT,path TEXT NOT NULL,base TEXT NOT NULL,result TEXT NOT NULL,done INTEGER NOT NULL,PRIMARY KEY(project,user,op));`

func (db *DB) WorkspaceGrant(ctx context.Context, project, user, scope string, enable bool) error {
	switch scope {
	case "read", "write", "tasks", "terminal-view", "terminal-control":
	default:
		return errors.New("choose read, write, tasks, terminal-view or terminal-control")
	}
	if !enable {
		_, err := db.sql.ExecContext(ctx, "DELETE FROM workspace_grants WHERE project=? AND user=? AND scope=?", project, user, scope)
		return err
	}
	result, err := db.sql.ExecContext(ctx, `INSERT OR REPLACE INTO workspace_grants SELECT p.id,u.user_id,? FROM project_registry p,cluster_users u WHERE p.id=? AND u.user_id=? AND u.admission='approved'`, scope, project, user)
	return changedProject(result, err)
}
func (db *DB) WorkspaceLocation(ctx context.Context, project, user, scope string) (string, error) {
	var location string
	err := db.sql.QueryRowContext(ctx, `SELECT p.location FROM project_registry p JOIN workspace_grants g ON p.id=g.project WHERE p.id=? AND g.user=? AND g.scope=?`, project, user, scope).Scan(&location)
	return location, err
}
func (db *DB) ProjectLocation(ctx context.Context, id string) (string, error) {
	var path string
	err := db.sql.QueryRowContext(ctx, "SELECT location FROM project_registry WHERE id=?", id).Scan(&path)
	return path, err
}
func (db *DB) PutRemoteDraft(ctx context.Context, user, project, path, base, body, operation string) error {
	_, err := db.sql.ExecContext(ctx, `INSERT INTO remote_drafts VALUES(?,?,?,?,?,?) ON CONFLICT(user,project,path) DO UPDATE SET base=excluded.base,body=excluded.body,operation=CASE WHEN remote_drafts.base=excluded.base AND remote_drafts.body=excluded.body AND excluded.operation='' THEN remote_drafts.operation ELSE excluded.operation END`, user, project, path, base, body, operation)
	return err
}
func (db *DB) RemoteDraft(ctx context.Context, user, project, path string) (base, body, operation string, found bool, err error) {
	err = db.sql.QueryRowContext(ctx, "SELECT base,body,operation FROM remote_drafts WHERE user=? AND project=? AND path=?", user, project, path).Scan(&base, &body, &operation)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", false, nil
	}
	return base, body, operation, err == nil, err
}
func (db *DB) ClearRemoteDraft(ctx context.Context, user, project, path, base, body string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM remote_drafts WHERE user=? AND project=? AND path=? AND base=? AND body=?", user, project, path, base, body)
	return err
}
func (db *DB) PrepareRemoteSave(ctx context.Context, project, user, op, path, base, result string) (existing, done bool, err error) {
	res, err := db.sql.ExecContext(ctx, "INSERT INTO remote_saves VALUES(?,?,?,?,?,?,0) ON CONFLICT DO NOTHING", project, user, op, path, base, result)
	if err != nil {
		return false, false, err
	}
	changed, err := res.RowsAffected()
	if err != nil || changed == 1 {
		return false, false, err
	}
	err = db.sql.QueryRowContext(ctx, "SELECT done FROM remote_saves WHERE project=? AND user=? AND op=? AND path=? AND base=? AND result=?", project, user, op, path, base, result).Scan(&done)
	return true, done, err
}
func (db *DB) FinishRemoteSave(ctx context.Context, project, user, op string) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE remote_saves SET done=1 WHERE project=? AND user=? AND op=?", project, user, op)
	return err
}
