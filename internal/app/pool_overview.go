// Builds routing pool tabs from authenticated SQLite replicas.
// Live administrative edits continue to use the worker's signed policy path.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"slices"
	"strings"
)

func poolAuthorized(v *persistence.GovernanceView, member persistence.Member, pool string) bool {
	if member.Admission != "approved" {
		return false
	}
	if member.Admin {
		return true
	}
	for _, g := range v.Grants {
		if g.Pool != pool || g.Service != persistence.HashService {
			continue
		}
		if g.Principal == "public" || g.Principal == member.ID {
			return true
		}
		for _, m := range v.Memberships {
			if m.User == member.ID && g.Principal == "group:"+m.Name {
				return true
			}
		}
	}
	return false
}

func (n *Runtime) poolOverview(ctx context.Context, r domain.ClusterRoute) (domain.ClusterPage, error) {
	overview := &domain.PoolOverview{}
	page := domain.ClusterPage{Title: r.Pool, Detail: "Tasks, workers and access at the last cluster synchronization.", Overview: overview}
	for _, p := range domain.DefaultPools() {
		if p.Name == r.Pool {
			page.Detail = p.Priority + " priority. " + page.Detail
		}
	}
	workers, err := n.KnownWorkers(ctx)
	if err != nil {
		return page, err
	}
	n.mu.Lock()
	db := n.database
	n.mu.Unlock()
	if db == nil || n.Personal == nil {
		return page, nil
	}
	user := n.Personal.Fingerprint()
	device := n.LocalClusterDevice().ID
	users := map[string]string{}
	for _, w := range workers {
		if (r.Worker != "" && r.Worker != w.Fingerprint) || !slices.Contains(w.PoolNames, r.Pool) {
			continue
		}
		a := persistence.Access{Cluster: w.Cluster, User: user, Device: device}
		replica, _, err := db.LocalReplica(ctx, w.Fingerprint, a)
		if err != nil {
			return page, err
		}
		v := replica.View
		if v.Revision == 0 {
			continue
		}
		for _, d := range v.Devices {
			if d.Worker && d.ID == w.Fingerprint && slices.Contains(d.Pools, r.Pool) {
				overview.Devices = append(overview.Devices, domain.ClusterOption{Label: w.DisplayName(), Route: domain.ClusterRoute{Page: "device-detail", Target: "worker-" + d.ID}})
			}
		}
		if slices.Contains(v.Pools, r.Pool) {
			users[user] = n.ScreenName()
			tasks, err := db.LocalPoolTasks(ctx, w.Fingerprint, a, r.Pool)
			if err != nil {
				return page, err
			}
			for _, t := range tasks {
				overview.Tasks = append(overview.Tasks, domain.PoolTaskRow{ID: t.ID, State: t.State, Changed: t.Changed, Worker: w.Fingerprint, Host: w.DisplayName()})
			}
		}
		gov := v.Governance
		if gov == nil || !v.Admin {
			continue
		}
		for _, m := range gov.Users {
			if poolAuthorized(gov, m, r.Pool) {
				users[m.ID] = m.Name
			}
		}
		opt := domain.ClusterOption{Label: w.DisplayName(), Route: domain.ClusterRoute{Page: "pool-add-user", Worker: w.Fingerprint, Pool: r.Pool}}
		if gov.Authority == w.Fingerprint {
			overview.Add = append(overview.Add, opt)
		}
		opt.Route.Page = "pool-policy"
		opt.Label = "Pool settings · " + w.DisplayName()
		overview.Policy = append(overview.Policy, opt)
	}
	for id, name := range users {
		if name == "" {
			name = id
		}
		if id == user {
			name += " (you)"
		}
		overview.Users = append(overview.Users, name)
	}
	slices.Sort(overview.Users)
	slices.SortFunc(overview.Tasks, func(a, b domain.PoolTaskRow) int {
		if a.Changed > b.Changed {
			return -1
		}
		if a.Changed < b.Changed {
			return 1
		}
		if a.Worker != b.Worker {
			return strings.Compare(a.Worker, b.Worker)
		}
		return strings.Compare(a.ID, b.ID)
	})
	return page, nil
}

func (b *clusterPageBuilder) poolAddUser() {
	b.Title = "Add authorized user"
	b.Detail = b.route.Pool + " · " + b.worker.DisplayName()
	for _, m := range b.view.Users {
		if m.Admission == "approved" && !poolAuthorized(&b.view, m, b.route.Pool) {
			b.edit(m.Name+" · "+m.ID[:min(8, len(m.ID))], "", "Allow "+m.Name+" to use "+b.route.Pool+" on "+b.worker.DisplayName()+"?", identity.PolicyEdit{Kind: "pool_grant", Pool: b.route.Pool, Service: persistence.HashService, Target: m.ID, Value: "add"})
			b.Options[len(b.Options)-1].Route = domain.ClusterRoute{Page: "pool-add-user", Worker: b.route.Worker, Pool: b.route.Pool, Target: m.ID}
		}
	}
	if len(b.Options) == 0 {
		b.Detail += "\nAll approved users already have access."
	}
}
