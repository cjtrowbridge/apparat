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
	code, err := db.IssueWorkerInvitation(ctx, worker.Public())
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
	server := serveWorkerClaims(listener, &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{cert}}, db, worker)
	defer server.Close()
	invite := Invitation{"fixture", listener.Addr().String(), identity.Fingerprint(worker.Public()), code}
	cluster := identity.ClaimDigest("test cluster")
	wrong := invite
	wrong.Fingerprint = identity.Fingerprint(client.Public())
	bad, err := owner.SignClaim(client, cluster, wrong.Fingerprint, code)
	if err != nil || PostClaim(ctx, wrong, bad) == nil {
		t.Fatal("incorrect worker fingerprint accepted or invalid test setup", err)
	}
	ok, err := owner.SignClaim(client, cluster, invite.Fingerprint, code)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		err = PostClaim(ctx, invite, ok)
		if err != nil {
			t.Fatal(err)
		}
	}
}
