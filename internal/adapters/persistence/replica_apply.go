// Applies a pinned worker's view and authorized queue pages atomically.
// Newly hidden pools lose their cached rows when a new view is installed.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"slices"
)

type SyncPage struct {
	Snapshot *SignedReplica           `json:",omitempty"`
	Queues   map[string][]QueueChange `json:",omitempty"`
}

func (db *DB) ApplyReplica(ctx context.Context, w string, a Access, cursor SyncCursor, page SyncPage) error {
	if page.Snapshot == nil && len(page.Queues) == 0 {
		return nil
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var stored uint64
	var raw []byte
	err = tx.QueryRowContext(ctx, "SELECT revision,body FROM worker_replicas WHERE worker=? AND user=? AND device=?", w, a.User, a.Device).Scan(&stored, &raw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if stored != cursor.Revision {
		return errors.New("concurrent replica change")
	}
	var r SignedReplica
	if page.Snapshot != nil {
		r = *page.Snapshot
		if !r.Valid(w, a) || r.View.Revision < stored || r.View.Admin != a.Admin {
			return errors.New("invalid worker snapshot")
		}
		raw, err = json.Marshal(r)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO worker_replicas VALUES(?,?,?,?,?) ON CONFLICT(worker,user,device) DO UPDATE SET revision=excluded.revision,body=excluded.body", w, a.User, a.Device, r.View.Revision, raw)
		if err != nil {
			return err
		}
		if !r.View.Admin {
			_, err = tx.ExecContext(ctx, "DELETE FROM replica_policy WHERE worker=? AND user=? AND device=?", w, a.User, a.Device)
			if err != nil {
				return err
			}
		}
		for pool := range cursor.Queues {
			if !slices.Contains(r.View.Pools, pool) {
				_, err = tx.ExecContext(ctx, "DELETE FROM replica_queue WHERE worker=? AND user=? AND device=? AND pool=?", w, a.User, a.Device, pool)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = json.Unmarshal(raw, &r)
		if err != nil || !r.Valid(w, a) || a.Admin != r.View.Admin {
			return errors.New("missing current authorization snapshot")
		}
	}
	for pool, changes := range page.Queues {
		if !slices.Contains(r.View.Pools, pool) {
			return errors.New("unexpected pool")
		}
		var head int64
		err = tx.QueryRowContext(ctx, "SELECT coalesce(max(revision),0) FROM replica_queue WHERE worker=? AND user=? AND device=? AND pool=?", w, a.User, a.Device, pool).Scan(&head)
		if err != nil || head != cursor.Queues[pool] {
			return errors.New("stale queue page")
		}
		for _, e := range changes {
			if e.Revision <= head || e.ID == "" {
				return errors.New("invalid queue revision")
			}
			data, err := json.Marshal(e)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, "INSERT INTO replica_queue VALUES(?,?,?,?,?,?)", w, a.User, a.Device, pool, e.Revision, data)
			if err != nil {
				return err
			}
			head = e.Revision
		}
	}
	return tx.Commit()
}

func (db *DB) ForgetReplica(ctx context.Context, w string, a Access) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, stmt := range []string{"DELETE FROM replica_policy WHERE worker=? AND user=? AND device=?", "DELETE FROM replica_queue WHERE worker=? AND user=? AND device=?", "DELETE FROM worker_replicas WHERE worker=? AND user=? AND device=?"} {
		_, err = tx.ExecContext(ctx, stmt, w, a.User, a.Device)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, "UPDATE claimed_workers SET hardware=NULL WHERE worker=? AND owner=?", w, a.User)
	if err != nil {
		return err
	}
	return tx.Commit()
}
