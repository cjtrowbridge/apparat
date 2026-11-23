// Exercises both signature checks and atomic, one-use worker ownership under contention.
// Uses disposable user and device identities with the production migration and claim path.
package persistence

import (
	"context"
	"slices"
	"sync"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

func TestOnlyOneOwner(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, t.TempDir()+"/worker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	device := identity.NewSession(t.TempDir())
	err = device.Install(nil, "", "plain", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	personal := identity.NewPersonal(t.TempDir())
	err = personal.Create("Test owner", []byte("test password"))
	if err != nil {
		t.Fatal(err)
	}
	worker, err := identity.Generate()
	if err != nil {
		t.Fatal(err)
	}
	second := identity.NewPersonal(t.TempDir())
	err = second.Create("Other", []byte("another password"))
	if err != nil {
		t.Fatal(err)
	}
	err = db.ConfirmOwner(ctx, identity.Fingerprint(worker.Public), personal.Fingerprint())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan Access, 2)
	for _, p := range []*identity.Personal{personal, second} {
		proof, err := p.SignLogin(device, identity.Fingerprint(worker.Public), identity.ClaimDigest("test"))
		if err != nil {
			t.Fatal(err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a, err := db.Login(ctx, proof)
			if err != nil {
				if p == personal {
					t.Error(err)
				}
				return
			}
			results <- a
		}()
	}
	wg.Wait()
	close(results)
	admins := 0
	for a := range results {
		if a.Admin {
			admins++
			if a.User != personal.Fingerprint() {
				t.Fatal("wrong claim owner")
			}
			v, err := db.Governance(ctx, a)
			if err != nil || !slices.Equal(v.Pools, []string{"Chat Pool", "Comrade Pool", "Project Pool", "Research Pool"}) || len(v.Services) != 0 || len(v.Grants) != 0 {
				t.Fatalf("unexpected initial pools or permissions: %v, %v", v.Pools, err)
			}
			profile, _, err := personal.Credentials(device.Public())
			if err != nil {
				t.Fatal(err)
			}
			p := personal
			if a.User != identity.Fingerprint(profile.Profile.Public) {
				p = second
			}
			verifyPolicy(ctx, t, db, p, device, a.Cluster)
		} else if a.Admission != "pending" {
			t.Fatal("expected pending")
		}
	}
	if admins != 1 {
		t.Fatal("expected exactly one first admin", admins)
	}

}

func verifyPolicy(ctx context.Context, t *testing.T, db *DB, p *identity.Personal, d *identity.Session, cluster string) {
	t.Helper()
	edit := identity.PolicyEdit{ID: "poll", Cluster: cluster, Kind: "poll_seconds", Value: "200"}
	proposal, err := p.Propose(d, edit)
	if err != nil {
		t.Fatal(err)
	}
	event, err := db.CommitPolicy(ctx, proposal, d)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := db.CommitPolicy(ctx, proposal, d)
	if err != nil || retry.Hash() != event.Hash() {
		t.Fatal("idempotent proposal failed", err)
	}
	user := identity.Fingerprint(proposal.Device.Binding.User)
	device := identity.Fingerprint(d.Public())
	page, err := db.AdminPolicyPage(ctx, user, device, cluster, PolicyCursor{})
	if err != nil || len(page.Events) != 1 || page.Head.Hash != event.Hash() {
		t.Fatal("missing policy event", err)
	}
	empty, err := db.AdminPolicyPage(ctx, user, device, cluster, page.Head)
	if err != nil || len(empty.Events) != 0 {
		t.Fatal("unexpected policy events", err)
	}
	_, err = db.AdminPolicyPage(ctx, user, "another device", cluster, page.Head)
	if err == nil {
		t.Fatal("disclosed head to unauthenticated device")
	}
	event.Previous = "changed"
	if db.ApplyPolicy(ctx, event, d.Public()) == nil {
		t.Fatal("accepted altered receipt")
	}
	edit.ID = "demote"
	edit.Kind = "user_role"
	edit.Value = "blocked"
	edit.Base = 1
	edit.Target = user
	proposal, err = p.Propose(d, edit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CommitPolicy(ctx, proposal, d)
	if err == nil {
		t.Fatal("allowed removal of last administrator")
	}
}

func TestLostOrderingDeviceRecovery(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, t.TempDir()+"/worker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	nodes := []*identity.Session{}
	for range 3 {
		keys := identity.NewSession(t.TempDir())
		err = keys.Install(nil, "", "plain", "", "", true)
		if err != nil {
			t.Fatal(err)
		}
		nodes = append(nodes, keys)
	}
	owner := identity.NewPersonal(t.TempDir())
	err = owner.Create("Owner", []byte("regression password"))
	if err != nil {
		t.Fatal(err)
	}
	var access Access
	worker := identity.Fingerprint(nodes[0].Public())
	err = db.ConfirmOwner(ctx, worker, owner.Fingerprint())
	if err != nil {
		t.Fatal(err)
	}
	for _, device := range nodes[1:] {
		login, err := owner.SignLogin(device, worker, identity.ClaimDigest("challenge"))
		if err != nil {
			t.Fatal(err)
		}
		access, err = db.Login(ctx, login)
		if err != nil {
			t.Fatal(err)
		}
	}
	epoch := uint64(0)
	edit := identity.PolicyEdit{ID: "recover", Kind: "recover_order", Cluster: access.Cluster, Epoch: &epoch, Value: access.User, Service: identity.Fingerprint(nodes[1].Public()), Target: worker}
	proposal, err := owner.Propose(nodes[2], edit)
	if err != nil {
		t.Fatal(err)
	}
	event, err := db.CommitPolicy(ctx, proposal, nodes[0])
	if err != nil {
		t.Fatal(err)
	}
	view, err := db.Governance(ctx, access)
	if err != nil || view.Epoch != 1 || view.Authority != worker || identity.Fingerprint(view.OriginalDevice.Binding.Device) != edit.Service || view.Head.Hash != event.Hash() {
		t.Fatal("failed to fence old authority", err)
	}
	edit.ID = "stale"
	edit.Kind = "poll_seconds"
	edit.Value = "200"
	edit.Base = 1
	proposal, err = owner.Propose(nodes[1], edit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CommitPolicy(ctx, proposal, nodes[1])
	if err == nil {
		t.Fatal("retired authority accepted")
	}
	_, err = db.CommitPolicy(ctx, proposal, nodes[0])
	if err == nil {
		t.Fatal("stale epoch accepted")
	}
	epoch = 1
	edit.Kind = "revoke_device"
	edit.ID = "revoke-retired-device"
	edit.Target = identity.Fingerprint(nodes[1].Public())
	proposal, err = owner.Propose(nodes[2], edit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CommitPolicy(ctx, proposal, nodes[0])
	if err != nil {
		t.Fatal("revoking retired device", err)
	}
	_, err = db.CurrentAccess(ctx, access.Cluster, access.User, edit.Target)
	if err == nil {
		t.Fatal("revoked device retained access")
	}

}

func TestApprovedWorkerReplacementPersists(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/node.db"
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a := Access{Cluster: "cluster", User: "owner", Admission: "approved"}
	err = db.SaveWorkerAccess(ctx, "host", "10.0.0.5:9000", "old", a)
	if err != nil {
		t.Fatal(err)
	}
	err = db.ReplaceWorkerPin(ctx, "old", "new", "10.0.0.5:9000", a)
	if err == nil {
		t.Fatal("accepted a non-admin replacement")
	}
	a.Admin = true
	err = db.ReplaceWorkerPin(ctx, "old", "new", "10.0.0.5:9000", a)
	if err != nil {
		t.Fatal(err)
	}
	err = db.SaveWorkerAccess(ctx, "host", "10.0.0.5:9000", "old", a)
	if err == nil {
		t.Fatal("stale login restored obsolete identity")
	}
	reopened, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	pin, err := reopened.WorkerPin(ctx, "alias", "10.0.0.5:9000")
	if err != nil || pin != "new" {
		t.Fatal("new pin missing after reopening database")
	}
}
