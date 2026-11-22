// Reads task state and artifact references by the authenticated creator and pool.
// Queue cursors expose only events the caller may see.
package persistence

import (
	"context"
	"errors"
	"time"
)

type QueueChange struct {
	Revision, Fence, Changed int64
	ID, State                string
}

func (db *DB) QueueAfter(ctx context.Context, a Access, pool string, after int64) ([]QueueChange, error) {
	if after < 0 {
		return nil, errors.New("invalid queue cursor")
	}
	_, err := db.CurrentAccess(ctx, a.Cluster, a.User, a.Device)
	if err != nil {
		return nil, err
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT revision,task,state,fence,changed_ms FROM pool_queue_events
	WHERE cluster=? AND pool=? AND creator=? AND revision>? ORDER BY revision LIMIT 64`, a.Cluster, pool, a.User, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	changes := []QueueChange{}
	for rows.Next() {
		var change QueueChange
		err = rows.Scan(&change.Revision, &change.ID, &change.State, &change.Fence, &change.Changed)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}

func (db *DB) ResultReference(ctx context.Context, a Access, pool, id string) (string, error) {
	_, err := db.CurrentAccess(ctx, a.Cluster, a.User, a.Device)
	if err != nil {
		return "", err
	}
	var hash string
	err = db.sql.QueryRowContext(ctx, `SELECT r.digest FROM pool_task_artifacts r JOIN pool_tasks t
	ON r.cluster=t.cluster AND r.pool=t.pool AND r.task=t.id
	WHERE t.cluster=? AND t.pool=? AND t.id=? AND t.creator=? AND t.state='succeeded' AND r.slot='result'`, a.Cluster, pool, id, a.User).Scan(&hash)
	return hash, err
}

func (db *DB) MayDispatch(ctx context.Context, a Access, pool, worker string) bool {
	var allowed bool
	err := db.sql.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pool_services s WHERE cluster=? AND pool=? AND worker=? AND service=?
	AND EXISTS(SELECT 1 FROM pool_tasks t WHERE t.cluster=s.cluster AND t.pool=s.pool AND t.creator=?
	AND (t.state='queued' OR (t.state='running' AND lease_until<=?))))`, a.Cluster, pool, worker, HashService, a.User, time.Now().UnixMilli()).Scan(&allowed)
	return err == nil && allowed
}
