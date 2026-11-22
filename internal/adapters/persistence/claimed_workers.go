// Records acknowledged workers on a personal app and idempotent receipts on the target.
// Only a verified exact retry can recover a lost first-claim acknowledgement.
package persistence

import (
	"context"
	"encoding/json"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const claimReceiptSchema = `
CREATE TABLE claimed_workers(worker TEXT PRIMARY KEY,cluster TEXT NOT NULL,owner TEXT NOT NULL,host TEXT NOT NULL,address TEXT NOT NULL);
CREATE TABLE claim_receipts(worker TEXT PRIMARY KEY,digest TEXT NOT NULL);
`

func claimReceipt(c identity.Claim) string {
	data, _ := json.Marshal(c)
	return identity.ClaimDigest(string(data))
}

func (db *DB) ConfirmedClaim(ctx context.Context, c identity.Claim) bool {
	if !c.Valid() {
		return false
	}
	var exists int
	err := db.sql.QueryRowContext(ctx, "SELECT 1 FROM claim_receipts WHERE worker=? AND digest=?", c.Worker, claimReceipt(c)).Scan(&exists)
	return err == nil
}

func (db *DB) RememberClaim(ctx context.Context, c identity.Claim, host, address string) error {
	_, err := db.sql.ExecContext(ctx, `INSERT INTO claimed_workers(worker,cluster,owner,host,address) VALUES(?,?,?,?,?)
	ON CONFLICT(worker) DO UPDATE SET host=excluded.host,address=excluded.address WHERE owner=excluded.owner AND cluster=excluded.cluster`, c.Worker, c.Cluster, identity.Fingerprint(c.Device.Binding.User), host, address)
	return err
}
