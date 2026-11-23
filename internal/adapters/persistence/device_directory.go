// Derives the visible device roster from current admission and signed policy.
// GUI assignments record preferences separately from executable worker services.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

const directorySchema = `
CREATE TABLE paired_gui_devices(owner TEXT NOT NULL,device TEXT NOT NULL,host TEXT NOT NULL,os TEXT NOT NULL,PRIMARY KEY(owner,device));
ALTER TABLE cluster_devices_authorized ADD COLUMN platform TEXT NOT NULL DEFAULT '';
CREATE TABLE gui_pool_devices(cluster TEXT NOT NULL,pool TEXT NOT NULL,device TEXT NOT NULL,PRIMARY KEY(cluster,pool,device), FOREIGN KEY(cluster,pool) REFERENCES cluster_pools(cluster,pool));
CREATE TRIGGER directory_insert AFTER INSERT ON cluster_devices_authorized BEGIN UPDATE view_revision SET revision=revision+1 WHERE cluster=NEW.cluster; END;
CREATE TRIGGER directory_update AFTER UPDATE OF hostname,platform ON cluster_devices_authorized WHEN OLD.hostname!=NEW.hostname OR OLD.platform!=NEW.platform BEGIN UPDATE view_revision SET revision=revision+1 WHERE cluster=NEW.cluster; END;
UPDATE view_revision SET revision=revision+1;
`

const approvedDevices = ` FROM cluster_devices_authorized d JOIN cluster_users u USING(cluster,user_id)
 WHERE d.cluster=? AND u.admission='approved' AND NOT EXISTS(SELECT 1 FROM revoked_devices r WHERE r.cluster=d.cluster AND r.device=d.device_id) `

func guiPoolEdit(ctx context.Context, tx *sql.Tx, e identity.PolicyEdit) (sql.Result, error) {
	switch e.Value {
	case "add":
		return tx.ExecContext(ctx, "INSERT INTO gui_pool_devices SELECT DISTINCT d.cluster,?,d.device_id"+approvedDevices+" AND d.device_id=?", e.Pool, e.Cluster, e.Target)
	case "remove":
		return tx.ExecContext(ctx, "DELETE FROM gui_pool_devices WHERE cluster=? AND pool=? AND device=?", e.Cluster, e.Pool, e.Target)
	default:
		return nil, errors.New("unknown device assignment")
	}
}

type directoryReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func readDirectory(ctx context.Context, db directoryReader, a Access, worker string, admin bool) ([]domain.PoolDevice, error) {
	rows, err := db.QueryContext(ctx, `WITH devices(id,host,os,worker) AS (
 SELECT DISTINCT d.device_id,d.hostname,d.platform,0 `+approvedDevices+` AND (? OR d.user_id=?)
 UNION ALL SELECT ?,'','',1), assignments(id,pool) AS (
 SELECT worker,pool FROM pool_services WHERE cluster=? AND service=?)
 SELECT devices.id,coalesce(n.name,host),os,worker,coalesce(pool,'') FROM devices LEFT JOIN cluster_device_names n ON n.cluster=? AND n.device=devices.id LEFT JOIN assignments ON assignments.id=devices.id AND (? OR EXISTS(SELECT 1 FROM effective_pool_grants g WHERE g.cluster=? AND g.pool=assignments.pool AND g.service=? AND g.principal IN (?,'public'))) ORDER BY worker,host COLLATE NOCASE,devices.id,pool`, a.Cluster, admin, a.User, worker, a.Cluster, HashService, a.Cluster, admin, a.Cluster, HashService, a.User)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []domain.PoolDevice
	for rows.Next() {
		var d domain.PoolDevice
		var pool string
		err = rows.Scan(&d.ID, &d.Host, &d.OS, &d.Worker, &pool)
		if err != nil {
			return nil, err
		}
		if len(devices) == 0 || devices[len(devices)-1].ID != d.ID {
			devices = append(devices, d)
		}
		if pool != "" {
			last := &devices[len(devices)-1]
			last.Pools = append(last.Pools, pool)
		}
	}
	return devices, rows.Err()
}

func (db *DB) RememberGUIPeer(ctx context.Context, user string, d domain.PoolDevice) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO paired_gui_devices VALUES(?,?,?,?) ON CONFLICT(owner,device) DO UPDATE SET host=excluded.host,os=excluded.os", user, d.ID, d.Host, d.OS)
	return err
}
func (db *DB) GUIPeers(ctx context.Context, user string) ([]domain.PoolDevice, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT device,host,os FROM paired_gui_devices WHERE owner=? ORDER BY host,device", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PoolDevice
	for rows.Next() {
		var d domain.PoolDevice
		err = rows.Scan(&d.ID, &d.Host, &d.OS)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func renameDevice(ctx context.Context, tx *sql.Tx, e identity.PolicyEdit) (sql.Result, error) {
	if !domain.ValidDeviceName(e.Value) {
		return nil, errors.New("choose a specific hostname using letters, digits and hyphens")
	}
	return tx.ExecContext(ctx, `INSERT INTO cluster_device_names(cluster,device,name) SELECT ?,?,?
 WHERE EXISTS(SELECT 1 FROM worker_cluster WHERE cluster=? AND worker=?) OR EXISTS(SELECT 1`+approvedDevices+` AND d.device_id=?)
 ON CONFLICT(cluster,device) DO UPDATE SET name=excluded.name`, e.Cluster, e.Target, e.Value, e.Cluster, e.Target, e.Cluster, e.Target)
}

// Replaces just the target's selectable assignments in the policy receipt transaction.
// A missing pool, stale revision or ineligible target rolls back the entire edit.
func replaceDevicePools(ctx context.Context, tx *sql.Tx, e identity.PolicyEdit) error {
	var pools []string
	if json.Unmarshal([]byte(e.Value), &pools) != nil || pools == nil {
		return errors.New("invalid pool selection")
	}
	var worker, approved bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM worker_cluster WHERE cluster=? AND worker=?), EXISTS(SELECT 1`+approvedDevices+` AND d.device_id=?)`, e.Cluster, e.Target, e.Cluster, e.Target).Scan(&worker, &approved)
	if err != nil {
		return err
	}
	if !worker && !approved {
		return errors.New("device is unavailable")
	}
	if worker {
		_, err = tx.ExecContext(ctx, "DELETE FROM pool_services WHERE cluster=? AND worker=? AND service=?", e.Cluster, e.Target, HashService)
	} else {
		_, err = tx.ExecContext(ctx, "DELETE FROM gui_pool_devices WHERE cluster=? AND device=?", e.Cluster, e.Target)
	}
	if err != nil {
		return err
	}
	for _, pool := range pools {
		if worker {
			_, err = tx.ExecContext(ctx, "INSERT INTO pool_services(cluster,pool,worker,service) VALUES(?,?,?,?)", e.Cluster, pool, e.Target, HashService)
		} else {
			_, err = tx.ExecContext(ctx, "INSERT INTO gui_pool_devices(cluster,pool,device) VALUES(?,?,?)", e.Cluster, pool, e.Target)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
