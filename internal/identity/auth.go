// Binds a short-lived worker challenge to the authenticating user, device and worker.
// Clients sign only the fixed login domain. Both private keys stay local.
package identity

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"unicode"
)

type Login struct {
	Profile         SignedProfile
	Device          SignedBinding
	Worker          string
	Hostname        string
	Challenge       string
	UserSignature   []byte
	DeviceSignature []byte
}

func (l Login) message() []byte {
	l.UserSignature = nil
	l.DeviceSignature = nil
	body, _ := json.Marshal(l)
	return append([]byte("apparat/login/v1\x00"), body...)
}

func (l Login) Valid() bool {
	if l.Hostname == "" || len(l.Hostname) > 255 || strings.IndexFunc(l.Hostname, unicode.IsControl) >= 0 || len(l.Worker) != 43 || len(l.Challenge) != 43 || !l.Profile.Valid() || !l.Device.Valid() || !l.Profile.Profile.Public.Equal(l.Device.Binding.User) {
		return false
	}
	data := l.message()
	return ed25519.Verify(l.Device.Binding.User, data, l.UserSignature) && ed25519.Verify(l.Device.Binding.Device, data, l.DeviceSignature)
}

func (p *Personal) SignLogin(device *Session, worker, challenge string) (Login, error) {
	profile, binding, err := p.Credentials(device.Public())
	if err != nil {
		return Login{}, err
	}
	host, err := os.Hostname()
	if err != nil {
		return Login{}, err
	}
	login := Login{Hostname: host, Profile: profile, Device: binding, Worker: worker, Challenge: challenge}
	message := login.message()
	p.mu.Lock()
	if len(p.key) == 0 {
		p.mu.Unlock()
		return Login{}, errors.New("unlock personal identity")
	}
	login.UserSignature = ed25519.Sign(p.key, message)
	p.mu.Unlock()
	login.DeviceSignature, err = device.Sign(message)
	if err != nil || !login.Valid() {
		return Login{}, errors.New("invalid login")
	}
	return login, nil
}
