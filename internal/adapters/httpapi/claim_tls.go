// Issues short-lived TLS leaf credentials under an unlocked worker signing identity.
// Claim clients validate the chain against the worker fingerprint from the local invitation.
package httpapi

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
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
	leaf := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: claimDNS}, DNSNames: []string{claimDNS}, NotBefore: issuer.NotBefore, NotAfter: issuer.NotAfter, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage}}
	cert, err := x509.CreateCertificate(rand.Reader, leaf, issuer, public, key)
	if err != nil {
		return empty, err
	}
	return tls.Certificate{Certificate: [][]byte{cert, root}, PrivateKey: private}, nil
}

func claimClientTLS(pin string) *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS13, InsecureSkipVerify: true, VerifyConnection: func(state tls.ConnectionState) error {
		return verifyPeerChain(state, pin, x509.ExtKeyUsageServerAuth)
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
