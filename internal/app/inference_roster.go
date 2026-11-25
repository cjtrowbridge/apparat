// Derives inference membership from existing authenticated cluster registrations.
// Missing endpoints do not remove a previously authenticated worker from claim rounds.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"slices"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func membershipFor(cluster string, workers []domain.ClusterDevice) domain.InferenceRoster {
	r := domain.InferenceRoster{Cluster: cluster}
	for _, w := range workers {
		if w.Cluster == cluster {
			r.Peers = append(r.Peers, domain.InferencePeer{ID: w.Fingerprint, Address: w.Address})
		}
	}
	return r
}

// Called under the inference lock. Never rewrite the participants of an owned attempt.
func (p *inferencePool) mergeMembership(ctx context.Context, r domain.InferenceRoster) (domain.InferenceRoster, error) {
	if p.table == nil {
		return r, nil
	}
	if r.Cluster != p.roster.Cluster {
		return r, errors.New("cluster identity mismatch")
	}
	seen := map[string]bool{}
	peers := slices.Clone(r.Peers)
	for _, peer := range peers {
		seen[peer.ID] = true
	}
	for _, peer := range p.roster.Peers {
		if !seen[peer.ID] {
			peers = append(peers, peer)
		}
	}
	slices.SortFunc(peers, func(a, b domain.InferencePeer) int {
		return strings.Compare(a.ID, b.ID)
	})
	r.Peers = peers
	if r.Revision() != p.roster.Revision() {
		if p.withdrawal.Phase != "" || len(p.active) != 0 || p.table.Reserved() {
			return r, errors.New("retry membership after active claim")
		}
		err := p.db.PrepareMembershipChange(ctx, p.roster)
		if err != nil {
			return r, err
		}
	}
	return r, nil
}

// Relay independently verified peers across overlapping GUI directories. An offline peer
// remains in each worker's durable membership until an authenticated lifecycle transition.
func (n *Runtime) convergeMemberships(ctx context.Context, views map[string]domain.InferenceRoster, clients map[string]*httpapi.WorkerClient, workers []domain.ClusterDevice) {
	known := map[string]domain.InferencePeer{}
	for _, r := range views {
		for _, p := range r.Peers {
			known[p.ID] = p
		}
	}
	for _, r := range views {
		for _, p := range r.Peers {
			if clients[p.ID] != nil {
				continue
			}
			c, e := n.cachedWorker(ctx, p.ID+"/"+p.Address, func() (*httpapi.WorkerClient, error) {
				return httpapi.OpenWorker(ctx, p.Address, p.ID, n.Personal, n.Keys, &n.observations)
			})
			if e == nil && c.Access.Cluster == r.Cluster && c.Access.Admission == "approved" {
				clients[p.ID] = c
			}
		}
	}
	// A second GUI can learn a worker through a pinned roster before it has a
	// direct registration. Count it only while an authenticated roster lists it.
	local := map[string]domain.ClusterDevice{}
	for _, w := range workers {
		local[w.Fingerprint] = w
	}
	samples := map[string]domain.WorkerInventory{}
	for id, c := range clients {
		if local[id].Fingerprint != "" {
			continue
		}
		report, e := c.Inventory(ctx)
		if e == nil {
			samples[id] = report
		}
	}
	for id, r := range views {
		c := clients[id]
		previous := local[id]
		members := domain.RosterInventory{Hardware: map[string]domain.HardwareCapacity{}, Usage: map[string]domain.DeviceUsage{}}
		for _, peer := range r.Peers {
			if local[peer.ID].Fingerprint != "" {
				continue
			}
			members.Hardware[peer.ID] = previous.RosterHardware[peer.ID]
			prior, ok := previous.RosterUsage[peer.ID]
			if ok {
				members.Usage[peer.ID] = prior
			}
			sample, ok := samples[peer.ID]
			if ok && clients[peer.ID].Access.Cluster == r.Cluster {
				if sample.Resources.CPUs > 0 {
					members.Hardware[peer.ID] = sample.Capacity()
				}
				if sample.Usage != nil {
					members.Usage[peer.ID] = *sample.Usage
				}
			}
		}
		n.mu.Lock()
		if n.database != nil {
			_ = n.database.SaveRosterInventory(ctx, id, c.Access, members)
		}
		n.mu.Unlock()
	}
	for id, c := range clients {
		if !c.Access.Admin {
			continue
		}
		combined := views[id]
		if combined.Cluster == "" {
			combined.Cluster = c.Access.Cluster
		}
		combined.Peers = slices.Clone(combined.Peers)
		for other, peer := range known {
			v := clients[other]
			if v != nil && v.Access.Cluster == c.Access.Cluster && !slices.Contains(combined.IDs(), other) {
				combined.Peers = append(combined.Peers, peer)
			}
		}
		if len(combined.Peers) > len(views[id].Peers) {
			_, _ = c.Inference(ctx, domain.InferenceCommand{Operation: "status", Roster: combined})
		}
	}
}
