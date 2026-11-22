// Admits the first cluster admin on possession of a locally issued, expiring worker invitation.
// An invitation is consumed in the same SQLite transaction as the owner record.
package persistence

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const claimSchema = `
CREATE TABLE worker_claim_invitation (
 singleton INTEGER PRIMARY KEY CHECK(singleton=1), worker TEXT NOT NULL,
 digest TEXT NOT NULL, expires_at INTEGER NOT NULL
);
CREATE TABLE worker_cluster (
 singleton INTEGER PRIMARY KEY CHECK(singleton=1), cluster TEXT NOT NULL UNIQUE,
 worker TEXT NOT NULL, owner_user TEXT NOT NULL, authority_device TEXT NOT NULL,
 poll_seconds INTEGER NOT NULL DEFAULT 300 CHECK(poll_seconds>0),
 admission_open INTEGER NOT NULL DEFAULT 1 CHECK(admission_open IN (0,1))
);
CREATE TABLE cluster_users (
 cluster TEXT NOT NULL REFERENCES worker_cluster(cluster),
 user_id TEXT NOT NULL, public_profile BLOB NOT NULL,
 admission TEXT NOT NULL CHECK(admission IN ('approved','blocked','pending')),
 admin INTEGER NOT NULL CHECK(admin IN (0,1) AND (admin=0 OR admission='approved')),
 PRIMARY KEY(cluster,user_id)
);
CREATE TABLE cluster_devices_authorized (
 cluster TEXT NOT NULL, user_id TEXT NOT NULL, device_id TEXT NOT NULL, public_binding BLOB NOT NULL,
 PRIMARY KEY(cluster,user_id,device_id),
 FOREIGN KEY(cluster,user_id) REFERENCES cluster_users(cluster,user_id)
);
`

func (db *DB) IssueWorkerInvitation(ctx context.Context, public []byte) (string, error) {
	if len(public) != 32 {
		return "", errors.New("worker needs a device key")
	}
	var secret [32]byte
	_, err := rand.Read(secret[:])
	if err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(secret[:])
	digest := identity.ClaimDigest(code)
	result, err := db.sql.ExecContext(ctx, `INSERT INTO worker_claim_invitation(singleton,worker,digest,expires_at)
	SELECT 1,?,?,? WHERE NOT EXISTS(SELECT 1 FROM worker_cluster)
	ON CONFLICT(singleton) DO UPDATE SET worker=excluded.worker,digest=excluded.digest,expires_at=excluded.expires_at`,
		identity.Fingerprint(public), digest, time.Now().Add(10*time.Minute).Unix())
	if err != nil {
		return "", err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return "", errors.New("worker already claimed")
	}
	return code, nil
}

func (db *DB) ClaimWorker(ctx context.Context, c identity.Claim, workerPublic []byte) error {
	if !c.Valid() || c.Worker != identity.Fingerprint(workerPublic) {
		return errors.New("invalid claim proof")
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM worker_claim_invitation WHERE singleton=1 AND worker=? AND digest=? AND expires_at>?`, c.Worker, identity.ClaimDigest(c.Invitation), time.Now().Unix())
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return errors.New("invitation is invalid, expired or already used")
	}
	err = establishOwner(ctx, tx, c)
	if err == nil {
		_, err = tx.ExecContext(ctx, "INSERT INTO claim_receipts(worker,digest) VALUES(?,?)", c.Worker, claimReceipt(c))
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

func establishOwner(ctx context.Context, tx *sql.Tx, c identity.Claim) error {
	user := identity.Fingerprint(c.Device.Binding.User)
	device := identity.Fingerprint(c.Device.Binding.Device)
	profile, _ := json.Marshal(c.Profile)
	binding, _ := json.Marshal(c.Device)
	_, err := tx.ExecContext(ctx, "INSERT INTO worker_cluster(singleton,cluster,worker,owner_user,authority_device) VALUES(1,?,?,?,?)", c.Cluster, c.Worker, user, device)
	if err == nil {
		_, err = tx.ExecContext(ctx, "INSERT INTO cluster_users(cluster,user_id,public_profile,admission,admin) VALUES(?,?,?,'approved',1)", c.Cluster, user, profile)
	}
	if err == nil {
		_, err = tx.ExecContext(ctx, "INSERT INTO cluster_devices_authorized(cluster,user_id,device_id,public_binding) VALUES(?,?,?,?)", c.Cluster, user, device, binding)
	}
	return err
}
