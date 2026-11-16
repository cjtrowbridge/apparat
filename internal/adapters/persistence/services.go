// Persists service intent and replaces observations transactionally.
// Revision and timestamp checks reject stale probe results after edits/deletion.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (db *DB) SaveService(ctx context.Context, service domain.Service, expected int64) error {
	err := domain.ValidateService(service)
	if err != nil {
		return err
	}
	if service.Revision != expected+1 {
		return domain.ErrConflict
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if expected == 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO services(id,name,driver,enabled,fail,revision) VALUES(?,?,?,?,?,?)`, service.ID, service.Name, service.Driver, service.Enabled, service.Fail, service.Revision)
	} else {
		var result sql.Result
		result, err = tx.ExecContext(ctx, `UPDATE services SET name=?,driver=?,enabled=?,fail=?,revision=? WHERE id=? AND revision=?`, service.Name, service.Driver, service.Enabled, service.Fail, service.Revision, service.ID, expected)
		if err == nil {
			var count int64
			count, err = result.RowsAffected()
			if err == nil && count != 1 {
				return domain.ErrConflict
			}
		}
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM service_observations WHERE service_id=?`, service.ID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM service_capabilities WHERE service_id=?`, service.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (db *DB) RemoveService(ctx context.Context, id string) error {
	result, err := db.sql.ExecContext(ctx, `DELETE FROM services WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return domain.ErrNotFound
	}
	return err
}
func (db *DB) Observe(ctx context.Context, observation domain.Observation) error {
	if observation.At < 0 || (observation.Error != "" && observation.Error != "probe_failed") {
		return errors.New("invalid observation")
	}
	if !observation.Healthy {
		observation.Capabilities = nil
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	var revision int64
	var enabled bool
	err = tx.QueryRowContext(ctx, `SELECT revision,enabled FROM services WHERE id=?`, observation.ServiceID).Scan(&revision, &enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if revision != observation.Revision || !enabled {
		return domain.ErrConflict
	}
	var previous int64
	err = tx.QueryRowContext(ctx, `SELECT at_ms FROM service_observations WHERE service_id=?`, observation.ServiceID).Scan(&previous)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if observation.At < previous {
		return domain.ErrConflict
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO service_observations VALUES(?,?,?,?,?) ON CONFLICT(service_id) DO UPDATE SET revision=excluded.revision,at_ms=excluded.at_ms,healthy=excluded.healthy,safe_error=excluded.safe_error`, observation.ServiceID, revision, observation.At, observation.Healthy, observation.Error)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM service_capabilities WHERE service_id=?`, observation.ServiceID)
	if err != nil {
		return err
	}
	for _, capability := range observation.Capabilities {
		if capability.ServiceID != observation.ServiceID || !domain.ValidID(capability.ID) || (capability.Workload != "text_generation" && capability.Workload != "image_generation") {
			return errors.New("invalid capability ownership")
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO service_capabilities VALUES(?,?,?)`, capability.ServiceID, capability.ID, capability.Workload)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (db *DB) Services(ctx context.Context) ([]domain.Service, error) {
	tx, err := db.sql.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	rows, err := tx.QueryContext(ctx, `SELECT s.id,s.name,s.driver,s.enabled,s.fail,s.revision,COALESCE(o.at_ms,0),COALESCE(o.healthy,0),COALESCE(o.safe_error,'') FROM services s LEFT JOIN service_observations o ON o.service_id=s.id ORDER BY s.id`)
	if err != nil {
		return nil, err
	}
	services := []domain.Service{}
	for rows.Next() {
		var service domain.Service
		var healthy bool
		err = rows.Scan(&service.ID, &service.Name, &service.Driver, &service.Enabled, &service.Fail, &service.Revision, &service.ObservedAt, &healthy, &service.Error)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		service.Health = "unknown"
		if service.ObservedAt > 0 {
			service.Health = "unhealthy"
			if healthy {
				service.Health = "healthy"
			}
		}
		if !service.Enabled {
			service.Health = "disabled"
		}
		service.Capabilities = []domain.Capability{}
		services = append(services, service)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	for index := range services {
		rows, err = tx.QueryContext(ctx, `SELECT id,workload FROM service_capabilities WHERE service_id=? ORDER BY id`, services[index].ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			capability := domain.Capability{ServiceID: services[index].ID}
			err = rows.Scan(&capability.ID, &capability.Workload)
			if err != nil {
				_ = rows.Close()
				return nil, err
			}
			services[index].Capabilities = append(services[index].Capabilities, capability)
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return services, tx.Commit()
}

// SeedServices commits the demonstration as one unit only into an empty store.
// Restarting after interruption sees either all three rows or none.
func (db *DB) SeedServices(ctx context.Context, services []domain.Service) error {
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	var count int
	err = tx.QueryRowContext(ctx, "SELECT count(*) FROM services").Scan(&count)
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	for _, service := range services {
		{
			err = domain.ValidateService(service)
			if err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO services(id,name,driver,enabled,fail,revision) VALUES(?,?,?,?,?,?)`, service.ID, service.Name, service.Driver, service.Enabled, service.Fail, service.Revision)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
