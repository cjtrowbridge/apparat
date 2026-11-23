// Signs a recipient-specific directory of worker pins with the portable user identity.
// Transfers contain only the validated password-encrypted bundle and public worker metadata.
package identity

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

const transferDomain = "apparat/personal-pair/v1"

type pairTransfer struct {
	Domain   string
	Receiver ed25519.PublicKey
	Bundle   personalBundle
	Workers  []domain.ClusterDevice
}

type signedTransfer struct {
	Data      pairTransfer
	Signature []byte
}

func (p *Personal) PairOffer(receiver ed25519.PublicKey, workers []domain.ClusterDevice) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.key) != ed25519.PrivateKeySize || len(receiver) != ed25519.PublicKeySize {
		return nil, errTransfer
	}
	raw, err := readPrivate(p.path)
	if err != nil {
		return nil, err
	}
	var bundle personalBundle
	err = strictJSON(raw, &bundle)
	if err != nil || !bundle.Profile.Valid() || !bytes.Equal(bundle.Profile.Profile.Public, p.key.Public().(ed25519.PublicKey)) {
		return nil, errTransfer
	}
	_, _, _, _, err = decodePayload(bundle.Encrypted)
	if err != nil {
		return nil, errTransfer
	}
	packet := signedTransfer{Data: pairTransfer{transferDomain, receiver, bundle, workers}}
	body, err := json.Marshal(packet.Data)
	if err != nil {
		return nil, err
	}
	packet.Signature = ed25519.Sign(p.key, body)
	return json.Marshal(packet)
}

var errTransfer = errors.New("personal identity transfer could not be verified")

func (p *Personal) AcceptPair(raw, password []byte, receiver ed25519.PublicKey) ([]domain.ClusterDevice, error) {
	var packet signedTransfer
	if len(raw) > 256<<10 {
		return nil, errTransfer
	}
	err := strictJSON(raw, &packet)
	data := packet.Data
	if err != nil || data.Domain != transferDomain || len(receiver) != ed25519.PublicKeySize || !bytes.Equal(receiver, data.Receiver) || !data.Bundle.Profile.Valid() {
		return nil, errTransfer
	}
	body, err := json.Marshal(data)
	if err != nil || !ed25519.Verify(data.Bundle.Profile.Profile.Public, body, packet.Signature) {
		return nil, errTransfer
	}
	bundle, err := json.Marshal(data.Bundle)
	if err != nil {
		return nil, err
	}
	existing := p.Fingerprint()
	if existing != "" {
		if existing != Fingerprint(data.Bundle.Profile.Profile.Public) {
			return nil, errTransfer
		}
		_, key, err := unlockBundle(bundle, password)
		clear(key)
		if err != nil {
			return nil, err
		}
		return data.Workers, nil
	}
	err = p.Import(bundle, password)
	if err != nil {
		return nil, err
	}
	return data.Workers, nil
}
