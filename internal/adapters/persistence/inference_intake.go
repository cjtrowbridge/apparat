// Records a cluster-side receipt before any dependency on a transcript writer or model.
// Stable turn identities survive lost acknowledgements and repeated peer forwarding.
package persistence

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const inferenceMembershipSchema = `
CREATE TABLE inference_withdrawals(worker TEXT PRIMARY KEY,generation TEXT NOT NULL,phase TEXT NOT NULL);
ALTER TABLE planning_outbox ADD COLUMN received INTEGER NOT NULL DEFAULT 0;
ALTER TABLE planning_outbox ADD COLUMN acknowledged_ms INTEGER NOT NULL DEFAULT 0;
ALTER TABLE inference_roster RENAME TO inference_membership;
CREATE TABLE session_homes(user TEXT,session TEXT,worker TEXT NOT NULL,PRIMARY KEY(user,session));
CREATE TABLE inference_intake(received_ms INTEGER NOT NULL DEFAULT (unixepoch('subsec')*1000),user TEXT, session TEXT, turn TEXT, home TEXT NOT NULL, body BLOB NOT NULL, applied INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(user,session,turn));
`

type InferenceReceipt struct {
	User    string
	Request domain.GlobalRequest
}

func (db *DB) ReceiveInference(ctx context.Context, user string, q domain.GlobalRequest) error {
	body, err := json.Marshal(q)
	if err != nil {
		return err
	}
	result, err := db.sql.ExecContext(ctx, `INSERT INTO inference_intake(user,session,turn,home,body) VALUES(?,?,?,?,?)
 ON CONFLICT(user,session,turn) DO UPDATE SET turn=excluded.turn WHERE body=excluded.body`, user, q.Session, q.Turn, q.Home, body)
	return inferenceChange(result, err)
}

func (db *DB) PendingInferenceReceipts(ctx context.Context) ([]InferenceReceipt, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT user,body FROM inference_intake WHERE applied=0 ORDER BY rowid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InferenceReceipt
	for rows.Next() {
		var r InferenceReceipt
		var b []byte
		err = rows.Scan(&r.User, &b)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &r.Request)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (db *DB) AppliedInferenceReceipt(ctx context.Context, user, session, turn string) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE inference_intake SET applied=1 WHERE user=? AND session=? AND turn=?", user, session, turn)
	return err
}

func (db *DB) ReceiptTask(ctx context.Context, user, id string) (*domain.InferenceRecord, error) {
	var raw []byte
	err := db.sql.QueryRowContext(ctx, "SELECT body FROM inference_intake WHERE user=? AND turn=?", user, id).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var g domain.GlobalRequest
	err = json.Unmarshal(raw, &g)
	if err != nil {
		return nil, err
	}
	packet := domain.ChatRequest{ChatAddress: domain.ChatAddress{Session: g.Session, Turn: g.Turn}, Job: g.Turn, Model: g.Model, Messages: []domain.ChatMessage{{Role: "user", Text: g.Input}}}
	return &domain.InferenceRecord{Job: domain.InferenceJob{Packet: packet, User: user, Pool: "Chat Pool", Owner: g.Home}, State: "queued"}, nil
}
