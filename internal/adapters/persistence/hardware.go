// Stores installed capacity separately from live availability and model inventory.
// GUI reports bind to their authenticated device and are returned under directory visibility rules.
package persistence

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (db *DB) MemberHardware(ctx context.Context, a Access, report domain.HardwareCapacity) (map[string]domain.HardwareCapacity, error) {
	if report.Threads > 0 {
		raw, err := json.Marshal(report)
		if err != nil {
			return nil, err
		}
		_, err = db.sql.ExecContext(ctx, "UPDATE cluster_devices_authorized SET hardware=? WHERE cluster=? AND user_id=? AND device_id=? AND hardware IS NOT ?", raw, a.Cluster, a.User, a.Device, raw)
		if err != nil {
			return nil, err
		}
	}
	rows, err := db.sql.QueryContext(ctx, "SELECT d.device_id,d.hardware"+approvedDevices+" AND (? OR d.user_id=?) AND d.hardware IS NOT NULL", a.Cluster, a.Admin, a.User)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]domain.HardwareCapacity{}
	for rows.Next() {
		var id string
		var raw []byte
		err = rows.Scan(&id, &raw)
		if err != nil {
			return nil, err
		}
		var h domain.HardwareCapacity
		err = json.Unmarshal(raw, &h)
		if err != nil {
			return nil, err
		}
		out[id] = h
	}
	return out, rows.Err()
}

func (db *DB) SaveHardware(ctx context.Context, worker string, a Access, reports map[string]domain.HardwareCapacity) error {
	raw, err := json.Marshal(reports)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "UPDATE claimed_workers SET hardware=? WHERE worker=? AND owner=? AND cluster=? AND hardware IS NOT ?", raw, worker, a.User, a.Cluster, raw)
	return err
}

func (db *DB) SaveRosterInventory(ctx context.Context, worker string, a Access, reports domain.RosterInventory) error {
	raw, err := json.Marshal(reports)
	if err != nil {
		return err
	}
	_, err = db.sql.ExecContext(ctx, "UPDATE claimed_workers SET roster_hardware=? WHERE worker=? AND owner=? AND cluster=? AND roster_hardware IS NOT ?", raw, worker, a.User, a.Cluster, raw)
	return err
}
