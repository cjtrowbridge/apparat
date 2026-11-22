// Accepts first-owner claims over TLS at a private Unix socket behind scoped Linux network grants.
// The external listener stays under the existing interface/network supervisor.
package httpapi

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func workerClaims(db *persistence.DB, keys *identity.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		pub := keys.Public()
		if len(pub) == 0 {
			respond(w, 503, map[string]string{"error": "worker_locked"})
			return
		}
		if r.URL.Path == "/v1/public-key" && r.Method == "GET" {
			respond(w, 200, map[string]any{"kind": "device", "public_key": pub, "fingerprint": identity.Fingerprint(pub)})
			return
		}
		if r.URL.Path != "/v1/claim" {
			respond(w, 401, map[string]string{"error": "authentication_required"})
			return
		}
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			respond(w, 405, map[string]string{"error": "method_not_allowed"})
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			respond(w, 415, map[string]string{"error": "expected_json"})
			return
		}
		var claim identity.Claim
		in := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		in.DisallowUnknownFields()
		if in.Decode(&claim) != nil || in.Decode(new(any)) != io.EOF {
			respond(w, 400, map[string]string{"error": "invalid_claim"})
			return
		}
		err := db.ClaimWorker(r.Context(), claim, pub)
		if err != nil && !db.ConfirmedClaim(r.Context(), claim) {
			respond(w, 403, map[string]string{"error": "claim_denied"})
			return
		}
		respond(w, 201, map[string]string{"cluster": claim.Cluster, "worker": claim.Worker, "user": identity.Fingerprint(claim.Profile.Profile.Public)})
	})
}

func ListenWorkerClaims(path string, db *persistence.DB, keys *identity.Session) (*Server, error) {
	info, err := os.Lstat(path)
	if err == nil && info.Mode()&os.ModeSocket != 0 {
		err = os.Remove(path)
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	err = os.Chmod(path, 0600)
	if err != nil {
		_ = l.Close()
		return nil, err
	}
	config := &tls.Config{MinVersion: tls.VersionTLS13, GetCertificate: func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := workerCertificate(keys)
		return &cert, err
	}}
	return serveWorkerClaims(l, config, db, keys), nil
}

func serveWorkerClaims(l net.Listener, config *tls.Config, db *persistence.DB, keys *identity.Session) *Server {
	server := &http.Server{Handler: workerClaims(db, keys), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, MaxHeaderBytes: 8192}
	server.SetKeepAlivesEnabled(false)
	s := &Server{server: server, listener: l, done: make(chan error, 1)}
	go func() {
		s.done <- server.Serve(tls.NewListener(l, config))
	}()
	return s
}
