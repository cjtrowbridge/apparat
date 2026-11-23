// Maintains per-reader queue heads alongside the committed SQLite event journal.
// An unchanged reader checks its authorization and cursor without reading event rows.
package persistence

import (
	"context"
	"database/sql"
	"errors"
)

type queueView struct{ cluster, pool, user string }

func (db *DB) loadQueueHeads(ctx context.Context) error {
	db.queueMu.Lock()
	defer db.queueMu.Unlock()
	rows, err := db.sql.QueryContext(ctx, "SELECT cluster,pool,creator,MAX(revision) FROM pool_queue_events GROUP BY cluster,pool,creator")
	if err != nil {
		return err
	}
	defer rows.Close()
	heads := make(map[queueView]int64)
	for rows.Next() {
		var v queueView
		var revision int64
		err = rows.Scan(&v.cluster, &v.pool, &v.user, &revision)
		if err != nil {
			return err
		}
		heads[v] = revision
	}
	err = rows.Err()
	if err == nil {
		db.queueHeads = heads
	}
	return err
}

func (db *DB) commitQueue(ctx context.Context, tx *sql.Tx, v queueView) error {
	var head int64
	err := tx.QueryRowContext(ctx, "SELECT revision FROM pool_queue_events WHERE cluster=? AND pool=? AND creator=? ORDER BY revision DESC LIMIT 1", v.cluster, v.pool, v.user).Scan(&head)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil {
		db.queueHeads[v] = head
	}
	return err
}

func queueAccess(ctx context.Context, tx *sql.Tx, a Access, pool string) error {
	var allowed bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cluster_users u JOIN cluster_devices_authorized d USING(cluster,user_id)
	WHERE u.cluster=? AND u.user_id=? AND d.device_id=? AND u.admission='approved'
	AND NOT EXISTS(SELECT 1 FROM revoked_devices WHERE cluster=u.cluster AND device=d.device_id)
	AND EXISTS(SELECT 1 FROM pool_services WHERE cluster=u.cluster AND pool=? AND service=?)
	AND (u.admin=1 OR EXISTS(SELECT 1 FROM effective_pool_grants WHERE cluster=u.cluster AND pool=? AND service=? AND principal IN(u.user_id,'public'))))`, a.Cluster, a.User, a.Device, pool, HashService, pool, HashService).Scan(&allowed)
	if err != nil || !allowed {
		return errors.New("pool service is unavailable")
	}
	return nil
}
