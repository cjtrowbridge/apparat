// Accepts first-owner claims over TLS at a private Unix socket behind scoped Linux network grants.
// The external listener stays under the existing interface/network supervisor.
package httpapi

import (
	"crypto/tls"

	"net"
	"net/http"
	"os"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func workerClaims(db *persistence.DB, keys *identity.Session, artifacts string) http.Handler {
	auth := newWorkerAuth(db, artifacts)
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
		if auth.route(w, r, keys) {
			return
		}
		respond(w, 401, map[string]string{"error": "authentication_required"})
	})
}

func ListenWorkerClaims(path, artifacts string, db *persistence.DB, keys *identity.Session) (*Server, error) {
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
	config := &tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequestClientCert, GetCertificate: func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := workerCertificate(keys)
		return &cert, err
	}}
	return serveWorkerClaims(l, config, db, keys, artifacts), nil
}

func serveWorkerClaims(l net.Listener, config *tls.Config, db *persistence.DB, keys *identity.Session, artifacts string) *Server {
	server := &http.Server{Handler: workerClaims(db, keys, artifacts), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, MaxHeaderBytes: 8192}
	server.SetKeepAlivesEnabled(false)
	s := &Server{server: server, listener: l, done: make(chan error, 1)}
	go func() {
		s.done <- server.Serve(tls.NewListener(l, config))
	}()
	return s
}
