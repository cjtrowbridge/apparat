// Maintains a revision for exported governance views and stores verified app-side replicas.
// A cache belongs to one worker, user and device; the worker still authorizes every operation.
package persistence

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const replicaSchema = replicaPolicySchema + groupSchema + `
ALTER TABLE pool_task_artifacts ADD COLUMN receipt BLOB;
CREATE TABLE pool_limits(cluster TEXT NOT NULL,pool TEXT NOT NULL,max_active INTEGER NOT NULL CHECK(max_active BETWEEN 1 AND 4096),PRIMARY KEY(cluster,pool),FOREIGN KEY(cluster,pool) REFERENCES cluster_pools(cluster,pool));
ALTER TABLE worker_cluster ADD COLUMN authority_epoch INTEGER NOT NULL DEFAULT 0;
UPDATE worker_cluster SET authority_epoch=1 WHERE authority_device=worker;
CREATE TRIGGER next_authority_epoch AFTER UPDATE OF authority_device ON worker_cluster WHEN OLD.authority_device!=NEW.authority_device BEGIN UPDATE worker_cluster SET authority_epoch=OLD.authority_epoch+1 WHERE cluster=NEW.cluster; END;
CREATE TABLE view_revision(cluster TEXT PRIMARY KEY,revision INTEGER NOT NULL);
INSERT INTO view_revision SELECT cluster,1 FROM worker_cluster;
CREATE TRIGGER revision_worker AFTER INSERT ON worker_cluster BEGIN INSERT INTO view_revision VALUES(NEW.cluster,1); END;
CREATE TRIGGER revision_change AFTER INSERT ON policy_events BEGIN UPDATE view_revision SET revision=revision+1 WHERE cluster=NEW.cluster; END;
CREATE TRIGGER revision_member AFTER INSERT ON cluster_users BEGIN UPDATE view_revision SET revision=revision+1 WHERE cluster=NEW.cluster; END;
CREATE TRIGGER revision_profile AFTER UPDATE OF public_profile ON cluster_users WHEN OLD.public_profile!=NEW.public_profile BEGIN UPDATE view_revision SET revision=revision+1 WHERE cluster=NEW.cluster; END;
CREATE TABLE worker_replicas(worker TEXT NOT NULL,user TEXT NOT NULL,device TEXT NOT NULL,revision INTEGER NOT NULL,body BLOB NOT NULL,PRIMARY KEY(worker,user,device));
CREATE TABLE replica_queue(worker TEXT NOT NULL,user TEXT NOT NULL,device TEXT NOT NULL,pool TEXT NOT NULL,revision INTEGER NOT NULL,body BLOB NOT NULL,PRIMARY KEY(worker,user,device,pool,revision));
`

type SyncCursor struct {
	Revision uint64
	Queues   map[string]int64
}
type Replica struct {
	Domain, Worker, Cluster, User, Device string
	Revision                              uint64
	Admin                                 bool
	PollSeconds                           int
	Pools                                 []string
	Devices                               []domain.PoolDevice `json:",omitempty"`
	Governance                            *GovernanceView     `json:",omitempty"`
}
type SignedReplica struct {
	View      Replica
	Key       ed25519.PublicKey
	Signature []byte
}

func (r SignedReplica) Valid(w string, a Access) bool {
	v := r.View
	if v.Domain != "apparat/worker-replica/v1" || v.Worker != w || v.Cluster != a.Cluster || v.User != a.User || v.Device != a.Device || len(r.Key) != ed25519.PublicKeySize || identity.Fingerprint(r.Key) != w || v.Revision == 0 {
		return false
	}
	body, err := json.Marshal(v)
	return err == nil && ed25519.Verify(r.Key, body, r.Signature)
}

func (db *DB) refreshViewRevision(ctx context.Context) error {
	var rev uint64
	err := db.sql.QueryRowContext(ctx, "SELECT coalesce(max(revision),0) FROM view_revision").Scan(&rev)
	if err == nil {
		db.viewRevision = rev
	}
	return err
}

func (db *DB) ReadReplica(ctx context.Context, a Access, w string, after uint64, directory bool) (*Replica, error) {
	db.policyMu.RLock()
	defer db.policyMu.RUnlock()
	now, err := db.CurrentAccess(ctx, a.Cluster, a.User, a.Device)
	if err != nil {
		return nil, err
	}
	if after == db.viewRevision && after != 0 {
		return nil, nil
	}
	v := Replica{Domain: "apparat/worker-replica/v1", Worker: w, Cluster: a.Cluster, User: a.User, Device: a.Device, Revision: db.viewRevision, Admin: now.Admin}
	err = db.sql.QueryRowContext(ctx, "SELECT poll_seconds FROM worker_cluster WHERE cluster=? AND worker=?", a.Cluster, w).Scan(&v.PollSeconds)
	if err != nil {
		return nil, err
	}
	if now.Admin {
		g, err := db.Governance(ctx, now)
		if err != nil {
			return nil, err
		}
		if directory {
			v.Devices = g.Devices
		}
		g.Devices = nil
		g.Limits = nil // Preserve the v1 signed view representation for older peers.
		v.Governance = &g
	} else if directory {
		v.Devices, err = readDirectory(ctx, db.sql, a, w, false)
		if err != nil {
			return nil, err
		}
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT DISTINCT s.pool FROM pool_services s WHERE s.cluster=? AND s.service=? AND
 (? OR EXISTS(SELECT 1 FROM effective_pool_grants g WHERE g.cluster=s.cluster AND g.pool=s.pool AND g.service=s.service AND g.principal IN (?,'public'))) ORDER BY s.pool`, a.Cluster, HashService, now.Admin, a.User)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pool string
		err = rows.Scan(&pool)
		if err != nil {
			return nil, err
		}
		if !domain.RetiredPool(pool) {
			v.Pools = append(v.Pools, pool)
		}
	}
	return &v, rows.Err()
}

func (db *DB) LocalReplica(ctx context.Context, w string, a Access) (SignedReplica, SyncCursor, error) {
	var r SignedReplica
	cursor := SyncCursor{Queues: map[string]int64{}}
	var raw []byte
	err := db.sql.QueryRowContext(ctx, "SELECT body FROM worker_replicas WHERE worker=? AND user=? AND device=?", w, a.User, a.Device).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return r, cursor, nil
	}
	if err != nil {
		return r, cursor, err
	}
	err = json.Unmarshal(raw, &r)
	if err != nil || !r.Valid(w, a) {
		return r, cursor, errors.New("invalid local replica")
	}
	cursor.Revision = r.View.Revision
	rows, err := db.sql.QueryContext(ctx, "SELECT pool,MAX(revision) FROM replica_queue WHERE worker=? AND user=? AND device=? GROUP BY pool", w, a.User, a.Device)
	if err != nil {
		return r, cursor, err
	}
	defer rows.Close()
	for rows.Next() {
		var pool string
		var rev int64
		err = rows.Scan(&pool, &rev)
		if err != nil {
			return r, cursor, err
		}
		cursor.Queues[pool] = rev
	}
	return r, cursor, rows.Err()
}

func (db *DB) commitView(ctx context.Context, tx *sql.Tx) error {
	var head uint64
	err := tx.QueryRowContext(ctx, "SELECT coalesce(max(revision),0) FROM view_revision").Scan(&head)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err == nil {
		db.viewRevision = head
	}
	return err
}
