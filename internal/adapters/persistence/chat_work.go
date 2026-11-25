// Stores owner-side inference outbox entries across disconnects and process restarts.
// A turn references an immutable packet by job ID even when repository instructions change.
package persistence

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const chatWorkSchema = `
CREATE TABLE chat_work(project TEXT,session TEXT,turn TEXT,user TEXT NOT NULL,pool TEXT NOT NULL,PRIMARY KEY(project,session,turn));
CREATE TABLE chat_packets(job TEXT PRIMARY KEY,body BLOB NOT NULL,revision TEXT NOT NULL DEFAULT '');`

type ChatWork struct {
	domain.ChatAddress
	User, Pool string
}

func (db *DB) AddChatWork(ctx context.Context, w ChatWork) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO chat_work VALUES(?,?,?,?,?) ON CONFLICT DO NOTHING", w.Project, w.Session, w.Turn, w.User, w.Pool)
	return err
}

func (db *DB) ChatWork(ctx context.Context) ([]ChatWork, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT project,session,turn,user,pool FROM chat_work ORDER BY turn")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var work []ChatWork
	for rows.Next() {
		var w ChatWork
		err = rows.Scan(&w.Project, &w.Session, &w.Turn, &w.User, &w.Pool)
		if err != nil {
			return nil, err
		}
		work = append(work, w)
	}
	return work, rows.Err()
}

func (db *DB) CompleteChatWork(ctx context.Context, w ChatWork) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM chat_work WHERE project=? AND session=? AND turn=?", w.Project, w.Session, w.Turn)
	return err
}

func (db *DB) FreezeChatPacket(ctx context.Context, job string, body []byte) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO chat_packets(job,body) VALUES(?,?)", job, body)
	return err
}

func (db *DB) BoundPacket(ctx context.Context, job, revision string) (q domain.ChatRequest, err error) {
	var body []byte
	err = db.sql.QueryRowContext(ctx, "UPDATE chat_packets SET revision=? WHERE job=? AND revision IN ('',?) RETURNING body", revision, job, revision).Scan(&body)
	if err == nil {
		err = json.Unmarshal(body, &q)
	}
	return
}

func (db *DB) PacketAdmitted(ctx context.Context, job string) (bool, error) {
	var admitted bool
	err := db.sql.QueryRowContext(ctx, "SELECT admitted FROM chat_packets WHERE job=?", job).Scan(&admitted)
	return admitted, err
}
func (db *DB) AcknowledgePacket(ctx context.Context, job string) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE chat_packets SET admitted=1 WHERE job=?", job)
	return err
}
