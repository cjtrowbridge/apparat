// Signs cluster policy proposals with the current admin's personal and device identities.
// Authority countersignatures cover the ordered proposal and its previous chain hash.
package identity

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
)

type PolicyEdit struct {
	Epoch   *uint64 `json:",omitempty"`
	Cluster string
	ID      string
	Base    uint64
	Kind    string
	Target  string
	Service string
	Pool    string
	Value   string
}

type Proposal struct {
	Edit            PolicyEdit
	Device          SignedBinding
	UserSignature   []byte
	DeviceSignature []byte
}

func (p Proposal) message() []byte {
	data, _ := json.Marshal(struct {
		Domain string
		Edit   PolicyEdit
		Device SignedBinding
	}{"apparat/policy-proposal/v1", p.Edit, p.Device})
	return data
}

func (p Proposal) Valid() bool {
	if !p.Device.Valid() || p.Edit.ID == "" || len(p.Edit.ID) > 128 || len(p.Edit.Cluster) != 43 {
		return false
	}
	message := p.message()
	return ed25519.Verify(p.Device.Binding.User, message, p.UserSignature) && ed25519.Verify(p.Device.Binding.Device, message, p.DeviceSignature)
}

func (p *Personal) Propose(device *Session, edit PolicyEdit) (Proposal, error) {
	_, binding, err := p.Credentials(device.Public())
	if err != nil {
		return Proposal{}, err
	}
	proposal := Proposal{Edit: edit, Device: binding}
	message := proposal.message()
	p.mu.Lock()
	if len(p.key) == 0 {
		p.mu.Unlock()
		return Proposal{}, errors.New("personal identity locked")
	}
	proposal.UserSignature = ed25519.Sign(p.key, message)
	p.mu.Unlock()
	proposal.DeviceSignature, err = device.Sign(message)
	if err != nil || !proposal.Valid() {
		return Proposal{}, errors.New("invalid policy proposal")
	}
	return proposal, nil
}
