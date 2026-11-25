// Stores owner-side inference outbox entries across disconnects and process restarts.
// A turn references an immutable packet by job ID even when repository instructions change.
package persistence

import (
	"context"
	"encoding/json"
	"errors"
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

func (db *DB) BoundPacket(ctx context.Context, job, revision string) (q domain.ChatRequest, bound string, err error) {
	var body []byte
	err = db.sql.QueryRowContext(ctx, "UPDATE chat_packets SET revision=CASE WHEN revision='' THEN ? ELSE revision END WHERE job=? RETURNING body,revision", revision, job).Scan(&body, &bound)
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

// Keep existing transcript writers and replay pending packets to newly joined replicas.
func (db *DB) PrepareMembershipChange(ctx context.Context, previous domain.InferenceRoster) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var busy bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM inference_jobs WHERE state IN ('owned','dispatching') OR state='finished' AND delivered=0 AND json_extract(claim,'$.Executor')=(SELECT worker FROM worker_cluster))").Scan(&busy)
	if err != nil {
		return err
	}
	if busy {
		return errors.New("worker has outstanding ownership receipts")
	}
	rows, err := tx.QueryContext(ctx, "SELECT user,session FROM global_chats")
	if err != nil {
		return err
	}
	var homes [][2]string
	for rows.Next() {
		var h [2]string
		err = rows.Scan(&h[0], &h[1])
		if err != nil {
			rows.Close()
			return err
		}
		homes = append(homes, h)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, h := range homes {
		_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO session_homes VALUES(?,?,?)", h[0], h[1], previous.GlobalWriter(h[0], h[1]))
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, "UPDATE chat_packets SET admitted=0; UPDATE global_chats SET pending=1 WHERE EXISTS(SELECT 1 FROM session_homes h JOIN worker_cluster w ON w.worker=h.worker WHERE h.user=global_chats.user AND h.session=global_chats.session)")
	if err != nil {
		return err
	}
	return tx.Commit()
}
