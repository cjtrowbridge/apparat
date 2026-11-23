// Records acknowledged workers on a personal app and idempotent receipts on the target.
// Only a verified exact retry can recover a lost first-claim acknowledgement.
package persistence

import (
	"context"
	"database/sql"
	"github.com/cjtrowbridge/apparat/internal/domain"

	"errors"
)

const claimReceiptSchema = `
CREATE TABLE claimed_workers(worker TEXT PRIMARY KEY,cluster TEXT NOT NULL,owner TEXT NOT NULL,host TEXT NOT NULL,address TEXT NOT NULL);
CREATE TABLE claim_receipts(worker TEXT PRIMARY KEY,digest TEXT NOT NULL);
`

func (db *DB) WorkerPin(ctx context.Context, host string) (string, error) {
	var pin string
	err := db.sql.QueryRowContext(ctx, "SELECT worker FROM claimed_workers WHERE host=? COLLATE NOCASE", host).Scan(&pin)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return pin, err
}
func (db *DB) SaveWorkerAccess(ctx context.Context, host, address, worker string, a Access) error {
	_, err := db.sql.ExecContext(ctx, `INSERT INTO claimed_workers VALUES(?,?,?,?,?) ON CONFLICT(worker) DO UPDATE SET address=excluded.address,host=excluded.host`, worker, a.Cluster, a.User, host, address)
	return err
}

func (db *DB) KnownWorkers(ctx context.Context, user string) ([]domain.ClusterDevice, error) {
	return db.WorkerSummaries(ctx, user, "")
}

func (db *DB) WorkerSummaries(ctx context.Context, user, device string) ([]domain.ClusterDevice, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT w.worker,w.host,w.address,w.cluster,coalesce(r.revision,0),
 coalesce(json_extract(r.body,'$.View.Governance.Head.Revision'),0),coalesce(json_extract(r.body,'$.View.Admin'),0),
 coalesce(json_array_length(r.body,'$.View.Pools'),0),coalesce(json_extract(r.body,'$.View.PollSeconds'),0)
 FROM claimed_workers w LEFT JOIN worker_replicas r ON r.worker=w.worker AND r.user=w.owner AND r.device=?
 WHERE w.owner=? ORDER BY host COLLATE NOCASE,w.worker`, device, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []domain.ClusterDevice
	for rows.Next() {
		var d domain.ClusterDevice
		err = rows.Scan(&d.Fingerprint, &d.Host, &d.Address, &d.Cluster, &d.ReplicaRevision, &d.PolicyRevision, &d.Admin, &d.PoolCount, &d.PollSeconds)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
