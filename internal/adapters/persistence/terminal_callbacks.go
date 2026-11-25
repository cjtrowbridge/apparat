// Keeps outstanding terminal completion callbacks on the initiating device.
// Identical approvals recover one immutable request across restarts and lost replies.
package persistence

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

const terminalCallbacksSchema = "CREATE TABLE terminal_callbacks(user TEXT, job TEXT, request BLOB NOT NULL, PRIMARY KEY(user,job)); CREATE TABLE chat_terminal_hosts(user TEXT,session TEXT,project TEXT NOT NULL,PRIMARY KEY(user,session));"

func (db *DB) RememberTerminalCallback(ctx context.Context, user string, q domain.WorkspaceRequest) error {
	body, err := json.Marshal(q)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "INSERT OR IGNORE INTO terminal_callbacks VALUES(?,?,?)", user, q.ID, body)
	if err != nil {
		return err
	}
	var existing []byte
	err = db.sql.QueryRowContext(ctx, "SELECT request FROM terminal_callbacks WHERE user=? AND job=?", user, q.ID).Scan(&existing)
	if err == nil && !bytes.Equal(existing, body) {
		return errors.New("terminal approval changed")
	}
	return err
}
func (db *DB) TerminalCallbacks(ctx context.Context, user string) ([]domain.WorkspaceRequest, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT request FROM terminal_callbacks WHERE user=? ORDER BY job", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WorkspaceRequest
	for rows.Next() {
		var raw []byte
		err = rows.Scan(&raw)
		if err != nil {
			return nil, err
		}
		var q domain.WorkspaceRequest
		err = json.Unmarshal(raw, &q)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}
func (db *DB) ForgetTerminalCallback(ctx context.Context, user, job string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM terminal_callbacks WHERE user=? AND job=?", user, job)
	return err
}
func (db *DB) OwnsChatWork(ctx context.Context, user string, a domain.ChatAddress) (bool, error) {
	var allowed bool
	err := db.sql.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM chat_work WHERE user=? AND project=? AND session=? AND turn=?)", user, a.Project, a.Session, a.Turn).Scan(&allowed)
	return allowed, err
}

func (db *DB) TerminalProject(ctx context.Context, user, session, choice string) (string, error) {
	if choice != "" {
		_, err := db.sql.ExecContext(ctx, "INSERT OR IGNORE INTO chat_terminal_hosts VALUES(?,?,?)", user, session, choice)
		if err != nil {
			return "", err
		}
	}
	var project string
	err := db.sql.QueryRowContext(ctx, "SELECT project FROM chat_terminal_hosts WHERE user=? AND session=?", user, session).Scan(&project)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err == nil && choice != "" && choice != project {
		return "", errors.New("this chat already has a terminal host")
	}
	return project, err
}
