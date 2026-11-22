// Posts a signed first claim to the exact IP and pinned TLS identity in a worker invitation.
// Disallows redirects, proxy use and oversized responses so an invite stays with its target.
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"time"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

type Invitation struct {
	Host        string
	Address     string
	Fingerprint string
	Code        string
}

func PostClaim(ctx context.Context, i Invitation, c identity.Claim) error {
	target, err := netip.ParseAddrPort(i.Address)
	if err != nil || !target.Addr().IsValid() || target.Addr().IsUnspecified() || target.Addr().IsMulticast() || target.Port() == 0 || i.Fingerprint != c.Worker || i.Code != c.Invitation || !c.Valid() {
		return errors.New("invalid worker invitation")
	}
	body, err := json.Marshal(c)
	if err != nil {
		return err
	}
	tr := &http.Transport{TLSClientConfig: claimClientTLS(c.Worker), DisableKeepAlives: true}
	defer tr.CloseIdleConnections()
	client := http.Client{Transport: tr, Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+target.String()+"/v1/claim", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("could not reach and verify the invited worker")
	}
	defer resp.Body.Close()
	var result struct{ Cluster, Worker, User string }
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(raw) > 4096 || resp.StatusCode != 201 || json.Unmarshal(raw, &result) != nil || result.Cluster != c.Cluster || result.Worker != c.Worker || result.User != identity.Fingerprint(c.Profile.Profile.Public) {
		return errors.New("worker did not confirm this claim")
	}
	return nil
}
