// Exercises the worker's pinned TLS listener with two-key claims and exact retries.
// Uses an isolated loopback listener in place of the grant-enforcing Linux ingress.
package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"runtime"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func TestPinnedWorkerClaim(t *testing.T) {
	ctx := context.Background()
	worker := identity.NewSession(t.TempDir())
	client := identity.NewSession(t.TempDir())
	for _, session := range []*identity.Session{worker, client} {
		err := session.Install(nil, "", "plain", "", "", true)
		if err != nil {
			t.Fatal(err)
		}
	}
	owner := identity.NewPersonal(t.TempDir())
	err := owner.Create("Fixture", []byte("test password"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := persistence.Open(ctx, t.TempDir()+"/worker.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	cert, err := workerCertificate(worker)
	if err != nil {
		t.Fatal(err)
	}
	server := serveWorkerClaims(listener, &tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequestClientCert, Certificates: []tls.Certificate{cert}}, db, worker, t.TempDir())
	defer server.Close()
	address := listener.Addr().String()
	pin := identity.Fingerprint(worker.Public())
	_, err = OpenWorker(ctx, address, identity.Fingerprint(client.Public()), owner, client)
	if err == nil {
		t.Fatal("accepted incorrect saved fingerprint")
	}
	session, err := OpenWorker(ctx, address, "", owner, client)
	if err != nil {
		t.Fatal(err)
	}
	access, err := session.WhoAmI(ctx)
	if err != nil || !access.Admin || session.Worker != pin {
		t.Fatal("owner access failed", err)
	}
	cluster := access.Cluster
	phone := identity.NewSession(t.TempDir())
	err = phone.Install(nil, "", "plain", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	other, err := OpenWorker(ctx, address, pin, owner, phone)
	if err != nil || !other.Access.Admin || other.Access.Device == access.Device {
		t.Fatal("same user could not authenticate on a distinct device", err)
	}
	guest := identity.NewPersonal(t.TempDir())
	err = guest.Create("Fixture", []byte("guest password"))
	if err != nil {
		t.Fatal(err)
	}
	guestDevice := identity.NewSession(t.TempDir())
	err = guestDevice.Install(nil, "", "plain", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	unapproved, err := OpenWorker(ctx, address, pin, guest, guestDevice)
	if err != nil || unapproved.Access.Admission != "pending" {
		t.Fatal("guest registration should be pending", err)
	}
	_, err = unapproved.WhoAmI(ctx)
	if err == nil {
		t.Fatal("unapproved user accessed protected endpoint")
	}
	edit := identity.PolicyEdit{Cluster: cluster, ID: "handoff", Kind: "order_on_worker", Value: access.User, Service: access.Device, Target: pin}
	handoff, err := session.ChangePolicy(ctx, owner, client, edit, "")
	if err != nil {
		t.Fatal(err)
	}
	edit = identity.PolicyEdit{Cluster: cluster, ID: "promote", Base: 1, Kind: "user_role", Target: unapproved.Access.User, Value: "admin"}
	promotion, err := other.ChangePolicy(ctx, owner, phone, edit, handoff.Hash())
	if err != nil {
		t.Fatal(err)
	}
	secondAdmin, err := OpenWorker(ctx, address, pin, guest, guestDevice)
	if err != nil || !secondAdmin.Access.Admin {
		t.Fatal("second admin not approved", err)
	}
	edit = identity.PolicyEdit{Cluster: cluster, ID: "poll", Base: 2, Kind: "poll_seconds", Value: "240"}
	_, err = secondAdmin.ChangePolicy(ctx, guest, guestDevice, edit, promotion.Hash())
	if err != nil {
		t.Fatal("second admin could not change policy", err)
	}
	page, err := other.Policy(ctx, persistence.PolicyCursor{})
	if err != nil || page.Head.Revision != 3 || len(page.Events) != 3 {
		t.Fatal("missing policy history", err)
	}
	replicaDB, err := persistence.Open(ctx, t.TempDir()+"/replica.db")
	if err != nil {
		t.Fatal(err)
	}
	defer replicaDB.Close()
	err = replicaDB.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	seconds, err := other.Sync(ctx, replicaDB)
	if err != nil || seconds != 0 {
		t.Fatal("initial replica snapshot", seconds, err)
	}
	seconds, err = other.Sync(ctx, replicaDB)
	if err == nil && seconds == 0 {
		seconds, err = other.Sync(ctx, replicaDB)
	}
	if err != nil || seconds != 240 {
		t.Fatal("unchanged policy response", seconds, err)
	}
	cached, cursor, err := replicaDB.LocalReplica(ctx, other.Worker, other.Access)
	if err != nil || !cached.Valid(other.Worker, other.Access) || cursor.Revision == 0 || cached.View.Governance.Head != page.Head {
		t.Fatal("invalid restored governance", err)
	}
	cached.View.PollSeconds++
	err = replicaDB.ApplyReplica(ctx, other.Worker, other.Access, cursor, persistence.SyncPage{Snapshot: &cached})
	if err == nil {
		t.Fatal("accepted altered snapshot")
	}
	for _, change := range []identity.PolicyEdit{
		{Kind: "add_pool", Pool: "interactive"},
		{Kind: "pool_service", Pool: "interactive", Target: session.Worker, Service: persistence.HashService, Value: "add"},
	} {
		change.Base = page.Head.Revision
		change.Cluster = cluster
		change.ID = "queue-" + change.Kind
		event, err := session.ChangePolicy(ctx, owner, client, change, page.Head.Hash)
		if err != nil {
			t.Fatal(err)
		}
		page.Head = persistence.PolicyCursor{Revision: change.Base + 1, Hash: event.Hash()}
	}
	err = other.SubmitHash(ctx, "interactive", "sync-test", "Apparat")
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		_, err = other.Sync(ctx, replicaDB)
		if err != nil {
			t.Fatal(err)
		}
	}
	_, cursor, err = replicaDB.LocalReplica(ctx, other.Worker, other.Access)
	if err != nil || cursor.Queues["interactive"] != 3 {
		t.Fatal("missing submit, lease and result transitions", err, cursor)
	}
	var reply map[string]json.RawMessage
	err = other.call(ctx, "POST", "/v1/sync", cursor, &reply)
	if err != nil || len(reply) != 0 {
		t.Fatal("unchanged sync disclosed payload", err)
	}
	_, err = other.Result(ctx, "interactive", "sync-test")
	if err != nil {
		t.Fatal("artifact verification failed", err)
	}
	var timings []time.Duration
	done := make(chan error, 1)
	go func() {
		for range 20 {
			_, err := session.Policy(ctx, persistence.PolicyCursor{})
			if err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	for range 20 {
		begin := time.Now()
		err = other.call(ctx, "POST", "/v1/sync", cursor, &reply)
		if err != nil || len(reply) != 0 {
			t.Fatal("unchanged concurrent sync", err)
		}
		timings = append(timings, time.Since(begin))
	}
	err = <-done
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(timings, func(i, j int) bool {
		return timings[i] < timings[j]
	})
	body, _ := json.Marshal(reply)
	t.Logf("unchanged JSON object %d bytes before HTTP encoder newline, new TLS connection per request, concurrent full policy reads: median %v, p95 %v", len(body), timings[10], timings[18])
	for _, value := range []string{"add", "remove"} {
		edit := identity.PolicyEdit{Kind: "gui_pool", ID: "gui-" + value, Cluster: cluster, Base: page.Head.Revision, Pool: "Chat Pool", Target: other.Access.Device, Value: value}
		receipt, err := session.ChangePolicy(ctx, owner, client, edit, page.Head.Hash)
		if err != nil {
			t.Fatal(err)
		}
		page.Head = persistence.PolicyCursor{Revision: edit.Base + 1, Hash: receipt.Hash()}
		_, err = other.Sync(ctx, replicaDB)
		if err != nil {
			t.Fatal(err)
		}
		snapshot, _, err := replicaDB.LocalReplica(ctx, other.Worker, other.Access)
		found := false
		for _, d := range snapshot.View.Devices {
			if d.ID == other.Access.Device {
				found = d.Host != "" && d.OS == runtime.GOOS && slices.Contains(d.Pools, "Chat Pool") == (value == "add")
			}
		}
		if err != nil || !found {
			t.Fatal("device pool selection did not synchronize", err)
		}
	}
	for _, target := range []string{other.Access.Device, session.Worker} {
		edit := identity.PolicyEdit{Kind: "rename_device", ID: "name-" + target, Cluster: cluster, Base: page.Head.Revision, Target: target, Value: "localhost"}
		_, err = session.ChangePolicy(ctx, owner, client, edit, page.Head.Hash)
		if err == nil {
			t.Fatal("accepted localhost")
		}
		edit.Value = "lab-device"
		receipt, err := session.ChangePolicy(ctx, owner, client, edit, page.Head.Hash)
		if err != nil {
			t.Fatal(err)
		}
		page.Head = persistence.PolicyCursor{Revision: edit.Base + 1, Hash: receipt.Hash()}
		other, err = OpenWorker(ctx, address, pin, owner, phone)
		if err != nil {
			t.Fatal(err)
		}
		_, err = other.Sync(ctx, replicaDB)
		if err != nil {
			t.Fatal(err)
		}
		snapshot, _, err := replicaDB.LocalReplica(ctx, other.Worker, other.Access)
		found := false
		for _, d := range snapshot.View.Devices {
			if d.ID == target {
				found = d.Host == edit.Value
			}
		}
		if !found || err != nil {
			t.Fatal("rename lost after login or sync", err)
		}
	}
	for _, target := range []string{other.Access.Device, session.Worker} {
		for _, want := range [][]string{{"Chat Pool", "Research Pool"}, {"Project Pool"}, {}} {
			raw, _ := json.Marshal(want)
			edit := identity.PolicyEdit{Kind: "device_pools", ID: target + string(raw), Target: target, Base: page.Head.Revision, Cluster: cluster, Value: string(raw)}
			receipt, err := session.ChangePolicy(ctx, owner, client, edit, page.Head.Hash)
			if err != nil {
				t.Fatal(err)
			}
			page.Head = persistence.PolicyCursor{Revision: edit.Base + 1, Hash: receipt.Hash()}
			invalid := edit
			invalid.Base = page.Head.Revision
			invalid.ID += "invalid"
			invalid.Value = `["Chat Pool","absent pool"]`
			_, err = session.ChangePolicy(ctx, owner, client, invalid, page.Head.Hash)
			if err == nil {
				t.Fatal("accepted partial selection")
			}
			_, err = other.Sync(ctx, replicaDB)
			if err != nil {
				t.Fatal(err)
			}
			cached, _, err := replicaDB.LocalReplica(ctx, other.Worker, other.Access)
			found := false
			for _, d := range cached.View.Devices {
				if d.ID == target {
					found = slices.Equal(d.Pools, want)
				}
			}
			if err != nil || !found || cached.View.Governance.Head != page.Head {
				t.Fatal("saved set not replicated atomically", err)
			}
		}
	}
	other.token = session.token
	_, err = other.WhoAmI(ctx)
	if err == nil {
		t.Fatal("accepted another device's bearer token")
	}
}
