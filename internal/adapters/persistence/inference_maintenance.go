// Persists voluntary worker withdrawal and its replay fence across updater restarts.
// A returning worker imports the surviving replicas' receipts before it can vote.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type WorkerWithdrawal struct{ Worker, Generation, Phase string }

func (db *DB) Withdrawal(ctx context.Context) (w WorkerWithdrawal, err error) {
	err = db.sql.QueryRowContext(ctx, "SELECT worker,generation,phase FROM inference_withdrawals WHERE phase!='released'").Scan(&w.Worker, &w.Generation, &w.Phase)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return
}
func (db *DB) SetWithdrawal(ctx context.Context, w WorkerWithdrawal) error {
	result, err := db.sql.ExecContext(ctx, `INSERT INTO inference_withdrawals VALUES(?,?,?) ON CONFLICT(worker) DO UPDATE SET generation=excluded.generation,phase=excluded.phase
 WHERE (excluded.generation>generation AND phase='released') OR excluded.generation=generation AND (phase=excluded.phase OR phase!='released' AND (excluded.phase IN ('closing','releasing','released') OR phase='prepared' AND excluded.phase='away'))`, w.Worker, w.Generation, w.Phase)
	return inferenceChange(result, err)
}
func (db *DB) OpenOwnership(ctx context.Context) (bool, error) {
	var busy bool
	err := db.sql.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM inference_jobs WHERE state IN ('owned','dispatching') OR state='finished' AND delivered=0 AND json_extract(claim,'$.Executor')=(SELECT worker FROM worker_cluster))`).Scan(&busy)
	return busy, err
}
func (db *DB) NextInferenceRecord(ctx context.Context, after string) (r domain.InferenceRecord, err error) {
	return scanInference(db.sql.QueryRowContext(ctx, "SELECT body,state,claim,output FROM inference_jobs WHERE id>? ORDER BY id LIMIT 1", after))
}
func (db *DB) RestoreInferenceRecord(ctx context.Context, r domain.InferenceRecord) error {
	if !r.Job.Valid() || r.State != "queued" && r.State != "finished" {
		return errors.New("invalid recovery record")
	}
	err := db.AdmitInference(ctx, r.Job)
	if err != nil || r.State == "queued" {
		return err
	}
	if r.Output == nil || r.Claim.Digest != domain.InferenceDigest(r.Job) {
		return errors.New("invalid completed receipt")
	}
	claim, err := json.Marshal(r.Claim)
	if err != nil {
		return err
	}
	output, err := json.Marshal(r.Output)
	if err != nil {
		return err
	}
	result, err := db.sql.ExecContext(ctx, `UPDATE inference_jobs SET state='finished',claim=?,output=?,delivered=1 WHERE id=? AND (state='queued' OR state='finished' AND claim=? AND output=?)`, claim, output, r.Job.Packet.Job, claim, output)
	return inferenceChange(result, err)
}
