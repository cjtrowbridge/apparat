// Stores post-claim registrations and checks current user and device rights per operation.
// A new device belonging to an approved user must prove possession of both keys.
package persistence

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type Access struct {
	Cluster   string
	User      string
	Device    string
	Admission string
	Admin     bool
}

const admissionSchema = `
CREATE TABLE revoked_devices(cluster TEXT NOT NULL REFERENCES worker_cluster(cluster),device TEXT NOT NULL,PRIMARY KEY(cluster,device));
`

func (db *DB) Login(ctx context.Context, l identity.Login) (Access, error) {
	var a Access
	if !l.Valid() {
		return a, errors.New("invalid login proof")
	}
	db.policyMu.Lock()
	defer db.policyMu.Unlock()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return a, err
	}
	defer tx.Rollback()
	a.User = identity.Fingerprint(l.Device.Binding.User)
	a.Device = identity.Fingerprint(l.Device.Binding.Device)
	cluster := identity.ClaimDigest("apparat/personal-cluster/v1/" + a.User)
	result, err := tx.ExecContext(ctx, `INSERT INTO worker_cluster(singleton,cluster,worker,owner_user,authority_device)
		SELECT 1,?,?,?,? WHERE EXISTS(SELECT 1 FROM local_runtime WHERE key='confirmed_worker_owner' AND value=?) ON CONFLICT(singleton) DO NOTHING`, cluster, l.Worker, a.User, a.Device, l.Worker+"/"+a.User)
	if err != nil {
		return a, err
	}
	first, err := result.RowsAffected()
	if err != nil {
		return a, err
	}
	if first == 1 {
		for _, pool := range domain.DefaultPools() {
			_, err = tx.ExecContext(ctx, "INSERT INTO cluster_pools(cluster,pool) VALUES(?,?)", cluster, pool.Name)
			if err != nil {
				return a, err
			}
		}
		profile, _ := json.Marshal(l.Profile)
		_, err = tx.ExecContext(ctx, "INSERT INTO cluster_users VALUES(?,?,?,'approved',1)", cluster, a.User, profile)
		if err != nil {
			return a, err
		}
	}
	var open bool
	err = tx.QueryRowContext(ctx, "SELECT cluster,admission_open FROM worker_cluster WHERE worker=?", l.Worker).Scan(&a.Cluster, &open)
	if err != nil {
		return a, err
	}
	var revoked bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM revoked_devices WHERE cluster=? AND device=?)", a.Cluster, a.Device).Scan(&revoked)
	if err != nil || revoked {
		return a, errors.New("device unavailable")
	}
	err = tx.QueryRowContext(ctx, "SELECT admission,admin FROM cluster_users WHERE cluster=? AND user_id=?", a.Cluster, a.User).Scan(&a.Admission, &a.Admin)
	if errors.Is(err, sql.ErrNoRows) && open {
		profile, _ := json.Marshal(l.Profile)
		a.Admission = "pending"
		_, err = tx.ExecContext(ctx, "INSERT INTO cluster_users(cluster,user_id,public_profile,admission,admin) VALUES(?,?,?,'pending',0)", a.Cluster, a.User, profile)
	}
	if err != nil || a.Admission == "blocked" {
		return Access{}, errors.New("registration unavailable")
	}
	var previous []byte
	err = tx.QueryRowContext(ctx, "SELECT public_profile FROM cluster_users WHERE cluster=? AND user_id=?", a.Cluster, a.User).Scan(&previous)
	if err != nil {
		return Access{}, err
	}
	var old identity.SignedProfile
	err = json.Unmarshal(previous, &old)
	if err != nil || !old.Valid() {
		return Access{}, errors.New("invalid stored user profile")
	}
	current, _ := json.Marshal(l.Profile)
	if old.Profile.Revision == l.Profile.Profile.Revision && !bytes.Equal(previous, current) {
		return Access{}, errors.New("conflicting signed profile")
	}
	if old.Profile.Revision < l.Profile.Profile.Revision {
		_, err = tx.ExecContext(ctx, "UPDATE cluster_users SET public_profile=? WHERE cluster=? AND user_id=?", current, a.Cluster, a.User)
		if err != nil {
			return Access{}, err
		}
	}
	binding, _ := json.Marshal(l.Device)
	_, err = tx.ExecContext(ctx, "INSERT INTO cluster_devices_authorized(cluster,user_id,device_id,public_binding,hostname,platform) VALUES(?,?,?,?,?,?) ON CONFLICT(cluster,user_id,device_id) DO UPDATE SET hostname=excluded.hostname,platform=CASE WHEN excluded.platform='' THEN platform ELSE excluded.platform END", a.Cluster, a.User, a.Device, binding, l.Hostname, l.Platform)
	if err != nil {
		return Access{}, err
	}
	return a, db.commitView(ctx, tx)
}

func (db *DB) CurrentAccess(ctx context.Context, cluster, user, device string) (Access, error) {
	a := Access{Cluster: cluster, User: user, Device: device}
	err := db.sql.QueryRowContext(ctx, `SELECT u.admission,u.admin FROM cluster_users u JOIN cluster_devices_authorized d USING(cluster,user_id)
	WHERE u.cluster=? AND u.user_id=? AND d.device_id=? AND u.admission='approved' AND NOT EXISTS
	(SELECT 1 FROM revoked_devices WHERE cluster=u.cluster AND device=d.device_id)`, cluster, user, device).Scan(&a.Admission, &a.Admin)
	return a, err
}

// ConfirmOwner is a host-local operation. An established worker can only confirm its original owner.
func (db *DB) ConfirmOwner(ctx context.Context, worker, user string) error {
	return db.confirmOwner(ctx, worker, user, false)
}

// ReserveOwner keeps a previous installation reservation even before the first login.
func (db *DB) ReserveOwner(ctx context.Context, worker, user string) error {
	return db.confirmOwner(ctx, worker, user, true)
}

func (db *DB) confirmOwner(ctx context.Context, worker, user string, reserve bool) error {
	for _, pin := range []string{worker, user} {
		raw, err := base64.RawURLEncoding.Strict().DecodeString(pin)
		if err != nil || len(raw) != 32 || base64.RawURLEncoding.EncodeToString(raw) != pin {
			return errors.New("invalid identity fingerprint")
		}
	}
	result, err := db.sql.ExecContext(ctx, `INSERT INTO local_runtime(key,value,updated_at_ms)
	SELECT 'confirmed_worker_owner',?,0 WHERE NOT EXISTS
	(SELECT 1 FROM worker_cluster WHERE worker<>? OR owner_user<>?)
	ON CONFLICT(key) DO UPDATE SET value=excluded.value WHERE NOT ? OR local_runtime.value=excluded.value`, worker+"/"+user, worker, user, reserve)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil || count != 1 {
		return errors.New("worker or owner differs from existing claim")
	}
	return nil
}

func (db *DB) OwnerConfirmed(ctx context.Context, worker string) bool {
	var confirmed bool
	err := db.sql.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM worker_cluster w JOIN local_runtime r
	ON r.key='confirmed_worker_owner' AND r.value=w.worker||'/'||w.owner_user WHERE w.worker=?)`, worker).Scan(&confirmed)
	return err == nil && confirmed
}
