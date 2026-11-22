// Exercises the worker's pinned TLS listener with two-key claims and exact retries.
// Uses an isolated loopback listener in place of the grant-enforcing Linux ingress.
package httpapi

import (
	"context"
	"crypto/tls"
	"net"
	"testing"

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
	other.token = session.token
	_, err = other.WhoAmI(ctx)
	if err == nil {
		t.Fatal("accepted another device's bearer token")
	}
}
