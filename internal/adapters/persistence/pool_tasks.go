// Owns a worker's SQLite queue and fenced lease transitions for its local hash-text service.
// Ordered per-pool event rows are written by SQLite triggers in the mutation transaction.
package persistence

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

const queueSchema = `
CREATE TABLE pool_tasks (
 cluster TEXT NOT NULL, pool TEXT NOT NULL, id TEXT NOT NULL, creator TEXT NOT NULL,
 service TEXT NOT NULL, prompt TEXT NOT NULL, created_ms INTEGER NOT NULL, updated_ms INTEGER NOT NULL,
 state TEXT NOT NULL CHECK(state IN ('queued','running','succeeded','cancelled')), fence INTEGER NOT NULL DEFAULT 0,
 lease_worker TEXT NOT NULL DEFAULT '',lease_until INTEGER NOT NULL DEFAULT 0,
 PRIMARY KEY(cluster,pool,id),FOREIGN KEY(cluster,pool) REFERENCES cluster_pools(cluster,pool));
CREATE INDEX pool_tasks_eligible ON pool_tasks(cluster,pool,state,created_ms);
CREATE TABLE pool_task_attempts(cluster TEXT NOT NULL,pool TEXT NOT NULL,task TEXT NOT NULL,fence INTEGER NOT NULL,
 worker TEXT NOT NULL,started_ms INTEGER NOT NULL,expires_ms INTEGER NOT NULL,outcome TEXT NOT NULL,
 PRIMARY KEY(cluster,pool,task,fence),FOREIGN KEY(cluster,pool,task) REFERENCES pool_tasks(cluster,pool,id));
CREATE TABLE pool_task_artifacts(cluster TEXT NOT NULL,pool TEXT NOT NULL,task TEXT NOT NULL,slot TEXT NOT NULL,
 digest TEXT NOT NULL, PRIMARY KEY(cluster,pool,task,slot),FOREIGN KEY(cluster,pool,task) REFERENCES pool_tasks(cluster,pool,id));
CREATE TABLE pool_queue_events(cluster TEXT NOT NULL,pool TEXT NOT NULL,revision INTEGER NOT NULL,task TEXT NOT NULL,
 creator TEXT NOT NULL,state TEXT NOT NULL,fence INTEGER NOT NULL,changed_ms INTEGER NOT NULL,
 PRIMARY KEY(cluster,pool,revision),FOREIGN KEY(cluster,pool,task) REFERENCES pool_tasks(cluster,pool,id));
CREATE INDEX pool_queue_user_events ON pool_queue_events(cluster,pool,creator,revision);
CREATE TRIGGER pool_queue_created AFTER INSERT ON pool_tasks BEGIN
 INSERT INTO pool_queue_events SELECT NEW.cluster,NEW.pool,COALESCE(MAX(revision),0)+1,NEW.id,NEW.creator,NEW.state,NEW.fence,NEW.updated_ms
 FROM pool_queue_events WHERE cluster=NEW.cluster AND pool=NEW.pool;
END;
CREATE TRIGGER pool_queue_changed AFTER UPDATE ON pool_tasks BEGIN
 INSERT INTO pool_queue_events SELECT NEW.cluster,NEW.pool,COALESCE(MAX(revision),0)+1,NEW.id,NEW.creator,NEW.state,NEW.fence,NEW.updated_ms
 FROM pool_queue_events WHERE cluster=NEW.cluster AND pool=NEW.pool;
END;
`

const HashService = "hash-text"

type Task struct {
	Cluster, Pool, ID, Creator, Service, Prompt, State string
	Fence, Created, Updated                            int64
}

func (db *DB) SubmitHash(ctx context.Context, a Access, pool, id, prompt string) error {
	if len(prompt) > 4096 || len(id) == 0 || len(id) > 128 {
		return errors.New("invalid hash task")
	}
	db.queueMu.Lock()
	defer db.queueMu.Unlock()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = queueAccess(ctx, tx, a, pool)
	if err != nil {
		return err
	}
	var matches bool
	err = tx.QueryRowContext(ctx, "SELECT creator=? AND prompt=? AND service=? FROM pool_tasks WHERE cluster=? AND pool=? AND id=?", a.User, prompt, HashService, a.Cluster, pool, id).Scan(&matches)
	if err == nil {
		if matches {
			return nil
		}
		return errors.New("duplicate task ID conflicts")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var capacity bool
	err = tx.QueryRowContext(ctx, `SELECT count(*)<coalesce((SELECT max_active FROM pool_limits WHERE cluster=? AND pool=?),64)
 FROM pool_tasks WHERE cluster=? AND pool=? AND creator=? AND state IN ('queued','running')`, a.Cluster, pool, a.Cluster, pool, a.User).Scan(&capacity)
	if err != nil || !capacity {
		return errors.New("pool task quota reached")
	}
	var draining bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM local_runtime WHERE key IN ('worker_update_drain_requested','worker_update_installing') AND value='1')").Scan(&draining)
	if err != nil || draining {
		return errors.New("worker is draining for an update")
	}
	now := time.Now().UnixMilli()
	_, err = tx.ExecContext(ctx, `INSERT INTO pool_tasks(cluster,pool,id,creator,service,prompt,created_ms,updated_ms,state)
	VALUES(?,?,?,?,?,?,?,?,'queued') ON CONFLICT DO NOTHING`, a.Cluster, pool, id, a.User, HashService, prompt, now, now)
	if err != nil {
		return err
	}
	return db.commitQueue(ctx, tx, queueView{a.Cluster, pool, a.User})
}

func (db *DB) LeaseHash(ctx context.Context, cluster, pool, worker string, now time.Time) (Task, error) {
	var t Task
	db.queueMu.Lock()
	defer db.queueMu.Unlock()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return t, err
	}
	defer tx.Rollback()
	at := now.UnixMilli()
	until := now.Add(time.Minute).UnixMilli()
	err = tx.QueryRowContext(ctx, `UPDATE pool_tasks SET fence=fence+1,state='running',lease_worker=?,lease_until=?,updated_ms=?
	WHERE rowid=(SELECT t.rowid FROM pool_tasks t JOIN cluster_users u ON u.cluster=t.cluster AND u.user_id=t.creator
	WHERE t.cluster=? AND t.pool=? AND t.service=? AND (t.state='queued' OR (t.state='running' AND t.lease_until<=?))
	AND NOT EXISTS(SELECT 1 FROM local_runtime WHERE key IN ('worker_update_drain_requested','worker_update_installing') AND value='1')
 AND u.admission='approved' AND (u.admin=1 OR EXISTS(SELECT 1 FROM effective_pool_grants WHERE cluster=t.cluster AND pool=t.pool AND service=t.service AND principal IN(t.creator,'public')))
	AND EXISTS(SELECT 1 FROM pool_services WHERE cluster=t.cluster AND pool=t.pool AND worker=? AND service=t.service)
	ORDER BY t.created_ms,t.id LIMIT 1) RETURNING cluster,pool,id,creator,service,prompt,state,fence,created_ms,updated_ms`,
		worker, until, at, cluster, pool, HashService, at, worker).Scan(&t.Cluster, &t.Pool, &t.ID, &t.Creator, &t.Service, &t.Prompt, &t.State, &t.Fence, &t.Created, &t.Updated)
	if err != nil {
		return t, err
	}
	_, err = tx.ExecContext(ctx, "UPDATE pool_task_attempts SET outcome='expired' WHERE cluster=? AND pool=? AND task=? AND outcome='active'", cluster, pool, t.ID)
	if err == nil {
		_, err = tx.ExecContext(ctx, "INSERT INTO pool_task_attempts VALUES(?,?,?,?,?,?,?,'active')", cluster, pool, t.ID, t.Fence, worker, at, until)
	}
	if err != nil {
		return Task{}, err
	}
	return t, db.commitQueue(ctx, tx, queueView{t.Cluster, t.Pool, t.Creator})
}

func (db *DB) CompleteHash(ctx context.Context, t Task, worker string, digest string, now time.Time) error {
	return db.completeHash(ctx, t, worker, digest, now, nil)
}
func (db *DB) completeHash(ctx context.Context, t Task, worker, digest string, now time.Time, receipt *ResultReceipt) error {
	db.queueMu.Lock()
	defer db.queueMu.Unlock()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var prompt, creator string
	err = tx.QueryRowContext(ctx, "SELECT prompt,creator FROM pool_tasks WHERE cluster=? AND pool=? AND id=? AND fence=? AND state='running' AND lease_worker=? AND lease_until>?", t.Cluster, t.Pool, t.ID, t.Fence, worker, now.UnixMilli()).Scan(&prompt, &creator)
	if errors.Is(err, sql.ErrNoRows) {
		var exact bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pool_tasks t JOIN pool_task_artifacts r ON t.cluster=r.cluster AND t.pool=r.pool AND t.id=r.task
  WHERE t.cluster=? AND t.pool=? AND t.id=? AND t.creator=? AND t.lease_worker=? AND t.fence=? AND t.state='succeeded' AND r.slot='result' AND r.digest=?)`, t.Cluster, t.Pool, t.ID, t.Creator, worker, t.Fence, digest).Scan(&exact)
		if err == nil && exact {
			return nil
		}
		return errors.New("inactive worker lease")
	}
	if err != nil || creator != t.Creator {
		return errors.New("inactive worker lease")
	}
	sum := sha256.Sum256([]byte(prompt))
	artifact := sha256.Sum256([]byte(hex.EncodeToString(sum[:]) + "\n"))
	if hex.EncodeToString(artifact[:]) != digest {
		return errors.New("hash does not match input")
	}
	var raw []byte
	if receipt != nil {
		raw, err = json.Marshal(receipt)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO pool_task_artifacts(cluster,pool,task,slot,digest,receipt) VALUES(?,?,?,'result',?,?)", t.Cluster, t.Pool, t.ID, digest, raw)
	if err == nil {
		_, err = tx.ExecContext(ctx, "UPDATE pool_task_attempts SET outcome='succeeded' WHERE cluster=? AND pool=? AND task=? AND fence=?", t.Cluster, t.Pool, t.ID, t.Fence)
	}
	if err == nil {
		_, err = tx.ExecContext(ctx, "UPDATE pool_tasks SET state='succeeded',updated_ms=?,lease_until=0 WHERE cluster=? AND pool=? AND id=?", now.UnixMilli(), t.Cluster, t.Pool, t.ID)
	}
	if err != nil {
		return err
	}
	return db.commitQueue(ctx, tx, queueView{t.Cluster, t.Pool, t.Creator})
}

func (db *DB) CancelTask(ctx context.Context, a Access, pool, id string) error {
	db.queueMu.Lock()
	defer db.queueMu.Unlock()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = queueAccess(ctx, tx, a, pool)
	if err != nil {
		return err
	}
	var fence int64
	err = tx.QueryRowContext(ctx, "UPDATE pool_tasks SET state='cancelled',lease_until=0,updated_ms=? WHERE cluster=? AND pool=? AND id=? AND creator=? AND state IN ('running','queued') RETURNING fence", time.Now().UnixMilli(), a.Cluster, pool, id, a.User).Scan(&fence)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("task unavailable for cancellation")
	}
	if err == nil {
		_, err = tx.ExecContext(ctx, "UPDATE pool_task_attempts SET outcome='cancelled' WHERE cluster=? AND pool=? AND task=? AND fence=? AND outcome='active'", a.Cluster, pool, id, fence)
	}
	if err != nil {
		return err
	}
	return db.commitQueue(ctx, tx, queueView{a.Cluster, pool, a.User})
}
