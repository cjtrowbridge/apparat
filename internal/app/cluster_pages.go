// Loads worker-scoped cluster settings through the saved, pinned connection.
// Option callbacks submit a signed edit against the exact revision shown.
package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type clusterPageBuilder struct {
	domain.ClusterPage
	route   domain.ClusterRoute
	worker  domain.ClusterDevice
	client  *httpapi.WorkerClient
	runtime *Runtime
	view    persistence.GovernanceView
}

func (n *Runtime) ClusterPage(ctx context.Context, r domain.ClusterRoute) (domain.ClusterPage, error) {
	if r.Page == "ollama" {
		return n.inventoryPage(ctx, r)
	}
	if r.Page == "performance" {
		return n.performance(ctx, r)
	}
	if r.Page == "pool" {
		return n.poolOverview(ctx, r)
	}
	if r.Page == "rename-local" {
		return n.localRenamePage(r)
	}
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return domain.ClusterPage{}, err
	}
	for _, w := range rows {
		if w.Fingerprint != r.Worker {
			continue
		}
		c, err := n.openSavedWorker(ctx, w)
		if err != nil {
			return domain.ClusterPage{}, err
		}
		b := clusterPageBuilder{route: r, worker: w, client: c, runtime: n}
		switch r.Page {
		case "tasks", "task", "work":
			err = b.taskPage(ctx)
		default:
			b.view, err = c.Governance(ctx)
			if err == nil {
				err = b.settingsPage()
			}
		}
		return b.ClusterPage, err
	}
	return domain.ClusterPage{}, errors.New("worker no longer available")
}
func (b *clusterPageBuilder) link(label, page, pool, target string) {
	b.Options = append(b.Options, domain.ClusterOption{Label: label, Route: domain.ClusterRoute{Worker: b.route.Worker, Page: page, Pool: pool, Target: target}})
}
func commandID() string {
	var id [16]byte
	rand.Read(id[:])
	return hex.EncodeToString(id[:])
}
func (b *clusterPageBuilder) edit(label, input, confirm string, e identity.PolicyEdit) {
	v := b.view
	e.ID = commandID()
	e.Cluster = v.Cluster
	e.Base = v.Head.Revision
	e.Epoch = &v.Epoch
	b.Options = append(b.Options, domain.ClusterOption{Label: label, Input: input, Confirm: confirm, Run: func(ctx context.Context, value string) error {
		edit := e
		if input != "" {
			if strings.TrimSpace(value) == "" {
				return errors.New("enter a value")
			}
			if edit.Kind == "add_pool" || edit.Kind == "add_group" {
				edit.Pool = value
			} else {
				edit.Value = value
			}
		}
		_, err := b.client.ChangePolicy(ctx, b.runtime.Personal, b.runtime.Keys, edit, v.Head.Hash)
		if err == nil {
			if b.route.Page == "pool-add-user" {
				b.runtime.mu.Lock()
				db := b.runtime.database
				b.runtime.mu.Unlock()
				if db != nil {
					_, syncErr := b.client.Sync(ctx, db)
					if syncErr != nil {
						return errors.New("access granted; the local directory will refresh on the next synchronization")
					}
				}
			}
			b.runtime.RequestSync()
		}
		return err
	}})
}
func (b *clusterPageBuilder) settingsPage() error {
	v, r := b.view, b.route
	if !b.client.Access.Admin {
		return errors.New("current administrator access required")
	}
	b.Detail = fmt.Sprintf("Policy revision %d on %s", v.Head.Revision, b.worker.DisplayName())
	if r.Page != "settings" && r.Page != "sync" && v.Authority != v.Worker {
		return errors.New("enable worker ordering in Synchronization first")
	}
	switch r.Page {
	case "settings":
		b.Title = "Cluster settings"
		b.link("People and groups", "people", "", "")
		b.link("Pools", "pools", "", "")
		b.link("Synchronization", "sync", "", "")
	case "sync":
		b.Title = "Synchronization"
		b.Detail += fmt.Sprintf("\nCheck interval: %d seconds", v.PollSeconds)
		if v.Authority == v.Worker {
			b.Detail += "\nOrdering device: " + b.worker.DisplayName()
			b.edit("Change interval", "Seconds between checks", "", identity.PolicyEdit{Kind: "poll_seconds"})
		} else {
			b.Detail += "\nOrdering device: " + v.Authority
			if v.Owner == b.runtime.Personal.Fingerprint() {
				kind, label := "order_on_worker", "Enable worker ordering"
				if v.Authority != identity.Fingerprint(b.runtime.Keys.Public()) {
					kind, label = "recover_order", "Recover lost authority"
				}
				b.edit(label, "", "Transfer policy ordering to "+b.worker.DisplayName()+"?", identity.PolicyEdit{Kind: kind, Target: v.Worker, Value: v.Owner, Service: v.Authority})
			}
		}
	case "rename":
		for _, d := range v.Devices {
			if d.ID == r.Target {
				b.Title = "Rename " + d.DisplayName(b.worker.DisplayName())
				b.Detail = "Choose a hostname for this device in the cluster."
				b.edit("Save name", "Device hostname", "", identity.PolicyEdit{Kind: "rename_device", Target: d.ID})
				return nil
			}
		}
		return errors.New("device unavailable")
	case "device-pools":
		return b.devicePools()
	case "machine":
		b.Title = b.worker.DisplayName()
		value, label := "1", "Open registration"
		if v.RegistrationOpen {
			value, label = "0", "Close registration"
		}
		b.edit(label, "", label+" on "+b.worker.DisplayName()+"?", identity.PolicyEdit{Kind: "admission_open", Value: value})
		b.edit("Drain and check for update", "", "Drain "+b.worker.DisplayName()+" and check for an update?", identity.PolicyEdit{Kind: "update_worker", Target: v.Worker})
	case "pools":
		b.Title = "Pools"
		for _, p := range domain.DefaultPools() {
			if slices.Contains(v.Pools, p.Name) {
				b.link(p.Name+" · "+p.Priority, "pool", p.Name, "")
			}
		}
		for _, p := range v.Pools {
			if !slices.ContainsFunc(domain.DefaultPools(), func(d domain.PoolDefault) bool {
				return d.Name == p
			}) {
				b.link(p, "pool", p, "")
			}
		}
		b.edit("Create pool", "Pool name", "", identity.PolicyEdit{Kind: "add_pool"})
	case "pool-policy", "pool-add-user", "access":
		if !slices.Contains(v.Pools, r.Pool) {
			return errors.New("pool unavailable")
		}
		if r.Page == "pool-add-user" {
			b.poolAddUser()
			break
		}
		b.Title = r.Pool
		if r.Page == "access" {
			b.Title += " · Access"
			b.poolAccess()
			break
		}
		b.poolDevices()
		quota := v.Limits[r.Pool]
		if quota > 0 {
			b.Detail += "\nActive tasks per user: " + strconv.Itoa(quota)
		}
		service := persistence.HashService
		has := slices.Contains(v.Services, persistence.Membership{Pool: r.Pool, Service: service, Principal: v.Worker})
		label, value := toggle(has, "Enable", "Disable")
		b.edit(label+" "+service+" on "+b.worker.DisplayName(), "", "", identity.PolicyEdit{Kind: "pool_service", Pool: r.Pool, Service: service, Target: v.Worker, Value: value})
		b.link("Access", "access", r.Pool, "")
		b.edit("Change quota", "Maximum active tasks per user", "", identity.PolicyEdit{Kind: "pool_quota", Pool: r.Pool})
		if has {
			b.link("My tasks", "tasks", r.Pool, "")
		}
	case "people", "user", "group":
		return b.people()
	default:
		return errors.New("unknown cluster page")
	}
	return nil
}
func toggle(on bool, add, remove string) (string, string) {
	if on {
		return remove, "remove"
	}
	return add, "add"
}
func (b *clusterPageBuilder) poolAccess() {
	v, r := b.view, b.route
	option := func(label, id string) {
		m := persistence.Membership{Pool: r.Pool, Service: persistence.HashService, Principal: id}
		action, value := toggle(slices.Contains(v.Grants, m), "Grant", "Remove")
		description := action + " hash-text access · " + label
		b.edit(description, "", description+" in "+r.Pool+"?", identity.PolicyEdit{Kind: "pool_grant", Pool: r.Pool, Service: m.Service, Target: id, Value: value})
	}
	option("Public", "public")
	for _, g := range v.Groups {
		option("Group · "+g, "group:"+g)
	}
	for _, u := range v.Users {
		option(u.Name+" · "+u.ID, u.ID)
	}
}
func (b *clusterPageBuilder) people() error {
	v, r := b.view, b.route
	switch r.Page {
	case "people":
		b.Title = "People and groups"
		for _, u := range v.Users {
			role := u.Admission
			if u.Admin {
				role = "admin"
			}
			b.link(u.Name+" · "+role, "user", "", u.ID)
		}
		for _, g := range v.Groups {
			b.link("Group · "+g, "group", "", g)
		}
		b.edit("Create group", "Group name", "", identity.PolicyEdit{Kind: "add_group"})
	case "group":
		if !slices.Contains(v.Groups, r.Target) {
			return errors.New("group unavailable")
		}
		b.Title = r.Target
		for _, u := range v.Users {
			label, value := toggle(slices.Contains(v.Memberships, persistence.GroupMember{Name: r.Target, User: u.ID}), "Add", "Remove")
			b.edit(label+" "+u.Name, "", label+" "+u.Name+" for "+r.Target+"?", identity.PolicyEdit{Kind: "group_member", Pool: r.Target, Target: u.ID, Value: value})
		}
	case "user":
		for _, u := range v.Users {
			if u.ID == r.Target {
				b.Title = u.Name
				role := u.Admission
				if u.Admin {
					role = "admin"
				}
				b.Detail += "\n" + u.ID + " · " + role
				for _, next := range []string{"approved", "admin", "blocked", "pending"} {
					if next != role {
						b.edit("Set "+next, "", "Set "+u.Name+" to "+next+" on "+b.worker.DisplayName()+"?", identity.PolicyEdit{Kind: "user_role", Target: u.ID, Value: next})
					}
				}
				return nil
			}
		}
		return errors.New("user unavailable")
	}
	return nil
}

func (b *clusterPageBuilder) poolDevices() {
	names := []string{}
	for _, d := range b.view.Devices {
		if d.Worker && slices.Contains(d.Pools, b.route.Pool) {
			name := d.DisplayName(b.worker.DisplayName())
			names = append(names, name)
			b.link(name+" · Pools", "device-pools", "", d.ID)
		}
	}
	if len(names) == 0 {
		b.Detail += "\nDevices: none assigned"
	} else {
		b.Detail += "\nDevices: " + strings.Join(names, ", ")
	}
}

func (b *clusterPageBuilder) devicePools() error {
	for _, d := range b.view.Devices {
		if !d.Worker || d.ID != b.route.Target {
			continue
		}
		b.Title = d.DisplayName(b.worker.DisplayName()) + " · Pools"
		b.Detail = "Select the pools this device should take requests from."
		b.Pools = b.view.Pools
		b.Assigned = d.Pools
		v := b.view
		id := commandID()
		b.SavePools = func(ctx context.Context, pools []string) error {
			value, err := json.Marshal(pools)
			if err != nil {
				return err
			}
			edit := identity.PolicyEdit{Kind: "device_pools", Target: d.ID, ID: id, Cluster: v.Cluster, Base: v.Head.Revision, Epoch: &v.Epoch, Value: string(value)}
			_, err = b.client.ChangePolicy(ctx, b.runtime.Personal, b.runtime.Keys, edit, v.Head.Hash)
			if err == nil {
				n := b.runtime
				n.mu.Lock()
				db := n.database
				n.mu.Unlock()
				if db != nil {
					_, _ = b.client.Sync(ctx, db)
					n.mu.Lock()
					if n.serviceManager != nil {
						n.serviceManager.notify()
					}
					n.mu.Unlock()
				}
				n.RequestSync()
			}
			return err
		}
		return nil
	}
	return errors.New("device unavailable")
}

// Bootstrap the local hostname before making any signed network request.
// The worker route, when present, publishes the alias for this same fingerprint.
func (n *Runtime) localRenamePage(r domain.ClusterRoute) (domain.ClusterPage, error) {
	d := n.LocalClusterDevice()
	if d.ID == "" || d.ID != r.Target {
		return domain.ClusterPage{}, errors.New("local device identity unavailable")
	}
	save := func(ctx context.Context, value string) error {
		if n.LocalClusterDevice().ID != r.Target {
			return errors.New("local device identity changed")
		}
		err := n.Keys.SetHostname(value)
		if err != nil {
			return err
		}
		n.mu.Lock()
		if n.serviceManager != nil {
			n.serviceManager.notify()
		}
		n.mu.Unlock()
		defer n.RequestSync()
		if r.Worker == "" {
			return nil
		}
		remote := r
		remote.Page = "rename"
		page, err := n.ClusterPage(ctx, remote)
		if err == nil {
			if len(page.Options) != 1 || page.Options[0].Run == nil {
				return errors.New("name saved locally, cluster rename unavailable")
			}
			name, _ := n.Keys.Hostname()
			err = page.Options[0].Run(ctx, name)
		}
		if err != nil {
			return fmt.Errorf("name saved on this device; retry to publish the cluster name: %w", err)
		}
		return nil
	}
	return domain.ClusterPage{Title: "Rename " + d.DisplayName(""), Detail: "Choose a hostname for this installation.", Options: []domain.ClusterOption{{Label: "Save name", Input: "Device hostname", Run: save}}}, nil
}
