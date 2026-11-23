// Authenticates a personal device with a pinned TLS worker and a one-time signed challenge.
// Short-lived credentials stay in memory and the client never follows a redirect.
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
	"net/netip"
	"runtime"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

var ErrDenied = errors.New("worker denied current access")

type WorkerClient struct {
	http      *http.Client
	directory bool
	base      string
	token     string
	Worker    string
	Access    persistence.Access
}

func OpenWorker(ctx context.Context, addr, pin string, person *identity.Personal, device *identity.Session) (*WorkerClient, error) {
	address, err := netip.ParseAddrPort(addr)
	if err != nil || address.Port() == 0 {
		return nil, errors.New("invalid worker address")
	}
	config := claimClientTLS(pin)
	// On first contact learn the worker key from its TLS chain. Every subsequent
	// handshake, including the signed login, must present that same key.
	if pin == "" {
		config.VerifyConnection = func(state tls.ConnectionState) error {
			if len(state.PeerCertificates) != 2 {
				return errors.New("missing worker identity")
			}
			key, ok := state.PeerCertificates[1].PublicKey.(ed25519.PublicKey)
			if !ok {
				return errors.New("invalid worker key")
			}
			candidate := identity.Fingerprint(key)
			err := verifyPeerChain(state, candidate, x509.ExtKeyUsageServerAuth)
			if err == nil && pin != "" && pin != candidate {
				return errors.New("worker key changed")
			}
			if err == nil && pin == "" {
				pin = candidate
			}
			return err
		}
	}
	leaf, err := nodeCertificate(device, x509.ExtKeyUsageClientAuth)
	if err != nil {
		return nil, err
	}
	config.Certificates = []tls.Certificate{leaf}
	client := &WorkerClient{base: "https://" + address.String(), http: &http.Client{Timeout: 8 * time.Second, Transport: &http.Transport{TLSClientConfig: config, DisableKeepAlives: true}, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
	var challenge struct{ Worker, Challenge, Directory string }
	err = client.call(ctx, "GET", "/v1/auth/challenge", nil, &challenge)
	if err != nil || challenge.Worker != pin {
		return nil, errors.New("worker challenge unavailable")
	}
	platform := ""
	if challenge.Directory == "1" {
		platform = runtime.GOOS
	}
	proof, err := person.SignLogin(device, pin, challenge.Challenge, platform)
	if err != nil {
		return nil, err
	}
	var response struct {
		Access persistence.Access
		Token  string
	}
	err = client.call(ctx, "POST", "/v1/auth", proof, &response)
	if err != nil {
		return nil, err
	}
	if response.Access.User != identity.Fingerprint(proof.Profile.Profile.Public) || response.Access.Device != identity.Fingerprint(proof.Device.Binding.Device) {
		return nil, errors.New("wrong authenticated identity")
	}
	client.directory = challenge.Directory == "1"
	client.Worker = pin
	client.token = response.Token
	client.Access = response.Access
	return client, nil
}

func (c *WorkerClient) call(ctx context.Context, method, path string, input, output any) error {
	var data []byte
	var err error
	if input != nil {
		data, err = json.Marshal(input)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Apparat-Device-Directory", "1")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return errors.New("could not verify or contact the worker")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return ErrDenied
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20+1))
	if err != nil || len(body) > 1<<20 || resp.StatusCode != 200 {
		return errors.New("worker denied the operation or returned an invalid response")
	}
	return json.Unmarshal(body, output)
}

func (c *WorkerClient) WhoAmI(ctx context.Context) (persistence.Access, error) {
	var access persistence.Access
	err := c.call(ctx, "GET", "/v1/me", nil, &access)
	return access, err
}

func (c *WorkerClient) Policy(ctx context.Context, after persistence.PolicyCursor) (persistence.PolicyPage, error) {
	var page persistence.PolicyPage
	err := c.call(ctx, "POST", "/v1/admin/policy", after, &page)
	return page, err
}

func (c *WorkerClient) ChangePolicy(ctx context.Context, p *identity.Personal, d *identity.Session, edit identity.PolicyEdit, previous string) (persistence.PolicyEvent, error) {
	proposal, err := p.Propose(d, edit)
	if err != nil {
		return persistence.PolicyEvent{}, err
	}
	var event persistence.PolicyEvent
	if edit.Kind == "order_on_worker" {
		signed, err := persistence.SignHandoff(proposal, previous, d)
		if err != nil {
			return event, err
		}
		err = c.call(ctx, "POST", "/v1/admin/order-on-worker", signed, &event)
		return event, err
	}
	err = c.call(ctx, "POST", "/v1/admin/propose", proposal, &event)
	return event, err
}

func (c *WorkerClient) Governance(ctx context.Context) (persistence.GovernanceView, error) {
	var v persistence.GovernanceView
	err := c.call(ctx, "GET", "/v1/admin/view", nil, &v)
	if err == nil && (v.Worker != c.Worker || v.Cluster != c.Access.Cluster) {
		err = errors.New("wrong worker governance view")
	}
	return v, err
}
