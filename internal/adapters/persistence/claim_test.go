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
	invite, err := db.IssueWorkerInvitation(ctx, worker.Public)
	if err != nil {
		t.Fatal(err)
	}
	claim, err := personal.SignClaim(device, identity.ClaimDigest("fixture cluster"), identity.Fingerprint(worker.Public), invite)
	if err != nil {
		t.Fatal(err)
	}
	err = db.RememberDevice(ctx, claim.Profile, claim.Device)
	if err != nil {
		t.Fatal(err)
	}
	bad := claim
	bad.Cluster = identity.ClaimDigest("other cluster")
	if db.ClaimWorker(ctx, bad, worker.Public) == nil {
		t.Fatal("accepted changed cluster")
	}
	var wg sync.WaitGroup
	outcomes := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			outcomes <- db.ClaimWorker(ctx, claim, worker.Public)
		}()
	}
	wg.Wait()
	close(outcomes)
	count := 0
	for err := range outcomes {
		if err == nil {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("successful claims: %d", count)
	}
	_, err = db.IssueWorkerInvitation(ctx, worker.Public)
	if err == nil {
		t.Fatal("issued a new invitation on an owned worker")
	}
	var owner string
	err = db.sql.QueryRowContext(ctx, "SELECT user_id FROM cluster_users WHERE admission='approved' AND admin=1").Scan(&owner)
	if err != nil || owner != identity.Fingerprint(claim.Profile.Profile.Public) {
		t.Fatal("missing approved owner", err)
	}
	verifyPolicy(ctx, t, db, personal, device, claim.Cluster)
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
