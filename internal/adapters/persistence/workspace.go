// Stores owner-local drafts, pending chat messages and manual run outcomes.
// Every record belongs to a registered project and survives process restart.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
)

const workspaceSchema = `
CREATE TABLE project_drafts(project TEXT REFERENCES project_registry(id) ON DELETE CASCADE, path TEXT, base TEXT NOT NULL, body TEXT NOT NULL, PRIMARY KEY(project,path));
CREATE TABLE project_entries(project TEXT REFERENCES project_registry(id) ON DELETE CASCADE, id TEXT, kind TEXT NOT NULL, body TEXT NOT NULL, PRIMARY KEY(project,id));`

func (db *DB) Draft(ctx context.Context, id, path string) (base, body string, found bool, err error) {
	err = db.sql.QueryRowContext(ctx, "SELECT base,body FROM project_drafts WHERE project=? AND path=?", id, path).Scan(&base, &body)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", false, nil
	}
	return base, body, err == nil, err
}
func (db *DB) SaveDraft(ctx context.Context, id, path, base, body string) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO project_drafts VALUES(?,?,?,?) ON CONFLICT(project,path) DO UPDATE SET base=excluded.base,body=excluded.body", id, path, base, body)
	return err
}
func (db *DB) ClearDraft(ctx context.Context, id, path string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM project_drafts WHERE project=? AND path=?", id, path)
	return err
}
func (db *DB) AddProjectEntry(ctx context.Context, project, id, kind, body string) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO project_entries VALUES(?,?,?,?) ON CONFLICT(project,id) DO NOTHING", project, id, kind, body)
	return err
}
func (db *DB) ProjectEntry(ctx context.Context, project, id string) (kind, body string, err error) {
	err = db.sql.QueryRowContext(ctx, "SELECT kind,body FROM project_entries WHERE project=? AND id=?", project, id).Scan(&kind, &body)
	return
}
func (db *DB) ProjectEntries(ctx context.Context, project, kind string) (string, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT id,body FROM project_entries WHERE project=? AND kind=? ORDER BY rowid DESC LIMIT 30", project, kind)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = rows.Close()
	}()
	var result strings.Builder
	for rows.Next() {
		var id, body string
		err = rows.Scan(&id, &body)
		if err != nil {
			return "", err
		}
		result.WriteString(id + " " + body + "\n")
	}
	return result.String(), rows.Err()
}

func (db *DB) FinishWorkspaceSave(ctx context.Context, project, id, path, base, body, record string, applied bool) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	_, err = tx.ExecContext(ctx, "UPDATE project_entries SET body=? WHERE project=? AND id=? AND kind='save'", record, project, id)
	if err != nil {
		return err
	}
	if applied {
		_, err = tx.ExecContext(ctx, "DELETE FROM project_drafts WHERE project=? AND path=? AND base=? AND body=?", project, path, base, body)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) EntryPage(ctx context.Context, project, kind, before string) ([]domain.WorkspaceChoice, string, error) {
	query := "SELECT id,body FROM project_entries WHERE project=? AND kind=?"
	args := []any{project, kind}
	if before != "" {
		query += " AND id<?"
		args = append(args, before)
	}
	rows, err := db.sql.QueryContext(ctx, query+" ORDER BY id DESC LIMIT 31", args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	var result []domain.WorkspaceChoice
	for rows.Next() {
		if len(result) == 30 {
			return result, result[29].ID, nil
		}
		var item domain.WorkspaceChoice
		err = rows.Scan(&item.ID, &item.Label)
		if err != nil {
			return nil, "", err
		}
		result = append(result, item)
	}
	return result, "", rows.Err()
}

func (db *DB) UpdateRun(ctx context.Context, project, id, body string) error {
	result, err := db.sql.ExecContext(ctx, "UPDATE project_entries SET body=? WHERE project=? AND id=? AND kind='run'", body, project, id)
	return changedProject(result, err)
}
func (db *DB) ArtifactInUse(ctx context.Context, digest string) (bool, error) {
	var used bool
	err := db.sql.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM project_entries WHERE kind='run' AND json_extract(body,'$.Digest')=? AND json_extract(body,'$.State')='completed')", digest).Scan(&used)
	return used, err
}
func (db *DB) RecoverWorkspaceRuns(ctx context.Context) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE project_entries SET body=json_set(body,'$.State','interrupted') WHERE kind='run' AND json_extract(body,'$.State')='running'")
	return err
}
func (db *DB) ChatThreads(ctx context.Context, project string) ([]domain.WorkspaceChoice, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT DISTINCT kind FROM project_entries WHERE project=? AND kind LIKE 'chat:%' ORDER BY kind", project)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var threads []domain.WorkspaceChoice
	for rows.Next() {
		var kind string
		err = rows.Scan(&kind)
		if err != nil {
			return nil, err
		}
		name := strings.TrimPrefix(kind, "chat:")
		threads = append(threads, domain.WorkspaceChoice{ID: name, Label: name})
	}
	return threads, rows.Err()
}

// PrepareWorkspaceSave commits both the recoverable content and its pending intent.
func (db *DB) PrepareWorkspaceSave(ctx context.Context, q domain.WorkspaceRequest, record string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	_, err = tx.ExecContext(ctx, "INSERT INTO project_drafts VALUES(?,?,?,?) ON CONFLICT(project,path) DO UPDATE SET base=excluded.base,body=excluded.body", q.ProjectID, q.Path, q.Base, q.Text)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO project_entries VALUES(?,?,'save',?)", q.ProjectID, q.ID, record)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (db *DB) PendingFileSaves(ctx context.Context, project, path string) ([]domain.WorkspaceChoice, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT id,body FROM project_entries WHERE project=? AND kind='save' AND json_extract(body,'$.Path')=? AND json_extract(body,'$.State')='pending' ORDER BY rowid", project, path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var items []domain.WorkspaceChoice
	for rows.Next() {
		var r domain.WorkspaceChoice
		err = rows.Scan(&r.ID, &r.Label)
		if err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, rows.Err()
}
