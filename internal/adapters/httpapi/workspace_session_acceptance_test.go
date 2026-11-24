// Checks the authentication boundary shared by file operations, catalog pages and watches.
// Uses valid registered project IDs and TLS leaf chains to isolate each reason for rejection.
package httpapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func TestWorkspaceSessionBoundary(t *testing.T) {
	ctx := t.Context()
	dir := t.TempDir()
	db, err := persistence.Open(ctx, filepath.Join(dir, "db"))
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(err)
	defer db.Close()
	must(db.Migrate(ctx))
	device := func() *identity.Session {
		d := identity.NewSession(t.TempDir())
		must(d.Install(nil, "", "plain", "", "", true))
		return d
	}
	worker := device()
	enrolled := device()
	other := device()
	owner := identity.NewPersonal(t.TempDir())
	must(owner.Create("Owner", []byte("fixture password")))
	pin := identity.Fingerprint(worker.Public())
	must(db.ConfirmOwner(ctx, pin, owner.Fingerprint()))
	proof, err := owner.SignLogin(enrolled, pin, identity.ClaimDigest("fixture"))
	must(err)
	access, err := db.Login(ctx, proof)
	must(err)
	project := "remote-private"
	file := "private-note"
	content := "private contents"
	must(os.WriteFile(filepath.Join(dir, file), []byte(content), 0600))
	must(db.RegisterProject(ctx, domain.Project{ID: project, Name: project, Location: dir}))
	auth := newWorkerAuth(db, dir)
	auth.gitContext = ctx
	auth.workspace = func(context.Context, persistence.Access, domain.WorkspaceRequest) (domain.WorkspaceView, error) {
		return domain.WorkspaceView{}, errors.New("unexpected workspace dispatch")
	}
	chain := func(d *identity.Session) *tls.ConnectionState {
		cert, err := nodeCertificate(d, x509.ExtKeyUsageClientAuth)
		must(err)
		state := &tls.ConnectionState{Version: tls.VersionTLS13}
		for _, der := range cert.Certificate {
			leaf, err := x509.ParseCertificate(der)
			must(err)
			state.PeerCertificates = append(state.PeerCertificates, leaf)
		}
		return state
	}
	valid := chain(enrolled)
	must(verifyPeerChain(*valid, access.Device, x509.ExtKeyUsageClientAuth))
	auth.tokens[identity.ClaimDigest("expired")] = credential{access: access, until: time.Now().Add(-time.Minute)}
	auth.tokens[identity.ClaimDigest("active")] = credential{access: access, until: time.Now().Add(time.Minute)}
	for _, session := range []struct {
		name, token string
		state       *tls.ConnectionState
	}{
		{"anonymous", "", &tls.ConnectionState{Version: tls.VersionTLS13}},
		{"expired", "expired", valid},
		{"wrong-device", "active", chain(other)},
	} {
		for _, op := range []string{"catalog", "list", "read", "save", "live", "inventory", "configure-ollama", "chat-models", "model-live"} {
			t.Run(session.name+"/"+op, func(t *testing.T) {
				path := "/v1/workspace"
				method := "POST"
				q := domain.WorkspaceRequest{ProjectID: project, Path: file, Action: op}
				switch op {
				case "chat-models":
					path = "/v1/chat/models"
				case "model-live":
					path = "/v1/chat/models/live"
					method = "GET"
				case "inventory":
					path = "/v1/inventory"
					method = "GET"
				case "configure-ollama":
					path = "/v1/admin/inventory/ollama"
				case "catalog", "live":
					path += "/" + op
					method = "GET"
				case "list":
					q.Path = "."
				case "save":
					q.Base = projectgit.Digest(content)
					q.Text = "changed"
					q.ID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
				}
				b, err := json.Marshal(q)
				if err != nil {
					t.Fatal(err)
				}
				r := httptest.NewRequest(method, path, bytes.NewReader(b))
				r.Header.Set("Content-Type", "application/json")
				if session.token != "" {
					r.Header.Set("Authorization", "Bearer "+session.token)
				}
				r.TLS = session.state
				w := httptest.NewRecorder()
				handled := auth.route(w, r, worker)
				if !handled || w.Code != 401 || strings.TrimSpace(w.Body.String()) != `{"error":"authentication_required"}` {
					t.Fatalf("expected only authentication_required, status=%d", w.Code)
				}
				data, err := os.ReadFile(filepath.Join(dir, file))
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != content {
					t.Fatal("denied request modified checkout")
				}
			})
		}
	}
}
