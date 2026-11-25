// Preserves pending planning sends and renames on the originating signed-in device.
// Stable request bodies survive GUI restarts and lost acknowledgements.
package persistence

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const planningOutboxSchema = `CREATE TABLE planning_outbox(seq INTEGER PRIMARY KEY,user TEXT,body BLOB,UNIQUE(user,body));`

type PlanningEntry struct {
	Seq       int64
	Operation string
	Request   domain.GlobalRequest
}

func (db *DB) SavePlanning(ctx context.Context, user, op string, q domain.GlobalRequest) error {
	b, err := json.Marshal(PlanningEntry{Operation: op, Request: q})
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "INSERT OR IGNORE INTO planning_outbox(user,body) VALUES(?,?)", user, b)
	return err
}
func (db *DB) PlanningEntries(ctx context.Context, user string) ([]PlanningEntry, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT seq,body FROM planning_outbox WHERE user=? ORDER BY seq", user)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var out []PlanningEntry
	for rows.Next() {
		var b []byte
		var seq int64
		err = rows.Scan(&seq, &b)
		var e PlanningEntry
		if err == nil {
			err = json.Unmarshal(b, &e)
		}
		if err != nil {
			return nil, err
		}
		e.Seq = seq
		out = append(out, e)
	}
	return out, rows.Err()
}
func (db *DB) AcknowledgePlanning(ctx context.Context, seq int64, user string) error {
	_, err := db.sql.ExecContext(ctx, "DELETE FROM planning_outbox WHERE seq=? AND user=?", seq, user)
	return err
}
