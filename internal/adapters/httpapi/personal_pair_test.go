// Exercises mutual code confirmation over TLS without exposing any real identity.
// Uses disposable channels for matching and mismatched code exchanges.
package httpapi

import (
	"context"
	"crypto/tls"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPersonalPairCodes(t *testing.T) {
	testServer := httptest.NewTLSServer(nil)
	defer testServer.Close()
	for _, same := range []bool{true, false} {
		a, b := net.Pipe()
		send := tls.Server(a, &tls.Config{MinVersion: tls.VersionTLS13, Certificates: testServer.TLS.Certificates})
		config := testServer.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
		config.ServerName = "example.com"
		recv := tls.Client(b, config)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		called := false
		done := make(chan error, 1)
		go func() {
			done <- SendPersonalPair(ctx, send, "0123", "0022", func() ([]byte, error) {
				called = true
				return []byte("fixture"), nil
			})
		}()
		code := "0022"
		if !same {
			code = "0023"
		}
		value, err := ReceivePersonalPair(ctx, recv, "0123", code)
		senderErr := <-done
		cancel()
		if same && (err != nil || senderErr != nil || string(value) != "fixture") {
			t.Fatal("matching pair failed", err, senderErr)
		}
		if !same && (err == nil || senderErr == nil || called || len(value) != 0) {
			t.Fatal("wrong code released identity")
		}
	}
}

func TestTransferPortableIdentityAndWorkerPin(t *testing.T) {
	source := identity.NewPersonal(t.TempDir())
	err := source.Create("owner", []byte("fixture password"))
	if err != nil {
		t.Fatal(err)
	}
	receiver := identity.NewPersonal(t.TempDir())
	deviceA := identity.NewSession(t.TempDir())
	deviceB := identity.NewSession(t.TempDir())
	for _, keys := range []*identity.Session{deviceA, deviceB} {
		err = keys.Install(nil, "", "plain", "", "", true)
		if err != nil {
			t.Fatal(err)
		}
	}
	a, b := net.Pipe()
	send, err := PairTLS(a, deviceA, true)
	if err != nil {
		t.Fatal(err)
	}
	recv, err := PairTLS(b, deviceB, false)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	want := domain.ClusterDevice{Host: "worker", Address: "192.0.2.2:8080", Cluster: "fixture", Fingerprint: "fixture-pin"}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go func() {
		done <- SendPersonalPair(ctx, send, "0004", "1000", func() ([]byte, error) {
			return source.PairOffer(PairDeviceKey(send), []domain.ClusterDevice{want})
		})
	}()
	packet, err := ReceivePersonalPair(ctx, recv, "0004", "1000")
	sendErr := <-done
	if err != nil || sendErr != nil {
		t.Fatal(err, sendErr)
	}
	_, err = receiver.AcceptPair(packet, []byte("fixture password"), deviceA.Public())
	if err == nil {
		t.Fatal("wrong recipient accepted")
	}
	_, err = receiver.AcceptPair(packet, []byte("bad password"), deviceB.Public())
	if err == nil {
		t.Fatal("wrong recovery password accepted")
	}
	peers, err := receiver.AcceptPair(packet, []byte("fixture password"), deviceB.Public())
	if err != nil || source.Fingerprint() != receiver.Fingerprint() || len(peers) != 1 || peers[0] != want {
		t.Fatal("failed to restore user and worker", err)
	}
}

func TestPairCodeFormat(t *testing.T) {
	code, err := PairCode()
	if err != nil || !ValidPairCode(code) || !ValidPairCode("0000") {
		t.Fatal("invalid generated code format")
	}
	for _, bad := range []string{"123", "12345", "12345678", "12a4", "１２３４"} {
		if ValidPairCode(bad) {
			t.Fatal("accepted invalid pairing code")
		}
	}
}
