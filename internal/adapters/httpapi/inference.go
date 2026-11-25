// Transports bounded inference operations over the existing pinned mutual-TLS worker listener.
// User admission uses device-bound authentication. Peer voting checks the agreed in-memory roster.
package httpapi

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type InferenceAPI interface {
	InferenceUser(context.Context, persistence.Access, domain.InferenceCommand) (domain.InferenceReply, error)
	InferencePeer(context.Context, string, domain.InferenceCommand) (domain.InferenceReply, error)
}

func decodeInference(w http.ResponseWriter, r *http.Request, q *domain.InferenceCommand) bool {
	if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
		return false
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512<<10))
	d.DisallowUnknownFields()
	return d.Decode(q) == nil && d.Decode(new(any)) == io.EOF
}
func (a *workerAuth) inferenceRoute(w http.ResponseWriter, r *http.Request, user *persistence.Access) {
	if a.inference == nil {
		respond(w, 503, map[string]string{"error": "inference_unavailable"})
		return
	}
	var q domain.InferenceCommand
	if !decodeInference(w, r, &q) {
		respond(w, 400, map[string]string{"error": "invalid_inference_request"})
		return
	}
	var reply domain.InferenceReply
	var err error
	if user != nil {
		reply, err = a.inference.InferenceUser(r.Context(), *user, q)
	} else {
		if r.TLS == nil || len(r.TLS.PeerCertificates) != 2 {
			respond(w, 403, map[string]string{"error": "peer_credentials_required"})
			return
		}
		pub, ok := r.TLS.PeerCertificates[1].PublicKey.(ed25519.PublicKey)
		if !ok || verifyPeerChain(*r.TLS, identity.Fingerprint(pub), x509.ExtKeyUsageClientAuth) != nil {
			respond(w, 403, map[string]string{"error": "invalid_peer_identity"})
			return
		}
		reply, err = a.inference.InferencePeer(r.Context(), identity.Fingerprint(pub), q)
	}
	if errors.Is(err, ErrDenied) {
		respond(w, 403, map[string]string{"error": "inference_access_denied"})
		return
	}
	if err != nil {
		respond(w, 409, map[string]string{"error": "inference_operation_rejected"})
		return
	}
	respond(w, 200, reply)
}
func (c *WorkerClient) Inference(ctx context.Context, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	if q.Operation == "global-get" {
		return ReadGlobal(ctx, q.Global.Session, c.Access.User, c.Inference)
	}
	err = c.call(ctx, "POST", "/v1/inference", q, &v)
	if err == nil && (q.Operation == "global-post" || q.Operation == "global-rename" || q.Operation == "global-terminal-result") {
		return ReadGlobal(ctx, q.Global.Session, c.Access.User, c.Inference)
	}
	return
}

type InferenceClient struct {
	client  *http.Client
	address string
}

func NewInferenceClient(p domain.InferencePeer, keys *identity.Session) (*InferenceClient, error) {
	leaf, err := nodeCertificate(keys, x509.ExtKeyUsageClientAuth)
	if err != nil {
		return nil, err
	}
	cfg := claimClientTLS(p.ID)
	cfg.Certificates = []tls.Certificate{leaf}
	return &InferenceClient{address: "https://" + p.Address + "/v1/inference/peer", client: &http.Client{Timeout: 8 * time.Second, Transport: &http.Transport{TLSClientConfig: cfg}, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}, nil
}
func (c *InferenceClient) Call(ctx context.Context, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	b, err := json.Marshal(q)
	if err != nil {
		return v, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.address, bytes.NewReader(b))
	if err != nil {
		return v, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return v, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return v, errors.New("peer rejected inference operation")
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 512<<10)).Decode(&v)
	return
}
