// Confirms two locally entered codes on a single TLS connection before transfer.
// Each call owns one attempt. Callers open sessions only after local confirmation.
package httpapi

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"math/big"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/bytemare/opaque"
)

var errPair = errors.New("personal pairing expired or could not be verified")

func PairCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

type pairWire struct {
	conn     *tls.Conn
	lines    *bufio.Scanner
	conf     *opaque.Configuration
	password []byte
}

func ValidPairCode(code string) bool {
	if len(code) != 4 {
		return false
	}
	for _, digit := range code {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func startPair(ctx context.Context, conn *tls.Conn, sender, receiver string) (*pairWire, error) {
	if !ValidPairCode(sender) || !ValidPairCode(receiver) {
		return nil, errPair
	}
	deadline := time.Now().Add(90 * time.Second)
	d, ok := ctx.Deadline()
	if ok && d.Before(deadline) {
		deadline = d
	}
	err := conn.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}
	err = conn.HandshakeContext(ctx)
	if err != nil {
		return nil, errPair
	}
	state := conn.ConnectionState()
	if state.Version != tls.VersionTLS13 {
		return nil, errPair
	}
	binding, err := state.ExportKeyingMaterial("EXPORTER-Apparat-Personal-Pair-v1", nil, 32)
	if err != nil {
		return nil, errPair
	}
	conf := opaque.DefaultConfiguration()
	conf.Context = binding
	lines := bufio.NewScanner(conn)
	lines.Buffer(make([]byte, 4096), 512<<10)
	return &pairWire{conn, lines, conf, []byte(sender + receiver)}, nil
}

func (w *pairWire) write(data []byte) error {
	return json.NewEncoder(w.conn).Encode(data)
}
func (w *pairWire) read(limit int) ([]byte, error) {
	if !w.lines.Scan() || len(w.lines.Bytes()) > base64.StdEncoding.EncodedLen(limit)+2 {
		return nil, errPair
	}
	var data []byte
	err := json.Unmarshal(w.lines.Bytes(), &data)
	if err != nil || len(data) == 0 || len(data) > limit {
		return nil, errPair
	}
	return data, nil
}

var senderID = []byte("apparat/personal/sender/v1")
var receiverID = []byte("apparat/personal/receiver/v1")

func (w *pairWire) sender() error {
	server, err := w.conf.Server()
	if err != nil {
		return err
	}
	priv, pub := w.conf.KeyGen()
	defer priv.Zero()
	err = server.SetKeyMaterial(&opaque.ServerKeyMaterial{Identity: senderID, PrivateKey: priv, PublicKeyBytes: pub.Encode(), OPRFGlobalSeed: w.conf.GenerateOPRFSeed()})
	if err != nil {
		return err
	}
	client, err := w.conf.Client()
	if err != nil {
		return err
	}
	request, err := client.RegistrationInit(w.password)
	if err != nil {
		return err
	}
	id := opaque.RandomBytes(64)
	reg, err := server.RegistrationResponse(request, id, nil)
	if err != nil {
		return err
	}
	record, _, err := client.RegistrationFinalize(reg, receiverID, senderID)
	if err != nil {
		return err
	}
	raw, err := w.read(4096)
	if err != nil {
		return err
	}
	ke1, err := server.Deserialize.KE1(raw)
	if err != nil {
		return errPair
	}
	ke2, out, err := server.GenerateKE2(ke1, &opaque.ClientRecord{CredentialIdentifier: id, ClientIdentity: receiverID, RegistrationRecord: record})
	if err != nil {
		return errPair
	}
	defer clear(out.SessionSecret)
	err = w.write(ke2.Serialize())
	if err != nil {
		return err
	}
	raw, err = w.read(4096)
	if err != nil {
		return err
	}
	ke3, err := server.Deserialize.KE3(raw)
	if err != nil || server.LoginFinish(ke3, out.ClientMAC) != nil {
		return errPair
	}
	return nil
}

func (w *pairWire) receiver() error {
	client, err := w.conf.Client()
	if err != nil {
		return err
	}
	ke1, err := client.GenerateKE1(w.password)
	if err != nil {
		return err
	}
	err = w.write(ke1.Serialize())
	if err != nil {
		return err
	}
	raw, err := w.read(4096)
	if err != nil {
		return err
	}
	ke2, err := client.Deserialize.KE2(raw)
	if err != nil {
		return errPair
	}
	ke3, key, export, err := client.GenerateKE3(ke2, receiverID, senderID)
	defer clear(key)
	defer clear(export)
	if err != nil {
		return errPair
	}
	return w.write(ke3.Serialize())
}

// SendPersonalPair releases the transfer only after both codes match on this TLS channel.
// The supplying callback returns the password-encrypted recovery envelope and public metadata.
func SendPersonalPair(ctx context.Context, conn *tls.Conn, sender, receiver string, offer func() ([]byte, error)) error {
	stop := context.AfterFunc(ctx, func() {
		conn.NetConn().Close()
	})
	defer stop()
	defer conn.NetConn().Close()
	wire, err := startPair(ctx, conn, sender, receiver)
	if err != nil {
		return err
	}
	defer clear(wire.password)
	err = wire.sender()
	if err != nil {
		return err
	}
	packet, err := offer()
	if err != nil {
		return err
	}
	defer clear(packet)
	if len(packet) == 0 || len(packet) > 256<<10 {
		return errPair
	}
	return wire.write(packet)
}

// ReceivePersonalPair authenticates the sender before returning the encrypted transfer.
func ReceivePersonalPair(ctx context.Context, conn *tls.Conn, sender, receiver string) ([]byte, error) {
	stop := context.AfterFunc(ctx, func() {
		conn.NetConn().Close()
	})
	defer stop()
	defer conn.NetConn().Close()
	wire, err := startPair(ctx, conn, sender, receiver)
	if err != nil {
		return nil, err
	}
	defer clear(wire.password)
	err = wire.receiver()
	if err != nil {
		return nil, err
	}
	return wire.read(256 << 10)
}

// PairTLS uses node-certified leaves for both devices; the two codes authenticate their
// shared channel on first contact. A presented chain must still prove the device key.
func PairTLS(conn net.Conn, device *identity.Session, sending bool) (*tls.Conn, error) {
	local, remote := x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth
	if sending {
		local, remote = remote, local
	}
	host, _ := os.Hostname()
	leaf, err := namedCertificate(device, local, pkix.Name{CommonName: claimDNS, Organization: []string{host}, OrganizationalUnit: []string{runtime.GOOS}})
	if err != nil {
		return nil, err
	}
	cfg := &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{leaf}, VerifyConnection: func(st tls.ConnectionState) error {
		pub := pairDevice(st)
		if len(pub) == 0 {
			return errPair
		}
		return verifyPeerChain(st, identity.Fingerprint(pub), remote)
	}}
	if sending {
		cfg.ClientAuth = tls.RequireAnyClientCert
		return tls.Server(conn, cfg), nil
	}
	cfg.InsecureSkipVerify = true // The node chain is verified above and authenticated by OPAQUE.
	return tls.Client(conn, cfg), nil
}

func PairDeviceKey(conn *tls.Conn) ed25519.PublicKey {
	return pairDevice(conn.ConnectionState())
}
func pairDevice(state tls.ConnectionState) ed25519.PublicKey {
	if len(state.PeerCertificates) != 2 {
		return nil
	}
	pub, _ := state.PeerCertificates[1].PublicKey.(ed25519.PublicKey)
	return pub
}

func PairedPeer(conn *tls.Conn) domain.PoolDevice {
	st := conn.ConnectionState()
	d := domain.PoolDevice{ID: identity.Fingerprint(pairDevice(st))}
	subject := st.PeerCertificates[0].Subject
	if len(subject.Organization) == 1 && len(subject.Organization[0]) <= 255 && strings.IndexFunc(subject.Organization[0], unicode.IsControl) < 0 {
		d.Host = subject.Organization[0]
	}
	if len(subject.OrganizationalUnit) == 1 && len(subject.OrganizationalUnit[0]) <= 32 && strings.IndexFunc(subject.OrganizationalUnit[0], unicode.IsControl) < 0 {
		d.OS = subject.OrganizationalUnit[0]
	}
	return d
}
