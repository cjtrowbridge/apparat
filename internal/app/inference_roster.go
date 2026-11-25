// Presents the administrator's explicit validator roster, including offline saved workers.
// The same pinned identities and revision must be accepted on every worker.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) rosterPage(ctx context.Context, cluster string) (p domain.ClusterPage, err error) {
	devices, err := n.KnownWorkers(ctx)
	if err != nil {
		return p, err
	}
	r := domain.InferenceRoster{Cluster: cluster}
	p.Title = "Inference validators"
	p.Detail = "Every listed worker must confirm a claim before inference starts. An offline worker holds new work in the queue."
	for _, d := range devices {
		if d.Cluster != cluster {
			continue
		}
		if !d.Admin {
			return p, errors.New("administrator access required on each worker")
		}
		r.Peers = append(r.Peers, domain.InferencePeer{ID: d.Fingerprint, Address: d.Address})
		p.Detail += "\n" + d.DisplayName() + " · " + d.Address
	}
	if len(r.Peers) == 0 {
		return p, errors.New("no saved workers for this cluster")
	}
	err = r.Validate(r.Peers[0].ID)
	if err != nil {
		return p, err
	}
	p.Options = append(p.Options, domain.ClusterOption{Label: "Enroll listed validators", Confirm: "Pin these workers for inference claims?", Run: func(ctx context.Context, _ string) error {
		for _, peer := range r.Peers {
			c, e := n.workspaceClient(ctx, "worker-"+peer.ID)
			if e != nil {
				return e
			}
			_, e = c.Inference(ctx, domain.InferenceCommand{Operation: "roster", Roster: r})
			if e != nil {
				return e
			}
		}
		select {
		case n.inventoryHint <- struct{}{}:
		default:
		}
		return nil
	}})
	return p, nil
}
