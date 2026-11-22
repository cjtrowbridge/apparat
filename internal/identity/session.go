// Owns one node's in-memory signing key and the atomic selection of its on-disk generation.
// Legacy key pairs remain readable; incomplete or replaced generations remain available for recovery.
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
)

const RecoveryWarning = "If you have used Apparat on this device before, find your original private key to keep your cluster connections. Generating a new key gives this device a different identity. You will need to pair it with your clusters again."

type generation struct {
	Directory string
	Mode      string
	Reference string
}

type Session struct {
	mu  sync.Mutex
	dir string
	key ed25519.PrivateKey
}

func NewSession(dir string) *Session {
	return &Session{dir: dir}
}

func current(dir string) (generation, string, error) {
	path := filepath.Join(dir, "active.json")
	data, err := readPrivate(path)
	if os.IsNotExist(err) {
		return generation{Mode: "password"}, dir, nil
	}
	var gen generation
	if err != nil || strictJSON(data, &gen) != nil || filepath.Base(gen.Directory) != gen.Directory || !strings.HasPrefix(gen.Directory, "pair-") {
		return gen, "", errors.New("invalid active identity")
	}
	switch gen.Mode {
	case "password", "plain":
	case "system":
		if gen.Reference == "" {
			return gen, "", errors.New("missing system credential reference")
		}
	default:
		return gen, "", errors.New("unknown key protection")
	}
	return gen, filepath.Join(dir, gen.Directory), nil
}

func loadPair(dir, mode string, password []byte) (ed25519.PrivateKey, error) {
	path := filepath.Join(dir, "device.key.json")
	var key ed25519.PrivateKey
	var err error
	if mode == "plain" {
		key, err = readPrivate(path)
	} else {
		key, err = LoadEncrypted(path, password)
	}
	if err != nil || len(key) != ed25519.PrivateKeySize || !bytes.Equal(ed25519.NewKeyFromSeed(key[:32]), key) {
		return nil, errors.New("could not unlock private key")
	}
	m, err := readManifest(filepath.Join(dir, "device.manifest.json"))
	if err != nil || m.Kind != "device" || m.RevokedAt != "" || m.Fingerprint != Fingerprint(key.Public().(ed25519.PublicKey)) {
		return nil, errors.New("key does not match an active device manifest")
	}
	return key, nil
}

func (s *Session) State() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.key) > 0 {
		return "unlocked"
	}
	gen, dir, err := current(s.dir)
	if err != nil {
		return "recovery"
	}
	status := Classify(s.dir)
	if status == StatusMissing {
		return "missing"
	}
	if status != StatusReady || dir == "" {
		return "recovery"
	}
	return gen.Mode
}

func (s *Session) Reference() string {
	gen, _, _ := current(s.dir)
	return gen.Reference
}

func (s *Session) Unlock(secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	gen, dir, err := current(s.dir)
	if err != nil {
		return err
	}
	key, err := loadPair(dir, gen.Mode, []byte(secret))
	if err == nil {
		clear(s.key)
		s.key = key
	}
	return err
}

func (s *Session) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.key)
	s.key = nil
}

func (s *Session) Install(backup []byte, oldPassword, mode, secret, reference string, generate bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if mode != "plain" && mode != "password" && mode != "system" || mode != "plain" && secret == "" || mode == "system" && reference == "" {
		return errors.New("choose valid key protection")
	}
	if len(backup) > 16384 || generate != (len(backup) == 0) {
		return errors.New("choose a private key or explicitly generate a new pair")
	}
	err := os.MkdirAll(s.dir, 0700)
	if err != nil {
		return err
	}
	err = protectDirectory(s.dir)
	if err != nil {
		return errors.New("could not restrict identity directory permissions")
	}
	dir, err := os.MkdirTemp(s.dir, "pair-")
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(dir)
		}
	}()
	var key ed25519.PrivateKey
	if generate {
		pair, err := Generate()
		if err != nil {
			return err
		}
		key = pair.Private
	} else if len(backup) == 64 && bytes.Equal(ed25519.NewKeyFromSeed(backup[:32]), backup) {
		key = bytes.Clone(backup)
	} else {
		key, err = decryptKey(backup, []byte(oldPassword))
		if err != nil {
			return errors.New("could not unlock the selected Apparat private key")
		}
	}
	defer clear(key)
	pub := key.Public().(ed25519.PublicKey)
	if !generate {
		_, previous, readErr := current(s.dir)
		if readErr == nil {
			m, manifestErr := readManifest(filepath.Join(previous, "device.manifest.json"))
			if manifestErr == nil && (m.Fingerprint != Fingerprint(pub) || m.RevokedAt != "" || m.Kind != "device") {
				return errors.New("selected key does not match this node's previous active identity")
			}
		}
	}
	path := filepath.Join(dir, "device.key.json")
	if mode == "plain" {
		err = atomicWrite(path, key, 0600)
	} else {
		err = SaveEncrypted(path, key, []byte(secret))
	}
	if err != nil {
		return err
	}
	err = WriteManifest(filepath.Join(dir, "device.manifest.json"), "device", pub)
	if err != nil {
		return err
	}
	verified, err := loadPair(dir, mode, []byte(secret))
	if err != nil {
		return err
	}
	data, err := json.Marshal(generation{filepath.Base(dir), mode, reference})
	if err == nil {
		err = atomicWrite(filepath.Join(s.dir, "active.json"), data, 0600)
	}
	if err != nil {
		clear(verified)
		return err
	}
	committed = true
	clear(s.key)
	s.key = verified
	return nil
}

// Sign requires an owner-unlocked identity. Callers authorize their operation separately.
func (s *Session) Sign(message []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.key) != ed25519.PrivateKeySize {
		return nil, errors.New("node identity is locked")
	}
	return Sign(s.key, message), nil
}
