// Verifies ordered admin receipts from the original owner device through worker handoff.
// The app journals verified events separately from the most recent worker snapshot.
package persistence

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

const replicaPolicySchema = `
CREATE TABLE replica_policy(worker TEXT NOT NULL,user TEXT NOT NULL,device TEXT NOT NULL,revision INTEGER NOT NULL,hash TEXT NOT NULL,body BLOB NOT NULL,PRIMARY KEY(worker,user,device,revision));
`

func (db *DB) ReplicaPolicyCursor(ctx context.Context, w string, a Access) (PolicyCursor, error) {
	var c PolicyCursor
	err := db.sql.QueryRowContext(ctx, "SELECT revision,hash FROM replica_policy WHERE worker=? AND user=? AND device=? ORDER BY revision DESC LIMIT 1", w, a.User, a.Device).Scan(&c.Revision, &c.Hash)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return c, err
}

func (db *DB) ApplyReceiptPage(ctx context.Context, w string, a Access, r SignedReplica, start PolicyCursor, page PolicyPage) error {
	if !r.Valid(w, a) || !r.View.Admin || r.View.Governance == nil || !a.Admin {
		return errors.New("current admin view required")
	}
	g := r.View.Governance
	goal := g.Head
	if start.Revision > goal.Revision || page.Head.Revision < goal.Revision || len(page.Events) == 0 {
		return errors.New("unexpected receipt head")
	}
	origin := g.OriginalDevice
	if !origin.Valid() || identity.Fingerprint(origin.Binding.User) != g.Owner {
		return errors.New("missing original owner binding")
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var rev uint64
	err = tx.QueryRowContext(ctx, "SELECT revision FROM worker_replicas WHERE worker=? AND user=? AND device=?", w, a.User, a.Device).Scan(&rev)
	if err != nil || rev != r.View.Revision {
		return errors.New("admin snapshot changed")
	}
	var current PolicyCursor
	err = tx.QueryRowContext(ctx, "SELECT revision,hash FROM replica_policy WHERE worker=? AND user=? AND device=? ORDER BY revision DESC LIMIT 1", w, a.User, a.Device).Scan(&current.Revision, &current.Hash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if start != current {
		return errors.New("stale local receipt cursor")
	}
	var handed bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM replica_policy WHERE worker=? AND user=? AND device=? AND json_extract(body,'$.Proposal.Edit.Kind') IN ('order_on_worker','recover_order'))`, w, a.User, a.Device).Scan(&handed)
	if err != nil {
		return err
	}
	for _, e := range page.Events {
		if current.Revision == goal.Revision {
			break
		}
		p := e.Proposal
		key := origin.Binding.Device
		if handed || p.Edit.Kind == "recover_order" {
			key = r.Key
		}
		if !p.Valid() || p.Edit.Cluster != g.Cluster || p.Edit.Base != current.Revision || e.Previous != current.Hash || !ed25519.Verify(key, e.message(), e.Signature) {
			return errors.New("invalid ordered policy receipt")
		}
		epoch := uint64(0)
		if handed {
			epoch = 1
		}
		if p.Edit.Epoch != nil && *p.Edit.Epoch != epoch {
			return errors.New("invalid receipt epoch")
		}
		if p.Edit.Kind == "order_on_worker" || p.Edit.Kind == "recover_order" {
			if handed || p.Edit.Target != w || p.Edit.Service != identity.Fingerprint(origin.Binding.Device) || p.Edit.Value != g.Owner || identity.Fingerprint(p.Device.Binding.User) != g.Owner {
				return errors.New("invalid owner handoff")
			}
			handed = true
		}
		current = PolicyCursor{Revision: current.Revision + 1, Hash: e.Hash()}
		body, err := json.Marshal(e)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO replica_policy VALUES(?,?,?,?,?,?)", w, a.User, a.Device, current.Revision, current.Hash, body)
		if err != nil {
			return err
		}
	}
	if current.Revision == goal.Revision && current != goal {
		return errors.New("policy snapshot and receipt chain disagree")
	}
	return tx.Commit()
}
