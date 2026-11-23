// Issues short-lived TLS leaf credentials under an unlocked worker signing identity.
// Claim clients validate the chain against the worker fingerprint from the local invitation.
package httpapi

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
	"net/netip"
	"time"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const claimDNS = "worker.apparat.invalid"

type nodeSigner struct{ keys *identity.Session }

func (s nodeSigner) Public() crypto.PublicKey {
	return s.keys.Public()
}

func (s nodeSigner) Sign(_ io.Reader, message []byte, opts crypto.SignerOpts) ([]byte, error) {
	if opts.HashFunc() != crypto.Hash(0) {
		return nil, errors.New("unsupported node signature")
	}
	return s.keys.Sign(message)
}

func workerCertificate(keys *identity.Session) (tls.Certificate, error) {
	return nodeCertificate(keys, x509.ExtKeyUsageServerAuth)
}

func nodeCertificate(keys *identity.Session, usage x509.ExtKeyUsage) (tls.Certificate, error) {
	return namedCertificate(keys, usage, pkix.Name{CommonName: claimDNS})
}

func namedCertificate(keys *identity.Session, usage x509.ExtKeyUsage, subject pkix.Name) (tls.Certificate, error) {
	var empty tls.Certificate
	now := time.Now()
	issuer := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Apparat worker identity"}, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(24 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	key := nodeSigner{keys}
	root, err := x509.CreateCertificate(rand.Reader, issuer, issuer, key.Public(), key)
	if err != nil {
		return empty, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return empty, err
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return empty, err
	}
	leaf := &x509.Certificate{SerialNumber: serial, Subject: subject, DNSNames: []string{claimDNS}, NotBefore: issuer.NotBefore, NotAfter: issuer.NotAfter, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage}}
	cert, err := x509.CreateCertificate(rand.Reader, leaf, issuer, public, key)
	if err != nil {
		return empty, err
	}
	return tls.Certificate{Certificate: [][]byte{cert, root}, PrivateKey: private}, nil
}

func claimClientTLS(pin string) *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS13, InsecureSkipVerify: true, VerifyConnection: func(state tls.ConnectionState) error {
		key, err := serverFingerprint(state)
		if err != nil {
			return err
		}
		if key != pin {
			return &ChangedWorker{Expected: pin, Presented: key}
		}
		return nil
	}}
}

func verifyPeerChain(state tls.ConnectionState, pin string, usage x509.ExtKeyUsage) error {
	if len(state.PeerCertificates) != 2 {
		return errors.New("worker certificate chain missing")
	}
	root := state.PeerCertificates[1]
	public, ok := root.PublicKey.(ed25519.PublicKey)
	if !ok || identity.Fingerprint(public) != pin || root.CheckSignatureFrom(root) != nil {
		return errors.New("worker fingerprint mismatch")
	}
	pool := x509.NewCertPool()
	pool.AddCert(root)
	_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{Roots: pool, DNSName: claimDNS, KeyUsages: []x509.ExtKeyUsage{usage}})
	return err
}

// ChangedWorker is returned only for a valid chain signed by a different identity.
type ChangedWorker struct{ Expected, Presented, Address string }

func (e *ChangedWorker) Error() string {
	return "worker identity has changed"
}

func serverFingerprint(state tls.ConnectionState) (string, error) {
	if len(state.PeerCertificates) != 2 {
		return "", errors.New("missing worker identity")
	}
	pub, ok := state.PeerCertificates[1].PublicKey.(ed25519.PublicKey)
	if !ok {
		return "", errors.New("unsupported worker identity")
	}
	pin := identity.Fingerprint(pub)
	return pin, verifyPeerChain(state, pin, x509.ExtKeyUsageServerAuth)
}

// ProbeWorker verifies the presented chain without sending client credentials or HTTP.
func ProbeWorker(ctx context.Context, address string) (string, error) {
	addr, err := netip.ParseAddrPort(address)
	if err != nil || addr.Port() == 0 {
		return "", errors.New("invalid worker address")
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	var pin string
	config := &tls.Config{MinVersion: tls.VersionTLS13, InsecureSkipVerify: true, VerifyConnection: func(state tls.ConnectionState) error {
		var err error
		pin, err = serverFingerprint(state)
		return err
	}}
	dialer := tls.Dialer{Config: config}
	conn, err := dialer.DialContext(ctx, "tcp", addr.String())
	if err != nil {
		return "", err
	}
	conn.Close()
	return pin, nil
}
