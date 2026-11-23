// Stores one portable user identity separately from the installation's node key.
// A signed profile binds the public key and screen name to the user protocol.
package identity

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const profileDomain = "apparat/user-profile/v1"

type UserProfile struct {
	Domain   string
	Public   ed25519.PublicKey
	Name     string
	Revision uint64
}

type SignedProfile struct {
	Profile   UserProfile
	Signature []byte
}

type personalBundle struct {
	Profile   SignedProfile
	Encrypted json.RawMessage
}

type Personal struct {
	mu   sync.Mutex
	path string
	key  ed25519.PrivateKey
}

func NewPersonal(dir string) *Personal {
	return &Personal{path: filepath.Join(dir, "user.json")}
}

func validName(name string) bool {
	return utf8.ValidString(name) && strings.TrimSpace(name) == name && name != "" && utf8.RuneCountInString(name) <= 64 && strings.IndexFunc(name, unicode.IsControl) < 0
}

func (p SignedProfile) Valid() bool {
	profile := p.Profile
	if profile.Domain != profileDomain || !validName(profile.Name) || profile.Revision == 0 || len(profile.Public) != ed25519.PublicKeySize {
		return false
	}
	data, err := json.Marshal(profile)
	return err == nil && ed25519.Verify(profile.Public, data, p.Signature)
}

func signProfile(key ed25519.PrivateKey, name string, revision uint64) SignedProfile {
	profile := UserProfile{profileDomain, key.Public().(ed25519.PublicKey), name, revision}
	data, _ := json.Marshal(profile)
	return SignedProfile{profile, ed25519.Sign(key, data)}
}

func unlockBundle(data, password []byte) (personalBundle, ed25519.PrivateKey, error) {
	var bundle personalBundle
	err := strictJSON(data, &bundle)
	if err != nil || !bundle.Profile.Valid() || len(password) == 0 {
		return bundle, nil, errors.New("invalid personal identity or password")
	}
	key, err := decryptKey(bundle.Encrypted, password)
	if err != nil || !bytes.Equal(key.Public().(ed25519.PublicKey), bundle.Profile.Profile.Public) {
		clear(key)
		return bundle, nil, errors.New("could not unlock personal identity")
	}
	return bundle, key, nil
}

func (p *Personal) Create(name string, password []byte) error {
	if !validName(name) || len(password) == 0 {
		return errors.New("enter a screen name of up to 64 characters and a password")
	}
	pair, err := Generate()
	if err != nil {
		return err
	}
	defer clear(pair.Private)
	encrypted, err := encryptKey(pair.Private, password)
	if err != nil {
		return err
	}
	data, err := json.Marshal(personalBundle{signProfile(pair.Private, name, 1), encrypted})
	if err != nil {
		return err
	}
	return p.Import(data, password)
}

func (p *Personal) Import(data, password []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, key, err := unlockBundle(data, password)
	if err != nil {
		return err
	}
	defer clear(key)
	_, err = os.Lstat(p.path)
	if !os.IsNotExist(err) {
		return errors.New("personal identity already exists or is inaccessible")
	}
	dir := filepath.Dir(p.path)
	err = os.MkdirAll(dir, 0700)
	if err == nil {
		err = protectDirectory(dir)
	}
	if err == nil {
		err = atomicWrite(p.path, data, 0600)
	}
	if err == nil {
		p.key = bytes.Clone(key)
	}
	return err
}

func (p *Personal) Unlock(password []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := readPrivate(p.path)
	if err != nil {
		return err
	}
	_, key, err := unlockBundle(data, password)
	if err == nil {
		clear(p.key)
		p.key = key
	}
	return err
}

func (p *Personal) Lock() {
	p.mu.Lock()
	defer p.mu.Unlock()
	clear(p.key)
	p.key = nil
}

func (p *Personal) Rename(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.key) == 0 || !validName(name) {
		return errors.New("unlock your identity and enter a valid screen name")
	}
	data, err := readPrivate(p.path)
	if err != nil {
		return err
	}
	var b personalBundle
	err = strictJSON(data, &b)
	if err != nil || !b.Profile.Valid() || !bytes.Equal(b.Profile.Profile.Public, p.key.Public().(ed25519.PublicKey)) {
		return errors.New("invalid local profile")
	}
	b.Profile = signProfile(p.key, name, b.Profile.Profile.Revision+1)
	data, err = json.Marshal(b)
	if err != nil {
		return err
	}
	return atomicWrite(p.path, data, 0600)
}

func (p *Personal) Fingerprint() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.key) == 0 {
		return ""
	}
	return Fingerprint(p.key.Public().(ed25519.PublicKey))
}
