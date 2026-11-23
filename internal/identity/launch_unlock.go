// Protects an opted-in personal unlock credential with the device's OS-gated secret.
// A portable device key backup cannot derive this local wrapping key.
package identity

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"path/filepath"
	"strings"
)

const launchPrefix = "apparat.launch.v1:"

func (s *Session) setLaunchKey(mode, secret string) {
	clear(s.launchKey)
	s.launchKey = nil
	if mode == "system" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte("apparat/local-personal-unlock/v1"))
		s.launchKey = mac.Sum(nil)
	}
}
func (s *Session) launchCipher() cipher.AEAD {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	block, err := aes.NewCipher(s.launchKey)
	if err != nil {
		return nil
	}
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil
	}
	return aead
}
func (s *Session) wrapLocal(password string) (string, bool) {
	aead := s.launchCipher()
	if aead == nil {
		return "", false
	}
	sealed := aead.Seal(nil, nil, []byte(password), nil)
	return launchPrefix + base64.RawStdEncoding.EncodeToString(sealed), true
}
func (p *Personal) unlockLocal(ui setupUI, device *Session, ref string) bool {
	encoded, local := strings.CutPrefix(ref, launchPrefix)
	if local {
		aead := device.launchCipher()
		sealed, err := base64.RawStdEncoding.DecodeString(encoded)
		if aead == nil || err != nil || len(sealed) < aead.Overhead() {
			return false
		}
		secret, err := aead.Open(nil, nil, sealed, nil)
		defer clear(secret)
		return err == nil && p.Unlock(secret) == nil
	}
	password, ok := ui.ask("load", "Unlock personal identity", "", ref)
	if !ok || p.Unlock([]byte(password)) != nil {
		return false
	}
	replacement, wrapped := device.wrapLocal(password)
	if wrapped {
		err := atomicWrite(filepath.Join(filepath.Dir(p.path), "unlock.ref"), []byte(replacement), 0600)
		if err == nil {
			ui.ask("remove", "", "", ref)
		}
	}
	return true
}
