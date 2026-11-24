// Signs cluster-scoped global model revisions so a member may relay an admin's exact edit.
// Receiving workers independently check the signer's current membership and revocation state.
package identity

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type modelProof struct {
	Binding   SignedBinding
	Signature []byte
}

func modelPayload(cluster string, d domain.ModelDefault) []byte {
	d.Proof = ""
	b, _ := json.Marshal(struct {
		Domain, Cluster string
		Default         domain.ModelDefault
	}{"apparat/global-chat-model/v1", cluster, d})
	return b
}
func (p *Personal) SignModelDefault(s *Session, cluster string, d domain.ModelDefault) (string, error) {
	_, binding, err := p.Credentials(s.Public())
	if err != nil {
		return "", err
	}
	sig, err := s.Sign(modelPayload(cluster, d))
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(modelProof{binding, sig})
	return string(raw), err
}
func ModelDefaultAuthor(cluster string, d domain.ModelDefault) (user, device string, err error) {
	var proof modelProof
	if !d.Model.Valid() || d.Revision <= 0 || d.Revision >= 1<<53 || !domain.ValidID(d.Operation) || len(d.Proof) > 2048 || json.Unmarshal([]byte(d.Proof), &proof) != nil || !proof.Binding.Valid() {
		return "", "", errors.New("invalid model default proof")
	}
	binding := proof.Binding.Binding
	if !ed25519.Verify(binding.Device, modelPayload(cluster, d), proof.Signature) {
		return "", "", errors.New("invalid model default signature")
	}
	return Fingerprint(binding.User), Fingerprint(binding.Device), nil
}
