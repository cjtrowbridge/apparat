// Persists revisioned scope defaults and session choices together.
// Session IDs and workspace thread names are namespaced by their owner and project.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const modelSchema = `
CREATE TABLE chat_model_defaults(cluster TEXT NOT NULL,project TEXT NOT NULL,model BLOB NOT NULL,revision INTEGER NOT NULL,operation TEXT NOT NULL,proof TEXT NOT NULL,PRIMARY KEY(cluster,project));
CREATE TABLE chat_session_models(cluster TEXT NOT NULL,project TEXT NOT NULL,session TEXT NOT NULL,model BLOB NOT NULL,PRIMARY KEY(cluster,project,session));
CREATE TABLE project_chat_turn_models(project TEXT NOT NULL,id TEXT NOT NULL,model BLOB NOT NULL,PRIMARY KEY(project,id),FOREIGN KEY(project,id) REFERENCES project_entries(project,id) ON DELETE CASCADE);
`

type modelReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readModelSettings(ctx context.Context, r modelReader, cluster, project, session string) (s domain.ModelSettings, err error) {
	var raw []byte
	err = r.QueryRowContext(ctx, "SELECT model,revision,operation,proof FROM chat_model_defaults WHERE cluster=? AND project=?", cluster, project).Scan(&raw, &s.Default.Revision, &s.Default.Operation, &s.Default.Proof)
	if err == nil {
		err = json.Unmarshal(raw, &s.Default.Model)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return s, err
	}
	err = r.QueryRowContext(ctx, "SELECT model FROM chat_session_models WHERE cluster=? AND project=? AND session=?", cluster, project, session).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err == nil {
		err = json.Unmarshal(raw, &s.Session)
	}
	return s, err
}
func (db *DB) ChatModels(ctx context.Context, cluster, project, session string) (domain.ModelSettings, error) {
	return readModelSettings(ctx, db.sql, cluster, project, session)
}
func (db *DB) SaveChatModel(ctx context.Context, cluster string, e domain.ModelEdit) (s domain.ModelSettings, err error) {
	if !e.Model.Valid() || e.Session == "" || len(e.Session) > 256 || !domain.ValidID(e.Operation) {
		return s, errors.New("invalid chat model choice")
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return s, err
	}
	defer tx.Rollback()
	s, err = readModelSettings(ctx, tx, cluster, e.Project, e.Session)
	if err != nil {
		return s, err
	}
	raw, err := json.Marshal(e.Model)
	if err != nil {
		return s, err
	}
	if e.Default {
		if s.Default.Operation == e.Operation && s.Default.Model == e.Model {
			return s, nil
		}
		if s.Default != e.Expected {
			return s, domain.ErrModelConflict
		}
		s.Default = domain.ModelDefault{Model: e.Model, Revision: s.Default.Revision + 1, Operation: e.Operation, Proof: e.Proof}
		_, err = tx.ExecContext(ctx, "INSERT INTO chat_model_defaults VALUES(?,?,?,?,?,?) ON CONFLICT(cluster,project) DO UPDATE SET model=excluded.model,revision=excluded.revision,operation=excluded.operation,proof=excluded.proof", cluster, e.Project, raw, s.Default.Revision, e.Operation, e.Proof)
		if err != nil {
			return s, err
		}
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO chat_session_models VALUES(?,?,?,?) ON CONFLICT(cluster,project,session) DO UPDATE SET model=excluded.model", cluster, e.Project, e.Session, raw)
	if err != nil {
		return s, err
	}
	s.Session = e.Model
	return s, tx.Commit()
}

// MergeGlobalModel orders concurrent offline edits by revision and then operation ID.
// Normal user edits still compare the complete previously observed default.
func (db *DB) MergeGlobalModel(ctx context.Context, cluster string, d domain.ModelDefault) error {
	return mergeGlobalModel(ctx, db.sql, cluster, d)
}

type modelExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func mergeGlobalModel(ctx context.Context, exec modelExecutor, cluster string, d domain.ModelDefault) error {
	if !d.Model.Valid() || !domain.ValidID(d.Operation) || d.Revision <= 0 || d.Revision >= 1<<53 {
		return errors.New("invalid replicated default")
	}
	raw, err := json.Marshal(d.Model)
	if err != nil {
		return err
	}
	_, err = exec.ExecContext(ctx, `INSERT INTO chat_model_defaults VALUES(?,'',?,?,?,?) ON CONFLICT(cluster,project) DO UPDATE SET model=excluded.model,revision=excluded.revision,operation=excluded.operation,proof=excluded.proof WHERE (excluded.revision,excluded.operation)>(revision,operation)`, cluster, raw, d.Revision, d.Operation, d.Proof)
	return err
}

// BindChatModel preserves the first choice even when two views open one chat concurrently.
func (db *DB) BindChatModel(ctx context.Context, cluster, project, session string, fallback domain.ModelRef) (domain.ModelSettings, error) {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return domain.ModelSettings{}, err
	}
	defer tx.Rollback()
	s, err := readModelSettings(ctx, tx, cluster, project, session)
	if err != nil {
		return s, err
	}
	if s.Session.Valid() || session == "" {
		return s, nil
	}
	m := s.Default.Model
	if !m.Valid() {
		m = fallback
	}
	if !m.Valid() {
		return s, nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return s, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO chat_session_models VALUES(?,?,?,?)", cluster, project, session, raw)
	if err != nil {
		return s, err
	}
	s.Session = m
	return s, tx.Commit()
}

// Authorizes a relayed signed device binding, including devices first seen through a relay.
func (db *DB) ModelAdmin(ctx context.Context, cluster, user, device string) bool {
	var ok bool
	err := db.sql.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cluster_users WHERE cluster=? AND user_id=? AND admin=1 AND admission='approved' AND NOT EXISTS(SELECT 1 FROM revoked_devices WHERE cluster=? AND device=?))`, cluster, user, cluster, device).Scan(&ok)
	return err == nil && ok
}

func (db *DB) HostModelCluster(ctx context.Context) (string, error) {
	var cluster string
	err := db.sql.QueryRowContext(ctx, "SELECT coalesce((SELECT cluster FROM worker_cluster WHERE singleton=1),'')").Scan(&cluster)
	return cluster, err
}

// Keep the requested model and pending message in one commit, including on redelivery.
func (db *DB) AddChatEntry(ctx context.Context, q domain.WorkspaceRequest, kind, body string) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, "INSERT INTO project_entries VALUES(?,?,?,?) ON CONFLICT(project,id) DO NOTHING", q.ProjectID, q.ID, kind, body)
	if err != nil {
		return err
	}
	added, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if added == 0 {
		var oldKind, oldBody string
		err = tx.QueryRowContext(ctx, "SELECT kind,body FROM project_entries WHERE project=? AND id=?", q.ProjectID, q.ID).Scan(&oldKind, &oldBody)
		if err != nil || oldKind != kind || oldBody != body {
			return errors.New("message ID already used")
		}
		return nil
	}
	model := q.Model
	if model != (domain.ModelRef{}) && !model.Valid() {
		return errors.New("invalid message model")
	}
	if !model.Valid() {
		s, err := readModelSettings(ctx, tx, "", q.ProjectID, q.Thread)
		if err != nil {
			return err
		}
		model = s.Session
	}
	if model.Valid() {
		raw, err := json.Marshal(model)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO project_chat_turn_models VALUES(?,?,?)", q.ProjectID, q.ID, raw)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Commit the authenticated acknowledgement and the initiating conversation's choice atomically.
func (db *DB) AcceptGlobalChoice(ctx context.Context, cluster, session string, d domain.ModelDefault) (s domain.ModelSettings, err error) {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return s, err
	}
	defer tx.Rollback()
	err = mergeGlobalModel(ctx, tx, cluster, d)
	if err != nil {
		return s, err
	}
	raw, err := json.Marshal(d.Model)
	if err != nil {
		return s, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO chat_session_models VALUES(?,'',?,?) ON CONFLICT(cluster,project,session) DO UPDATE SET model=excluded.model", cluster, session, raw)
	if err != nil {
		return s, err
	}
	s, err = readModelSettings(ctx, tx, cluster, "", session)
	if err != nil {
		return s, err
	}
	s.CanDefault = true
	return s, tx.Commit()
}
