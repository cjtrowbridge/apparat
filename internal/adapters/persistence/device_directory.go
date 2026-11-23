// Derives the visible device roster from current admission and signed policy.
// GUI assignments record preferences separately from executable worker services.
package persistence

import (
	"context"
	"database/sql"
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
 SELECT device,pool FROM gui_pool_devices WHERE cluster=? UNION SELECT worker,pool FROM pool_services WHERE cluster=? AND service=?)
 SELECT devices.id,host,os,worker,coalesce(pool,'') FROM devices LEFT JOIN assignments ON assignments.id=devices.id AND (? OR EXISTS(SELECT 1 FROM effective_pool_grants g WHERE g.cluster=? AND g.pool=assignments.pool AND g.service=? AND g.principal IN (?,'public'))) ORDER BY worker,host COLLATE NOCASE,devices.id,pool`, a.Cluster, admin, a.User, worker, a.Cluster, a.Cluster, HashService, admin, a.Cluster, HashService, a.User)
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
