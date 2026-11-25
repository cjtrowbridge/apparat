// Reads task state and artifact references by the authenticated creator and pool.
// Queue cursors expose only events the caller may see.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"time"
)

type QueueChange struct {
	Revision, Fence, Changed int64
	ID, State                string
}

func (db *DB) QueueAfter(ctx context.Context, a Access, pool string, after int64) ([]QueueChange, error) {
	db.queueMu.RLock()
	defer db.queueMu.RUnlock()
	if after < 0 {
		return nil, errors.New("invalid queue cursor")
	}
	tx, err := db.sql.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	err = queueAccess(ctx, tx, a, pool)
	if err != nil {
		return nil, err
	}
	head := db.queueHeads[queueView{a.Cluster, pool, a.User}]
	if after == head {
		return []QueueChange{}, nil
	}
	if after > head {
		return nil, errors.New("unknown queue cursor")
	}
	rows, err := tx.QueryContext(ctx, `SELECT revision,task,state,fence,changed_ms FROM pool_queue_events
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
	tx, err := db.sql.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	err = queueAccess(ctx, tx, a, pool)
	if err != nil {
		return "", err
	}
	var hash string
	err = tx.QueryRowContext(ctx, `SELECT r.digest FROM pool_task_artifacts r JOIN pool_tasks t
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

// LocalPoolTasks folds the synchronized event stream to one latest row per task.
func (db *DB) LocalPoolTasks(ctx context.Context, worker string, a Access, pool string) ([]QueueChange, error) {
	rows, err := db.sql.QueryContext(ctx, `WITH ranked AS (
 SELECT body,row_number() OVER(PARTITION BY json_extract(body,'$.ID') ORDER BY revision DESC) AS n
 FROM replica_queue WHERE worker=? AND user=? AND device=? AND pool=?)
 SELECT body FROM ranked WHERE n=1`, worker, a.User, a.Device, pool)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []QueueChange
	for rows.Next() {
		var data []byte
		var change QueueChange
		err = rows.Scan(&data)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &change)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, change)
	}
	return tasks, rows.Err()
}

// QueueHealth samples admission, worker assignment and leases in one SQLite snapshot.
func (db *DB) QueueHealth(ctx context.Context, now time.Time) (domain.QueueHealth, error) {
	h := domain.QueueHealth{ObservedMS: now.UnixMilli()}
	err := db.sql.QueryRowContext(ctx, `WITH rows AS (
 SELECT t.*, u.admission='approved' AND (u.admin=1 OR EXISTS(
 SELECT 1 FROM effective_pool_grants g WHERE g.cluster=t.cluster AND g.pool=t.pool AND g.service=t.service AND g.principal IN(t.creator,'public')))
 AND EXISTS(SELECT 1 FROM pool_services s JOIN worker_cluster w ON s.worker=w.worker AND s.cluster=w.cluster
 WHERE s.cluster=t.cluster AND s.pool=t.pool AND s.service=t.service) AS allowed
 FROM pool_tasks t JOIN cluster_users u ON u.cluster=t.cluster AND u.user_id=t.creator
 WHERE t.service=? AND t.state IN ('running','queued') AND lower(trim(t.pool)) NOT IN ('personal','personal pool'))
 SELECT count(*) FILTER(WHERE state='queued'),count(*) FILTER(WHERE state='queued' AND allowed),
 count(*) FILTER(WHERE state='running' AND lease_until>?),count(*) FILTER(WHERE state='running' AND lease_until<=?),
 coalesce(min(created_ms) FILTER(WHERE state='queued' AND allowed),0),
 EXISTS(SELECT 1 FROM local_runtime WHERE key IN ('worker_update_drain_requested','worker_update_installing') AND value='1')
 FROM rows`, HashService, h.ObservedMS, h.ObservedMS).Scan(&h.Queued, &h.Eligible, &h.Running, &h.Expired, &h.OldestEligibleMS, &h.Draining)
	return h, err
}
