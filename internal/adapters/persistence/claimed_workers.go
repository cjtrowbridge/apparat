// Records acknowledged workers on a personal app and idempotent receipts on the target.
// Only a verified exact retry can recover a lost first-claim acknowledgement.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"

	"errors"
)

const claimReceiptSchema = `
CREATE TABLE claimed_workers(worker TEXT PRIMARY KEY,cluster TEXT NOT NULL,owner TEXT NOT NULL,host TEXT NOT NULL,address TEXT NOT NULL);
CREATE TABLE claim_receipts(worker TEXT PRIMARY KEY,digest TEXT NOT NULL);
`

func (db *DB) WorkerPin(ctx context.Context, host, address string) (string, error) {
	var pin string
	var count int
	err := db.sql.QueryRowContext(ctx, "SELECT coalesce(min(worker),''),count(*) FROM claimed_workers WHERE host=? COLLATE NOCASE OR address=?", host, address).Scan(&pin, &count)
	if count > 1 {
		return "", errors.New("ambiguous saved worker identity")
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return pin, err
}
func (db *DB) SaveWorkerAccess(ctx context.Context, host, address, worker string, a Access) error {
	result, err := db.sql.ExecContext(ctx, `INSERT INTO claimed_workers(worker,cluster,owner,host,address) SELECT ?,?,?,?,?
	WHERE NOT EXISTS(SELECT 1 FROM claimed_workers WHERE worker<>? AND (host=? COLLATE NOCASE OR address=?))
	ON CONFLICT(worker) DO UPDATE SET address=excluded.address,host=excluded.host
	WHERE owner=excluded.owner AND cluster=excluded.cluster`, worker, a.Cluster, a.User, host, address, worker, host, address)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return errors.New("conflicting saved worker registration")
	}
	return nil
}

func (db *DB) KnownWorkers(ctx context.Context, user string) ([]domain.ClusterDevice, error) {
	return db.WorkerSummaries(ctx, user, "")
}

func (db *DB) WorkerSummaries(ctx context.Context, user, device string) ([]domain.ClusterDevice, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT w.worker,w.host,w.address,w.cluster,coalesce(r.revision,0),
 coalesce(json_extract(r.body,'$.View.Governance.Head.Revision'),0),coalesce(json_extract(r.body,'$.View.Admin'),0),
 coalesce(json_array_length(r.body,'$.View.Pools'),0),coalesce(json_extract(r.body,'$.View.PollSeconds'),0),coalesce(json_extract(r.body,'$.View.Devices'),'[]'),coalesce(json_extract(r.body,'$.View.Governance.Pools'),json_extract(r.body,'$.View.Pools'),'[]'),w.hardware,w.inventory,w.heard
 FROM claimed_workers w LEFT JOIN worker_replicas r ON r.worker=w.worker AND r.user=w.owner AND r.device=?
 WHERE w.owner=? ORDER BY host COLLATE NOCASE,w.worker`, device, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []domain.ClusterDevice
	for rows.Next() {
		var d domain.ClusterDevice
		var roster, pools string
		var hardware, inventory []byte
		var heard int64
		err = rows.Scan(&d.Fingerprint, &d.Host, &d.Address, &d.Cluster, &d.ReplicaRevision, &d.PolicyRevision, &d.Admin, &d.PoolCount, &d.PollSeconds, &roster, &pools, &hardware, &inventory, &heard)
		if err == nil {
			err = json.Unmarshal([]byte(roster), &d.Roster)
		}
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(pools), &d.PoolNames)
		if err != nil {
			return nil, err
		}
		if len(hardware) > 0 {
			err = json.Unmarshal(hardware, &d.Hardware)
			if err != nil {
				return nil, err
			}
		}
		if len(inventory) > 0 {
			var v domain.WorkerInventory
			err = json.Unmarshal(inventory, &v)
			if err != nil {
				return nil, err
			}
			v.LastContact = heard
			v.Retrying = true
			d.Inventory = &v
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (db *DB) UpdateWorkerAddress(ctx context.Context, pin, user, address string) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE claimed_workers SET address=? WHERE worker=? AND owner=?", address, pin, user)
	return err
}

// ReplaceWorkerPin changes one selected registration. Historical receipts remain keyed to the old signer.
func (db *DB) ReplaceWorkerPin(ctx context.Context, old, newPin, address string, a Access) error {
	if !a.Admin || a.Admission != "approved" {
		return errors.New("administrator required")
	}
	result, err := db.sql.ExecContext(ctx, "UPDATE claimed_workers SET worker=?,address=?,hardware=NULL,inventory=NULL,heard=0 WHERE worker=? AND owner=? AND cluster=?", newPin, address, old, a.User, a.Cluster)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return errors.New("worker registration changed during review")
	}
	return nil
}

func (db *DB) RememberInventory(ctx context.Context, user, worker string, v *domain.WorkerInventory) error {
	if v == nil {
		_, err := db.sql.ExecContext(ctx, "UPDATE claimed_workers SET inventory=NULL,heard=0 WHERE owner=? AND worker=?", user, worker)
		return err
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "UPDATE claimed_workers SET inventory=?,heard=? WHERE owner=? AND worker=? AND heard<=?", raw, v.LastContact, user, worker, v.LastContact)
	return err
}
