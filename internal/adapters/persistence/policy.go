// Orders authenticated admin edits and replays their signed hash chain into local SQLite.
// Policy records are administrative data; expose pages only after caller authorization.
package persistence

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const policySchema = `
CREATE TABLE cluster_pools(cluster TEXT NOT NULL REFERENCES worker_cluster(cluster), pool TEXT NOT NULL, PRIMARY KEY(cluster,pool));
CREATE TABLE pool_services(cluster TEXT NOT NULL,pool TEXT NOT NULL,worker TEXT NOT NULL,service TEXT NOT NULL,
 PRIMARY KEY(cluster,pool,worker,service),FOREIGN KEY(cluster,pool) REFERENCES cluster_pools(cluster,pool));
CREATE TABLE pool_grants(cluster TEXT NOT NULL,pool TEXT NOT NULL,principal TEXT NOT NULL,service TEXT NOT NULL,
 PRIMARY KEY(cluster,pool,principal,service),FOREIGN KEY(cluster,pool) REFERENCES cluster_pools(cluster,pool));
CREATE TABLE policy_events(cluster TEXT NOT NULL REFERENCES worker_cluster(cluster),revision INTEGER NOT NULL,command_id TEXT NOT NULL,
 hash TEXT NOT NULL,body BLOB NOT NULL, PRIMARY KEY(cluster,revision),UNIQUE(cluster,command_id));
`

type PolicyEvent struct {
	Proposal  identity.Proposal
	Previous  string
	Signature []byte
}

func (e PolicyEvent) message() []byte {
	e.Signature = nil
	data, _ := json.Marshal(e)
	return append([]byte("apparat/policy-receipt/v1\x00"), data...)
}

func (e PolicyEvent) Hash() string {
	data, _ := json.Marshal(e)
	return identity.ClaimDigest(string(data))
}

func authorizeEdit(ctx context.Context, tx *sql.Tx, p identity.Proposal) error {
	var admin int
	err := tx.QueryRowContext(ctx, `SELECT u.admin FROM cluster_users u JOIN cluster_devices_authorized d USING(cluster,user_id)
	WHERE u.cluster=? AND u.user_id=? AND u.admission='approved' AND d.device_id=? AND NOT EXISTS(SELECT 1 FROM revoked_devices WHERE cluster=u.cluster AND device=d.device_id)`, p.Edit.Cluster,
		identity.Fingerprint(p.Device.Binding.User), identity.Fingerprint(p.Device.Binding.Device)).Scan(&admin)
	if err != nil || admin != 1 {
		return errors.New("current cluster administrator required")
	}
	return nil
}

func applyEdit(ctx context.Context, tx *sql.Tx, e identity.PolicyEdit) error {
	if len(e.Target) > 256 || len(e.Service) > 128 || len(e.Pool) > 128 || len(e.Value) > 128 {
		return errors.New("policy value exceeds limit")
	}
	var result sql.Result
	var err error
	switch e.Kind {
	case "update_worker":
		if e.Target == "" {
			return errors.New("worker required")
		}
		result, err = tx.ExecContext(ctx, `INSERT INTO local_runtime(key,value,updated_at_ms)
   SELECT k,'1',0 FROM (SELECT 'worker_update_drain_requested' AS k UNION ALL SELECT 'worker_update_manual') WHERE EXISTS(SELECT 1 FROM worker_cluster WHERE cluster=? AND worker=?) ON CONFLICT(key) DO UPDATE SET value='1'`, e.Cluster, e.Target)
	case "poll_seconds":
		seconds, parseErr := strconv.Atoi(e.Value)
		if parseErr != nil || seconds < 1 || seconds > 86400 {
			return errors.New("invalid reconciliation interval")
		}
		result, err = tx.ExecContext(ctx, "UPDATE worker_cluster SET poll_seconds=? WHERE cluster=?", seconds, e.Cluster)
	case "admission_open":
		if e.Value != "0" && e.Value != "1" {
			return errors.New("invalid admission setting")
		}
		result, err = tx.ExecContext(ctx, "UPDATE worker_cluster SET admission_open=? WHERE cluster=?", e.Value, e.Cluster)
	case "user_role":
		if e.Value != "admin" && e.Value != "approved" && e.Value != "blocked" && e.Value != "pending" {
			return errors.New("invalid user role")
		}
		admission := e.Value
		admin := e.Value == "admin"
		if admin {
			admission = "approved"
		}
		result, err = tx.ExecContext(ctx, `UPDATE cluster_users SET admission=?,admin=? WHERE cluster=? AND user_id=?
		 AND (admin=0 OR ? OR (SELECT count(*) FROM cluster_users WHERE cluster=? AND admin=1)>1)`, admission, admin, e.Cluster, e.Target, admin, e.Cluster)
	case "order_on_worker", "recover_order":
		result, err = tx.ExecContext(ctx, `UPDATE worker_cluster SET authority_device=worker WHERE cluster=? AND worker=? AND authority_device=? AND owner_user=?`, e.Cluster, e.Target, e.Service, e.Value)
	case "revoke_device":
		result, err = tx.ExecContext(ctx, `INSERT INTO revoked_devices(cluster,device) SELECT cluster,device_id FROM cluster_devices_authorized
			WHERE cluster=? AND device_id=? AND device_id!=(SELECT authority_device FROM worker_cluster WHERE cluster=?) LIMIT 1`, e.Cluster, e.Target, e.Cluster)
	case "pool_quota":
		limit, parseErr := strconv.Atoi(e.Value)
		if parseErr != nil || limit < 1 || limit > 4096 {
			return errors.New("invalid pool quota")
		}
		result, err = tx.ExecContext(ctx, "INSERT INTO pool_limits VALUES(?,?,?) ON CONFLICT(cluster,pool) DO UPDATE SET max_active=excluded.max_active", e.Cluster, e.Pool, limit)
	case "add_group", "group_member":
		result, err = groupEdit(ctx, tx, e)
	case "add_pool":
		if e.Pool == "" {
			return errors.New("pool required")
		}
		result, err = tx.ExecContext(ctx, "INSERT INTO cluster_pools(cluster,pool) VALUES(?,?)", e.Cluster, e.Pool)
	case "rename_device":
		result, err = renameDevice(ctx, tx, e)
	case "gui_pool":
		result, err = guiPoolEdit(ctx, tx, e)
	case "pool_service", "pool_grant":
		if e.Pool == "" || e.Target == "" || e.Service == "" || e.Value != "add" && e.Value != "remove" {
			return errors.New("invalid pool membership")
		}
		table, column := "pool_services", "worker"
		if e.Kind == "pool_grant" {
			table, column = "pool_grants", "principal"
		}
		query := "INSERT INTO " + table + "(cluster,pool," + column + ",service) VALUES(?,?,?,?)"
		if e.Value == "remove" {
			query = "DELETE FROM " + table + " WHERE cluster=? AND pool=? AND " + column + "=? AND service=?"
		}
		result, err = tx.ExecContext(ctx, query, e.Cluster, e.Pool, e.Target, e.Service)
	default:
		return errors.New("unknown policy edit")
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	expected := int64(1)
	if e.Kind == "update_worker" {
		expected = 2
	}
	if err != nil || count != expected {
		return errors.New("policy target missing or last admin protected")
	}
	return nil
}

func (db *DB) CommitPolicy(ctx context.Context, p identity.Proposal, authority *identity.Session) (PolicyEvent, error) {
	return db.recordPolicy(ctx, PolicyEvent{Proposal: p}, authority, nil)
}

func (db *DB) ApplyPolicy(ctx context.Context, event PolicyEvent, authority ed25519.PublicKey) error {
	_, err := db.recordPolicy(ctx, event, nil, authority)
	return err
}

func (db *DB) recordPolicy(ctx context.Context, event PolicyEvent, signer *identity.Session, key ed25519.PublicKey) (PolicyEvent, error) {
	var empty PolicyEvent
	p := event.Proposal
	if !p.Valid() {
		return empty, errors.New("invalid policy signatures")
	}
	if signer != nil {
		key = signer.Public()
	}
	db.policyMu.Lock()
	defer db.policyMu.Unlock()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	var authority, worker, owner string
	var epoch uint64
	err = tx.QueryRowContext(ctx, "SELECT authority_device,worker,owner_user,authority_epoch FROM worker_cluster WHERE cluster=?", p.Edit.Cluster).Scan(&authority, &worker, &owner, &epoch)
	if err != nil {
		return empty, err
	}
	err = authorizeEdit(ctx, tx, p)
	if err != nil {
		return empty, err
	}
	var prior []byte
	err = tx.QueryRowContext(ctx, "SELECT body FROM policy_events WHERE cluster=? AND command_id=?", p.Edit.Cluster, p.Edit.ID).Scan(&prior)
	if err == nil {
		var stored PolicyEvent
		err = json.Unmarshal(prior, &stored)
		proposed, _ := json.Marshal(p)
		original, _ := json.Marshal(stored.Proposal)
		if err != nil || !bytes.Equal(proposed, original) || signer == nil && stored.Hash() != event.Hash() {
			return empty, errors.New("command ID conflict")
		}
		return stored, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return empty, err
	}
	orderKey := authority
	if p.Edit.Kind == "recover_order" {
		if authority == worker || p.Edit.Target != worker || p.Edit.Service != authority || p.Edit.Value != owner || identity.Fingerprint(p.Device.Binding.User) != owner || p.Edit.Epoch == nil {
			return empty, errors.New("owner recovery proof required")
		}
		orderKey = worker
	}
	if p.Edit.Epoch != nil && *p.Edit.Epoch != epoch {
		return empty, errors.New("stale authority epoch")
	}
	if len(key) != ed25519.PublicKeySize || identity.Fingerprint(key) != orderKey {
		return empty, errors.New("unknown policy authority")
	}
	if signer == nil && !ed25519.Verify(key, event.message(), event.Signature) {
		return empty, errors.New("invalid authority receipt")
	}
	var revision uint64
	var head string
	err = tx.QueryRowContext(ctx, "SELECT revision,hash FROM policy_events WHERE cluster=? ORDER BY revision DESC LIMIT 1", p.Edit.Cluster).Scan(&revision, &head)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return empty, err
	}
	if revision != p.Edit.Base || signer == nil && head != event.Previous {
		return empty, errors.New("stale or discontinuous policy revision")
	}
	if signer != nil {
		event.Previous = head
		event.Signature, err = signer.Sign(event.message())
		if err != nil || !ed25519.Verify(key, event.message(), event.Signature) {
			return empty, errors.New("authority signing failed")
		}
	}
	if p.Edit.Kind == "order_on_worker" && (p.Edit.Value != identity.Fingerprint(p.Device.Binding.User) || p.Edit.Service != authority) {
		return empty, errors.New("owner and current authority required for handoff")
	}
	err = applyEdit(ctx, tx, p.Edit)
	if err != nil {
		return empty, err
	}
	body, _ := json.Marshal(event)
	_, err = tx.ExecContext(ctx, "INSERT INTO policy_events(cluster,revision,command_id,hash,body) VALUES(?,?,?,?,?)", p.Edit.Cluster, revision+1, p.Edit.ID, event.Hash(), body)
	if err != nil {
		return empty, err
	}
	err = db.commitView(ctx, tx)
	if err != nil {
		return empty, err
	}
	db.policyHeads[p.Edit.Cluster] = PolicyCursor{revision + 1, event.Hash()}
	return event, nil
}

func SignHandoff(proposal identity.Proposal, previous string, authority *identity.Session) (PolicyEvent, error) {
	event := PolicyEvent{Proposal: proposal, Previous: previous}
	if !proposal.Valid() || proposal.Edit.Kind != "order_on_worker" || proposal.Edit.Service != identity.Fingerprint(authority.Public()) {
		return PolicyEvent{}, errors.New("invalid authority handoff")
	}
	signature, err := authority.Sign(event.message())
	event.Signature = signature
	return event, err
}
