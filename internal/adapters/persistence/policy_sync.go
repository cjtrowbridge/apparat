// Compares in-memory policy heads and serves bounded, ordered admin-only pages.
// The caller authenticates the principal before invoking this store; authorization is checked here.
package persistence

import (
	"context"
	"encoding/json"
	"errors"
)

type PolicyCursor struct {
	Revision uint64
	Hash     string
}

type PolicyPage struct {
	Head   PolicyCursor
	Events []PolicyEvent
}

func (db *DB) loadPolicyHeads(ctx context.Context) error {
	db.policyMu.Lock()
	defer db.policyMu.Unlock()
	heads := map[string]PolicyCursor{}
	rows, err := db.sql.QueryContext(ctx, `SELECT e.cluster,e.revision,e.hash FROM policy_events e
	WHERE e.revision=(SELECT MAX(other.revision) FROM policy_events other WHERE other.cluster=e.cluster)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cluster string
		var head PolicyCursor
		err = rows.Scan(&cluster, &head.Revision, &head.Hash)
		if err != nil {
			return err
		}
		heads[cluster] = head
	}
	err = rows.Err()
	if err == nil {
		db.policyHeads = heads
	}
	return err
}

func (db *DB) AdminPolicyPage(ctx context.Context, user, device, cluster string, after PolicyCursor) (PolicyPage, error) {
	db.policyMu.RLock()
	defer db.policyMu.RUnlock()
	var allowed int
	err := db.sql.QueryRowContext(ctx, `SELECT 1 FROM cluster_users u JOIN cluster_devices_authorized d USING(cluster,user_id)
	WHERE u.cluster=? AND u.user_id=? AND d.device_id=? AND u.admin=1 AND u.admission='approved' AND NOT EXISTS(SELECT 1 FROM revoked_devices WHERE cluster=u.cluster AND device=d.device_id)`, cluster, user, device).Scan(&allowed)
	if err != nil {
		return PolicyPage{}, errors.New("administrator authorization required")
	}
	page := PolicyPage{Head: db.policyHeads[cluster]}
	if after == page.Head {
		return page, nil
	}
	if after.Revision >= page.Head.Revision || after.Revision == 0 && after.Hash != "" {
		return PolicyPage{}, errors.New("unknown policy cursor")
	}
	rows, err := db.sql.QueryContext(ctx, "SELECT body FROM policy_events WHERE cluster=? AND revision>? ORDER BY revision LIMIT 64", cluster, after.Revision)
	if err != nil {
		return PolicyPage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var body []byte
		var event PolicyEvent
		err = rows.Scan(&body)
		if err == nil {
			err = json.Unmarshal(body, &event)
		}
		if err != nil || event.Previous != after.Hash {
			return PolicyPage{}, errors.New("discontinuous policy cursor")
		}
		page.Events = append(page.Events, event)
		after = PolicyCursor{event.Proposal.Edit.Base + 1, event.Hash()}
	}
	return page, rows.Err()
}
