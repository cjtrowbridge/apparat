// Persists local checkout registrations and account references.
// Location uniqueness protects repeated and concurrent registration.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
)

const projectsSchema = `
CREATE TABLE project_registry (
 id TEXT PRIMARY KEY, owner TEXT NOT NULL, name TEXT NOT NULL,
 location TEXT NOT NULL UNIQUE, account TEXT NOT NULL
);
INSERT INTO local_runtime(key,value,updated_at_ms)
 VALUES('project_owner',lower(hex(randomblob(16))),0);
`

func (db *DB) Projects(ctx context.Context) ([]domain.Project, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT id,owner,name,location,account FROM project_registry ORDER BY name, id")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		err := rows.Scan(&p.ID, &p.Owner, &p.Name, &p.Location, &p.Account)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (db *DB) RegisterProject(ctx context.Context, p domain.Project) error {
	_, err := db.sql.ExecContext(ctx, `INSERT INTO project_registry(id,owner,name,location,account)
 SELECT ?,value,?,?,? FROM local_runtime WHERE key='project_owner'
 ON CONFLICT(location) DO NOTHING`, p.ID, p.Name, p.Location, p.Account)
	return err
}
func (db *DB) UpdateProject(ctx context.Context, id, location, name, account string) error {
	result, err := db.sql.ExecContext(ctx, "UPDATE project_registry SET location=?,name=?,account=? WHERE id=?", location, name, account, id)
	return changedProject(result, err)
}
func (db *DB) RemoveProject(ctx context.Context, id string) error {
	result, err := db.sql.ExecContext(ctx, "DELETE FROM project_registry WHERE id=?", id)
	return changedProject(result, err)
}
func changedProject(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count != 1 {
		return errors.New("registration no longer exists")
	}
	return err
}
func (db *DB) ProjectDefault(ctx context.Context) (string, error) {
	var name string
	err := db.sql.QueryRowContext(ctx, "SELECT value FROM local_runtime WHERE key='project_account'").Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return name, err
}
func (db *DB) SetProjectDefault(ctx context.Context, name string) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO local_runtime VALUES('project_account',?,0) ON CONFLICT(key) DO UPDATE SET value=excluded.value", name)
	return err
}

// RegisteredThreads reads names for all registered projects without loading messages.
func (db *DB) RegisteredThreads(ctx context.Context) (map[string][]domain.WorkspaceChoice, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT DISTINCT project,kind FROM project_entries JOIN project_registry ON project_registry.id=project_entries.project WHERE kind LIKE 'chat:%' ORDER BY project,kind`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	threads := map[string][]domain.WorkspaceChoice{}
	for rows.Next() {
		var id, kind string
		err = rows.Scan(&id, &kind)
		if err != nil {
			return nil, err
		}
		name := strings.TrimPrefix(kind, "chat:")
		threads[id] = append(threads[id], domain.WorkspaceChoice{ID: name, Label: name})
	}
	return threads, rows.Err()
}
