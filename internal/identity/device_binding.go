// Issues and verifies public evidence that a user authorized a specific device key.
// A binding does not grant services; cluster policy decides whether it is still accepted.
package identity

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
)

const bindingDomain = "apparat/personal-device/v1"

type Binding struct {
	Domain string
	User   ed25519.PublicKey
	Device ed25519.PublicKey
}

type SignedBinding struct {
	Binding   Binding
	Signature []byte
}

func (b SignedBinding) Valid() bool {
	v := b.Binding
	if v.Domain != bindingDomain || len(v.User) != ed25519.PublicKeySize || len(v.Device) != ed25519.PublicKeySize || bytes.Equal(v.Device, v.User) {
		return false
	}
	data, err := json.Marshal(v)
	return err == nil && ed25519.Verify(v.User, data, b.Signature)
}

func (s *Session) Public() ed25519.PublicKey {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.key) == 0 {
		return nil
	}
	return s.key.Public().(ed25519.PublicKey)
}

func (p *Personal) Credentials(device ed25519.PublicKey) (SignedProfile, SignedBinding, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var bundle personalBundle
	if len(p.key) == 0 || len(device) != ed25519.PublicKeySize {
		return SignedProfile{}, SignedBinding{}, errors.New("unlock both identities")
	}
	data, err := readPrivate(p.path)
	if err == nil {
		err = strictJSON(data, &bundle)
	}
	user := p.key.Public().(ed25519.PublicKey)
	if err != nil || !bundle.Profile.Valid() || !bytes.Equal(user, bundle.Profile.Profile.Public) || bytes.Equal(user, device) {
		return SignedProfile{}, SignedBinding{}, errors.New("inconsistent user or device identity")
	}
	value := Binding{bindingDomain, user, device}
	message, _ := json.Marshal(value)
	return bundle.Profile, SignedBinding{value, ed25519.Sign(p.key, message)}, nil
}
