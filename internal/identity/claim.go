// Binds a first-admin claim to its one-time invitation, cluster, worker and two signing identities.
// Proof verification is independent of consuming the invitation in a durable transaction.
package identity

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type Claim struct {
	Profile         SignedProfile
	Device          SignedBinding
	Cluster         string
	Worker          string
	Invitation      string
	UserSignature   []byte
	DeviceSignature []byte
}

func ClaimDigest(invitation string) string {
	hash := sha256.Sum256([]byte(invitation))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func (c Claim) challenge() []byte {
	statement := struct {
		Domain           string
		Cluster          string
		Worker           string
		InvitationDigest string
		Profile          SignedProfile
		Device           SignedBinding
	}{"apparat/worker-first-claim/v1", c.Cluster, c.Worker, ClaimDigest(c.Invitation), c.Profile, c.Device}
	data, _ := json.Marshal(statement)
	return data
}

func (c Claim) Valid() bool {
	if !c.Profile.Valid() || !c.Device.Valid() || !bytes.Equal(c.Profile.Profile.Public, c.Device.Binding.User) {
		return false
	}
	for _, value := range []string{c.Cluster, c.Worker, c.Invitation} {
		raw, err := base64.RawURLEncoding.Strict().DecodeString(value)
		if err != nil || len(raw) != 32 {
			return false
		}
	}
	message := c.challenge()
	return ed25519.Verify(c.Device.Binding.User, message, c.UserSignature) && ed25519.Verify(c.Device.Binding.Device, message, c.DeviceSignature)
}

func (p *Personal) SignClaim(device *Session, cluster, worker, invitation string) (Claim, error) {
	profile, binding, err := p.Credentials(device.Public())
	if err != nil {
		return Claim{}, err
	}
	c := Claim{Profile: profile, Device: binding, Cluster: cluster, Worker: worker, Invitation: invitation}
	message := c.challenge()
	p.mu.Lock()
	if len(p.key) != ed25519.PrivateKeySize {
		p.mu.Unlock()
		return Claim{}, errors.New("unlock personal identity")
	}
	c.UserSignature = ed25519.Sign(p.key, message)
	p.mu.Unlock()
	c.DeviceSignature, err = device.Sign(message)
	if err != nil || !c.Valid() {
		return Claim{}, errors.New("invalid claim invitation or signing identities")
	}
	return c, nil
}
