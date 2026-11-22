// Records acknowledged workers on a personal app and idempotent receipts on the target.
// Only a verified exact retry can recover a lost first-claim acknowledgement.
package persistence

import (
	"context"
	"database/sql"

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
