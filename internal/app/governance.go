// Presents current worker policy and submits a signed, revision-checked administrator edit.
// The worker rechecks current privileges and the ordering authority for every change.
package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"strconv"
)

func (n *Runtime) ManageWorker(ctx context.Context, id string, dialog identity.Dialog) {
	if !n.claimMu.TryLock() {
		return
	}
	defer n.claimMu.Unlock()
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return
	}
	for _, w := range rows {
		if w.Fingerprint != id {
			continue
		}
		c, err := n.openSavedWorker(ctx, w)
		if err != nil {
			nativeQuestion(dialog, "notice", w.Host, "Could not reach the saved worker.")
			return
		}
		v, err := c.Governance(ctx)
		if err != nil {
			nativeQuestion(dialog, "notice", w.Host, "The worker requires current administrator access.")
			return
		}
		edit, ok := n.policyDialog(v, dialog)
		if !ok {
			return
		}
		_, ok = nativeQuestion(dialog, "choose", "Confirm policy change", fmt.Sprintf("%s · %s\n%s %s %s %s", w.Host, edit.Kind, edit.Pool, edit.Target, edit.Service, edit.Value), "Apply change")
		if !ok {
			return
		}
		_, err = c.ChangePolicy(ctx, n.Personal, n.Keys, edit, v.Head.Hash)
		if err != nil {
			nativeQuestion(dialog, "notice", w.Host, "The worker declined this change. Reopen administration to refresh its policy.")
			return
		}
		n.RequestSync()
		nativeQuestion(dialog, "notice", w.Host, "The worker committed the policy change.")
		return
	}
}

func (n *Runtime) policyDialog(v persistence.GovernanceView, dialog identity.Dialog) (identity.PolicyEdit, bool) {
	var id [16]byte
	rand.Read(id[:])
	edit := identity.PolicyEdit{ID: hex.EncodeToString(id[:]), Cluster: v.Cluster, Base: v.Head.Revision, Epoch: &v.Epoch}
	ask := func(title, msg string, opts ...string) (string, bool) {
		if len(opts) == 0 {
			return "", false
		}
		return nativeQuestion(dialog, "choose", title, msg, opts...)
	}
	if v.Authority != v.Worker {
		if v.Owner != n.Personal.Fingerprint() {
			nativeQuestion(dialog, "notice", "Policy ordering", "The owner can enable worker ordering from the current authority device.")
			return edit, false
		}
		edit.Kind = "order_on_worker"
		msg := "Allow this worker to order signed changes from your approved administrators?"
		label := "Enable worker ordering"
		if v.Authority != identity.Fingerprint(n.Keys.Public()) {
			edit.Kind = "recover_order"
			msg = "Use your personal owner identity to retire the old ordering device? The worker will reject new orders from the previous authority."
			label = "Recover lost authority"
		}
		_, ok := ask("Policy ordering", msg, label)
		edit.Target = v.Worker
		edit.Value = v.Owner
		edit.Service = v.Authority
		return edit, ok
	}
	choices := []string{"Registration", "Sync interval", "User permissions", "Create group", "Create pool", "Drain and check for update"}
	if len(v.Groups) > 0 {
		choices = append(choices, "Group members")
	}
	if len(v.Pools) > 0 {
		choices = append(choices, "Pool services", "Pool access", "Pool task quota")
	}
	op, ok := ask("Worker administration", fmt.Sprintf("Policy revision %d · %d pools · %d users", v.Head.Revision, len(v.Pools), len(v.Users)), choices...)
	if !ok {
		return edit, false
	}
	switch op {
	case "Drain and check for update":
		edit.Kind = "update_worker"
		edit.Target = v.Worker
	case "Registration":
		edit.Kind = "admission_open"
		edit.Value = "1"
		label := "Open registration"
		if v.RegistrationOpen {
			edit.Value = "0"
			label = "Close registration"
		}
		_, ok = ask("Registration", "This setting controls whether new users may register.", label)
	case "Sync interval":
		edit.Kind = "poll_seconds"
		edit.Value, ok = nativeQuestion(dialog, "text", "Sync interval", "Seconds between checks. Currently "+strconv.Itoa(v.PollSeconds)+".")
	case "Create group":
		edit.Kind = "add_group"
		edit.Pool, ok = nativeQuestion(dialog, "text", "Create group", "Enter a unique group name.")
	case "Group members":
		edit.Kind = "group_member"
		edit.Pool, ok = ask(op, "Select a group.", v.Groups...)
		if !ok {
			return edit, false
		}
		edit.Target, ok = selectPrincipal(dialog, v.Users, false, nil)
		edit.Value = "add"
		for _, m := range v.Memberships {
			if m.Name == edit.Pool && m.User == edit.Target {
				edit.Value = "remove"
			}
		}
	case "Create pool":
		edit.Kind = "add_pool"
		edit.Pool, ok = nativeQuestion(dialog, "text", "Create pool", "Enter a unique pool name.")
	case "Pool task quota":
		edit.Kind = "pool_quota"
		edit.Pool, ok = ask(op, "Choose a pool.", v.Pools...)
		if ok {
			edit.Value, ok = nativeQuestion(dialog, "text", op, "Maximum queued and running tasks per user (1 to 4096). The default is 64.")
		}
	case "User permissions":
		edit.Kind = "user_role"
		edit.Target, ok = selectPrincipal(dialog, v.Users, false, nil)
		if ok {
			edit.Value, ok = ask("User permissions", "Choose this user's access level.", "approved", "admin", "blocked", "pending")
		}
	case "Pool services", "Pool access":
		edit.Pool, ok = ask(op, "Select a pool.", v.Pools...)
		if !ok {
			return edit, false
		}
		edit.Service = persistence.HashService
		edit.Kind = "pool_service"
		edit.Target = v.Worker
		entries := v.Services
		if op == "Pool access" {
			edit.Kind = "pool_grant"
			edit.Target, ok = selectPrincipal(dialog, v.Users, true, v.Groups)
			entries = v.Grants
		}
		if !ok {
			return edit, false
		}
		edit.Value = "add"
		for _, m := range entries {
			if m.Pool == edit.Pool && m.Service == edit.Service && m.Principal == edit.Target {
				edit.Value = "remove"
			}
		}
	default:
		return edit, false
	}
	return edit, ok
}

func selectPrincipal(dialog identity.Dialog, members []persistence.Member, public bool, groups []string) (string, bool) {
	labels := []string{}
	ids := map[string]string{}
	if public {
		labels = append(labels, "Public")
		ids["Public"] = "public"
	}
	for _, group := range groups {
		label := "Group · " + group
		labels = append(labels, label)
		ids[label] = "group:" + group
	}
	for _, m := range members {
		role := m.Admission
		if m.Admin {
			role = "admin"
		}
		label := fmt.Sprintf("%s (%s) · %s", m.Name, role, m.ID)
		labels = append(labels, label)
		ids[label] = m.ID
	}
	choice, ok := nativeQuestion(dialog, "choose", "Select user", "Select the identity for this change.", labels...)
	id, found := ids[choice]
	return id, ok && found
}
