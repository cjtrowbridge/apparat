// Stores admitted model-step packets and write-once execution and result receipts.
// An interrupted dispatch remains owned. Reopening the database never requeues it.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const inferenceSchema = `
CREATE TABLE inference_roster(singleton INTEGER PRIMARY KEY CHECK(singleton=1),revision TEXT NOT NULL,body BLOB NOT NULL);
CREATE TABLE inference_jobs(id TEXT PRIMARY KEY,digest TEXT NOT NULL,body BLOB NOT NULL,state TEXT NOT NULL DEFAULT 'queued',claim BLOB,output BLOB,delivered INTEGER NOT NULL DEFAULT 0);
CREATE INDEX inference_pending ON inference_jobs(state,id);`

// InferenceCounts scopes task totals to the admitted user viewing the status page.
func (db *DB) InferenceCounts(ctx context.Context, user string, s *domain.InferenceStatus) error {
	err := db.sql.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE state='queued'),
	count(*) FILTER(WHERE state='owned'),count(*) FILTER(WHERE state='dispatching'),
	count(*) FILTER(WHERE state='finished') FROM inference_jobs WHERE json_extract(CAST(body AS TEXT),'$.User')=?`, user).Scan(&s.Queued, &s.Owned, &s.Dispatching, &s.Finished)
	if err != nil {
		return err
	}
	var oldest string
	err = db.sql.QueryRowContext(ctx, `SELECT count(*),coalesce(min(received_ms),0) FROM inference_intake WHERE user=? AND applied=0`, user).Scan(&s.Received, &s.OldestQueuedMS)
	if err != nil {
		return err
	}
	err = db.sql.QueryRowContext(ctx, `SELECT coalesce(min(id),'') FROM inference_jobs WHERE state='queued' AND json_extract(body,'$.User')=?`, user).Scan(&oldest)
	if oldest != "" {
		at, e := domain.ChatTime(oldest)
		if e == nil && (s.OldestQueuedMS == 0 || at.UnixMilli() < s.OldestQueuedMS) {
			s.OldestQueuedMS = at.UnixMilli()
		}
	}
	return err
}

func (db *DB) InferenceRoster(ctx context.Context) (r domain.InferenceRoster, err error) {
	var b []byte
	err = db.sql.QueryRowContext(ctx, "SELECT body FROM inference_membership WHERE singleton=1").Scan(&b)
	if err == nil {
		err = json.Unmarshal(b, &r)
	}
	return
}
func (db *DB) InstallInferenceRoster(ctx context.Context, r domain.InferenceRoster) error {
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "INSERT INTO inference_membership VALUES(1,?,?) ON CONFLICT(singleton) DO UPDATE SET body=excluded.body,revision=excluded.revision", r.Revision(), b)
	return err
}
func (db *DB) AdmitInference(ctx context.Context, j domain.InferenceJob) error {
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	result, err := db.sql.ExecContext(ctx, "INSERT INTO inference_jobs(id,digest,body) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET id=excluded.id WHERE digest=excluded.digest", j.Packet.Job, domain.InferenceDigest(j), b)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n != 1 {
		return errors.New("job or roster conflict")
	}
	return err
}

func scanInference(row interface{ Scan(...any) error }) (r domain.InferenceRecord, err error) {
	var body, claim, output []byte
	err = row.Scan(&body, &r.State, &claim, &output)
	if err == nil {
		err = json.Unmarshal(body, &r.Job)
	}
	if err == nil && len(claim) > 0 {
		err = json.Unmarshal(claim, &r.Claim)
	}
	if err == nil && len(output) > 0 {
		err = json.Unmarshal(output, &r.Output)
	}
	return
}
func (db *DB) InferenceRecord(ctx context.Context, id string) (domain.InferenceRecord, error) {
	return scanInference(db.sql.QueryRowContext(ctx, "SELECT body,state,claim,output FROM inference_jobs WHERE id=?", id))
}
func (db *DB) InferenceRecords(ctx context.Context, visit func(domain.InferenceRecord) error) error {
	rows, err := db.sql.QueryContext(ctx, "SELECT body,state,claim,output FROM inference_jobs")
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		r, err := scanInference(rows)
		if err != nil {
			return err
		}
		err = visit(r)
		if err != nil {
			return err
		}
	}
	return rows.Err()
}

// Seal is idempotent. A conflicting attempt cannot obtain a receipt after restart.
func (db *DB) SealInference(ctx context.Context, c domain.InferenceClaim) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	result, err := db.sql.ExecContext(ctx, "UPDATE inference_jobs SET claim=?,state=CASE WHEN state='queued' THEN 'owned' ELSE state END WHERE id=? AND digest=? AND (state='queued' OR claim=?)", b, c.Job, c.Digest, b)
	return inferenceChange(result, err)
}

// RecoverInterrupted records an explicit unknown outcome for a locally interrupted dispatch.
func (db *DB) RecoverInterrupted(ctx context.Context, self string) error {
	_, err := db.sql.ExecContext(ctx, `UPDATE inference_jobs SET state='finished',output=CAST('{"finish":"indeterminate"}' AS BLOB) WHERE state='dispatching' AND json_extract(claim,'$.Executor')=?`, self)
	return err
}
func (db *DB) StartInference(ctx context.Context, c domain.InferenceClaim) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	result, err := db.sql.ExecContext(ctx, "UPDATE inference_jobs SET state='dispatching' WHERE id=? AND state='owned' AND claim=? AND NOT EXISTS(SELECT 1 FROM local_runtime WHERE key='worker_update_installing' AND value='1')", c.Job, b)
	return inferenceChange(result, err)
}
func (db *DB) FinishInference(ctx context.Context, c domain.InferenceClaim, out domain.ChatOutput) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return err
	}
	result, err := db.sql.ExecContext(ctx, "UPDATE inference_jobs SET state='finished',output=? WHERE id=? AND claim=? AND state IN ('owned','dispatching','finished') AND (output IS NULL OR output=?)", raw, c.Job, b, raw)
	return inferenceChange(result, err)
}
func inferenceChange(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n != 1 {
		return errors.New("inactive or conflicting inference attempt")
	}
	return err
}

// InferenceAccess checks present user and pool policy before accepting or executing work.
func (db *DB) InferenceAccess(ctx context.Context, cluster, user, pool, worker string, owned bool) (bool, error) {
	var yes bool
	err := db.sql.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM cluster_users u WHERE u.cluster=? AND u.user_id=? AND u.admission='approved'
 AND (u.admin=1 OR EXISTS(SELECT 1 FROM effective_pool_grants g WHERE g.cluster=u.cluster AND g.pool=? AND g.service='text-generation' AND g.principal IN (u.user_id,'public')))

 AND (?='' OR NOT EXISTS(SELECT 1 FROM local_runtime WHERE (key='worker_update_installing' OR (key='worker_update_drain_requested' AND ?=0)) AND value='1')))`, cluster, user, pool, worker, owned).Scan(&yes)
	return yes, err
}

func (db *DB) RunnableInference(ctx context.Context, self string, ready bool, offset, limit int) ([]domain.InferenceRecord, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT body,state,claim,output FROM inference_jobs WHERE (? AND state='queued') OR
 (json_extract(claim,'$.Executor')=? AND ((? AND state='owned') OR (state='finished' AND delivered=0)))
 ORDER BY CASE state WHEN 'finished' THEN 0 WHEN 'owned' THEN 1 ELSE 2 END,
 CASE json_extract(body,'$.Pool') WHEN 'Chat Pool' THEN 0 WHEN 'Project Pool' THEN 1 WHEN 'Comrade Pool' THEN 2 ELSE 3 END,id LIMIT ? OFFSET ?`, ready, self, ready, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var out []domain.InferenceRecord
	for rows.Next() {
		r, e := scanInference(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (db *DB) InferenceDelivered(ctx context.Context, c domain.InferenceClaim) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "UPDATE inference_jobs SET delivered=1 WHERE id=? AND claim=? AND state='finished'", c.Job, b)
	return err
}

// Tasks returns a bounded user-scoped metadata page for the pool browser.
func (db *DB) InferenceTasks(ctx context.Context, user, pool, cursor string) ([]domain.PoolTaskRow, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT id,state FROM (
 SELECT id,state FROM inference_jobs WHERE json_extract(body,'$.User')=?1 AND json_extract(body,'$.Pool')=?2
 UNION ALL
 SELECT turn,'queued' FROM inference_intake r WHERE user=?1 AND ?2='Chat Pool' AND (applied=0 OR EXISTS(SELECT 1 FROM chat_work w WHERE w.user=r.user AND w.session=r.session AND w.turn=r.turn))
 AND NOT EXISTS(SELECT 1 FROM inference_jobs j WHERE json_extract(j.body,'$.User')=r.user AND json_extract(j.body,'$.Packet.Turn')=r.turn))
 WHERE (?3='' OR id<?3) ORDER BY id DESC LIMIT 64`, user, pool, cursor)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var out []domain.PoolTaskRow
	for rows.Next() {
		var t domain.PoolTaskRow
		err = rows.Scan(&t.ID, &t.State)
		if err != nil {
			return nil, err
		}
		at, _ := domain.ChatTime(t.ID)
		t.Changed = at.UnixMilli()
		out = append(out, t)
	}
	return out, rows.Err()
}
