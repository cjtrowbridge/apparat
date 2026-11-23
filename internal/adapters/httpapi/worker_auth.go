// Maintains bounded one-use login challenges and short-lived worker bearer sessions.
// Every protected request rechecks current SQLite admission, device and role state.
package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/adapters/telemetry"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type credential struct {
	access persistence.Access
	until  time.Time
}
type workerAuth struct {
	metrics *telemetry.Ring
	sync.Mutex
	db         *persistence.DB
	workspace  OwnerWorkspace
	git        gitJobs
	gitContext context.Context
	artifacts  string
	challenges map[string]time.Time
	tokens     map[string]credential
}

func newWorkerAuth(db *persistence.DB, artifacts string) *workerAuth {
	return &workerAuth{gitContext: context.Background(), db: db, artifacts: artifacts, challenges: map[string]time.Time{}, tokens: map[string]credential{}}
}

func nonce() string {
	var bytes [32]byte
	rand.Read(bytes[:])
	return base64.RawURLEncoding.EncodeToString(bytes[:])
}

func (a *workerAuth) challenge() string {
	a.Lock()
	defer a.Unlock()
	now := time.Now()
	for key, until := range a.challenges {
		if !now.Before(until) {
			delete(a.challenges, key)
		}
	}
	if len(a.challenges) >= 256 {
		return ""
	}
	key := nonce()
	a.challenges[key] = now.Add(time.Minute)
	return key
}

func (a *workerAuth) consume(key string) bool {
	a.Lock()
	defer a.Unlock()
	until := a.challenges[key]
	delete(a.challenges, key)
	return time.Now().Before(until)
}

func (a *workerAuth) issue(access persistence.Access) string {
	a.Lock()
	defer a.Unlock()
	now := time.Now()
	for key, c := range a.tokens {
		if !now.Before(c.until) {
			delete(a.tokens, key)
		}
	}
	if len(a.tokens) >= 512 {
		return ""
	}
	key := nonce()
	a.tokens[identity.ClaimDigest(key)] = credential{access, now.Add(5 * time.Minute)}
	return key
}

func boundedJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Header.Get("Content-Type") != "application/json" {
		return false
	}
	in := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
	in.DisallowUnknownFields()
	return in.Decode(v) == nil && in.Decode(new(any)) == io.EOF
}

func (a *workerAuth) route(w http.ResponseWriter, r *http.Request, keys *identity.Session) bool {
	worker := identity.Fingerprint(keys.Public())
	switch r.URL.Path {
	case "/v1/auth/challenge":
		if r.Method != http.MethodGet {
			respond(w, 405, map[string]string{"error": "method_not_allowed"})
			return true
		}
		key := a.challenge()
		if key == "" {
			respond(w, 429, map[string]string{"error": "busy"})
		} else {
			respond(w, 200, map[string]string{"worker": worker, "challenge": key, "directory": "1"})
		}
	case "/v1/auth":
		var l identity.Login
		if r.Method != http.MethodPost || !boundedJSON(w, r, &l) || l.Worker != worker || !a.consume(l.Challenge) || r.TLS == nil || verifyPeerChain(*r.TLS, identity.Fingerprint(l.Device.Binding.Device), x509.ExtKeyUsageClientAuth) != nil {
			respond(w, 403, map[string]string{"error": "login_denied"})
			return true
		}
		access, err := a.db.Login(r.Context(), l)
		if err != nil {
			respond(w, 403, map[string]string{"error": "login_denied"})
			return true
		}
		token := ""
		if access.Admission == "approved" {
			token = a.issue(access)
			if token == "" {
				respond(w, 429, map[string]string{"error": "busy"})
				return true
			}
		}
		respond(w, 200, map[string]any{"access": access, "token": token})
	case "/v1/workspace/catalog", "/v1/workspace/live", "/v1/workspace", "/v1/admin/git/projects", "/v1/admin/git/providers", "/v1/admin/git/operation", "/v1/admin/git/accounts", "/v1/sync", "/v1/admin/view", "/v1/tasks/submit", "/v1/tasks/changes", "/v1/tasks/cancel", "/v1/tasks/result", "/v1/me", "/v1/admin/propose", "/v1/admin/order-on-worker", "/v1/admin/policy":
		raw := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(raw, "Bearer ")
		a.Lock()
		record := a.tokens[identity.ClaimDigest(token)]
		a.Unlock()
		if !ok || !time.Now().Before(record.until) || r.TLS == nil || verifyPeerChain(*r.TLS, record.access.Device, x509.ExtKeyUsageClientAuth) != nil {
			respond(w, 401, map[string]string{"error": "authentication_required"})
			return true
		}
		access, err := a.db.CurrentAccess(r.Context(), record.access.Cluster, record.access.User, record.access.Device)
		if err != nil {
			respond(w, 403, map[string]string{"error": "access_denied"})
			return true
		}
		if r.URL.Path == "/v1/workspace/catalog" {
			a.catalog(w, r, access, worker)
			return true
		}
		if r.URL.Path == "/v1/workspace/live" {
			a.live(w, r, record, worker)
			return true
		}
		if r.URL.Path == "/v1/workspace" {
			a.ownerWorkspace(w, r, access, worker)
			return true
		}
		if strings.HasPrefix(r.URL.Path, "/v1/admin/git/") {
			w.Header().Set("Cache-Control", "no-store")
			if r.TLS.Version < tls.VersionTLS13 || !access.Admin || !a.db.OwnerConfirmed(r.Context(), worker) {
				respond(w, 403, map[string]string{"error": "access_denied"})
				return true
			}
			if r.URL.Path != "/v1/admin/git/accounts" {
				a.gitRoute(w, r, access, worker)
				return true
			}
			if r.Method != http.MethodGet {
				respond(w, 405, map[string]string{"error": "method_not_allowed"})
				return true
			}
			names, err := projectgit.SavedAccounts()
			if err != nil {
				respond(w, 503, map[string]string{"error": "accounts_unavailable"})
				return true
			}
			respond(w, 200, map[string]any{"usernames": names})
			return true
		}
		if r.URL.Path == "/v1/sync" {
			a.reconcile(w, r, access, keys)
			return true
		}
		if r.URL.Path == "/v1/admin/view" {
			if r.Method != http.MethodGet || !access.Admin {
				respond(w, 403, map[string]string{"error": "admin_required"})
				return true
			}
			view, err := a.db.Governance(r.Context(), access)
			if err != nil {
				respond(w, 403, map[string]string{"error": "access_denied"})
				return true
			}
			respond(w, 200, view)
			return true
		}
		if strings.HasPrefix(r.URL.Path, "/v1/tasks/") {
			a.tasks(w, r, access, keys)
			return true
		}
		if r.URL.Path == "/v1/me" && r.Method == http.MethodGet {
			respond(w, 200, access)
			return true
		}
		if r.URL.Path == "/v1/admin/propose" || r.URL.Path == "/v1/admin/order-on-worker" {
			a.admin(w, r, access, keys)
			return true
		}
		var cursor persistence.PolicyCursor
		if r.URL.Path != "/v1/admin/policy" || r.Method != http.MethodPost || !access.Admin || !boundedJSON(w, r, &cursor) {
			respond(w, 403, map[string]string{"error": "access_denied"})
			return true
		}
		page, err := a.db.AdminPolicyPage(r.Context(), access.User, access.Device, access.Cluster, cursor)
		if err != nil {
			respond(w, 409, map[string]string{"error": "invalid_policy_cursor"})
		} else {
			respond(w, 200, page)
		}
	default:
		return false
	}
	return true
}
