// Exercises both signature checks and atomic, one-use worker ownership under contention.
// Uses disposable user and device identities with the production migration and claim path.
package persistence

import (
	"context"
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
				t.Error(err)
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
