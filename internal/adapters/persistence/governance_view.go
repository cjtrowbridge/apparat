// Reads a consistent admin view of admission, authority and pool membership.
// The view and policy cursor are sampled inside one authorized SQLite transaction.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type Member struct {
	ID, Name, Admission string
	Admin               bool
}
type Membership struct{ Pool, Service, Principal string }
type GovernanceView struct {
	Devices                           []domain.PoolDevice `json:",omitempty"`
	Epoch                             uint64
	Groups                            []string
	Memberships                       []GroupMember
	OriginalDevice                    identity.SignedBinding
	Cluster, Worker, Owner, Authority string
	PollSeconds                       int
	RegistrationOpen                  bool
	Head                              PolicyCursor
	Users                             []Member
	Pools                             []string
	Limits                            map[string]int `json:",omitempty"`
	Services, Grants                  []Membership
}

func (db *DB) Governance(ctx context.Context, a Access) (GovernanceView, error) {
	var v GovernanceView
	tx, err := db.sql.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return v, err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `SELECT c.cluster,c.worker,c.owner_user,c.authority_device,c.poll_seconds,c.admission_open,c.authority_epoch FROM worker_cluster c JOIN cluster_users u USING(cluster) JOIN cluster_devices_authorized d USING(cluster,user_id)
 WHERE c.cluster=? AND u.user_id=? AND d.device_id=? AND u.admin=1 AND u.admission='approved' AND NOT EXISTS(SELECT 1 FROM revoked_devices WHERE cluster=c.cluster AND device=d.device_id)`, a.Cluster, a.User, a.Device).Scan(&v.Cluster, &v.Worker, &v.Owner, &v.Authority, &v.PollSeconds, &v.RegistrationOpen, &v.Epoch)
	if err != nil {
		return GovernanceView{}, errors.New("administrator authorization required")
	}
	err = tx.QueryRowContext(ctx, "SELECT revision,hash FROM policy_events WHERE cluster=? ORDER BY revision DESC LIMIT 1", v.Cluster).Scan(&v.Head.Revision, &v.Head.Hash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return GovernanceView{}, err
	}
	var binding []byte
	err = tx.QueryRowContext(ctx, `SELECT public_binding FROM cluster_devices_authorized WHERE cluster=? AND user_id=? AND device_id=coalesce((SELECT json_extract(body,'$.Proposal.Edit.Service') FROM policy_events WHERE cluster=? AND json_extract(body,'$.Proposal.Edit.Kind') IN ('order_on_worker','recover_order') ORDER BY revision LIMIT 1),?)`, v.Cluster, v.Owner, v.Cluster, v.Authority).Scan(&binding)
	if err != nil {
		return GovernanceView{}, err
	}
	err = json.Unmarshal(binding, &v.OriginalDevice)
	if err != nil || !v.OriginalDevice.Valid() {
		return GovernanceView{}, errors.New("invalid original owner device")
	}
	rows, err := tx.QueryContext(ctx, "SELECT user_id,public_profile,admission,admin FROM cluster_users WHERE cluster=? ORDER BY user_id", v.Cluster)
	if err != nil {
		return GovernanceView{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		var profile identity.SignedProfile
		var data []byte
		err = rows.Scan(&m.ID, &data, &m.Admission, &m.Admin)
		if err != nil {
			return GovernanceView{}, err
		}
		err = json.Unmarshal(data, &profile)
		if err != nil || !profile.Valid() || identity.Fingerprint(profile.Profile.Public) != m.ID {
			return GovernanceView{}, errors.New("invalid stored user profile")
		}
		m.Name = profile.Profile.Name
		v.Users = append(v.Users, m)
	}
	err = rows.Err()
	if err != nil {
		return GovernanceView{}, err
	}
	rows, err = tx.QueryContext(ctx, "SELECT p.pool,coalesce(l.max_active,64) FROM cluster_pools p LEFT JOIN pool_limits l USING(cluster,pool) WHERE p.cluster=? ORDER BY p.pool", v.Cluster)
	if err != nil {
		return GovernanceView{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var limit int
		err = rows.Scan(&name, &limit)
		if err != nil {
			return GovernanceView{}, err
		}
		if v.Limits == nil {
			v.Limits = map[string]int{}
		}
		v.Limits[name] = limit
		v.Pools = append(v.Pools, name)
	}
	err = rows.Err()
	if err != nil {
		return GovernanceView{}, err
	}
	rows, err = tx.QueryContext(ctx, `SELECT pool,service,worker,0 FROM pool_services WHERE cluster=? UNION ALL SELECT pool,service,principal,1 FROM pool_grants WHERE cluster=? ORDER BY 4,1,2,3`, v.Cluster, v.Cluster)
	if err != nil {
		return GovernanceView{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Membership
		var grant bool
		err = rows.Scan(&m.Pool, &m.Service, &m.Principal, &grant)
		if err != nil {
			return GovernanceView{}, err
		}
		if grant {
			v.Grants = append(v.Grants, m)
		} else {
			v.Services = append(v.Services, m)
		}
	}
	err = rows.Err()
	if err == nil {
		err = v.readGroups(ctx, tx)
	}
	if err == nil {
		v.Devices, err = readDirectory(ctx, tx, a, v.Worker, true)
	}
	return v, err
}
