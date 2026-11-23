// Creates and persists signing identities with encrypted private-key storage.
// Dependencies: golang.org/x/crypto/argon2, golang.org/x/crypto/chacha20poly1305.
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

type KeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}
type Manifest struct {
	Kind        string `json:"kind"`
	PublicKey   string `json:"public_key"`
	CreatedAt   string `json:"created_at"`
	Fingerprint string `json:"fingerprint"`
	RevokedAt   string `json:"revoked_at,omitempty"`
}
type Status string

const (
	StatusMissing      Status = "missing"
	StatusReady        Status = "ready"
	StatusInconsistent Status = "inconsistent"
)

func Generate() (KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return KeyPair{Public: pub, Private: priv}, err
}
func Sign(priv ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(priv, message)
}
func Verify(pub ed25519.PublicKey, message, sig []byte) bool {
	return ed25519.Verify(pub, message, sig)
}
func Fingerprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func SaveEncrypted(path string, priv ed25519.PrivateKey, passphrase []byte) error {
	data, err := encryptKey(priv, passphrase)
	if err != nil {
		return err
	}
	return atomicWrite(path, data, 0600)
}

func encryptKey(priv ed25519.PrivateKey, passphrase []byte) ([]byte, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key length")
	}
	if !priv.Equal(ed25519.NewKeyFromSeed(priv.Seed())) {
		return nil, errors.New("private key public component mismatch")
	}
	if len(passphrase) == 0 {
		return nil, errors.New("passphrase required")
	}
	var salt [16]byte
	var nonce [chacha20poly1305.NonceSizeX]byte
	{
		_, err := rand.Read(salt[:])
		if err != nil {
			return nil, err
		}
	}
	{
		_, err := rand.Read(nonce[:])
		if err != nil {
			return nil, err
		}
	}
	key := argon2.IDKey(passphrase, salt[:], 1, 64*1024, 4, chacha20poly1305.KeySize)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	payload := map[string]string{"algorithm": "argon2id-xchacha20poly1305", "public_key": base64.StdEncoding.EncodeToString(priv.Public().(ed25519.PublicKey)), "salt": base64.StdEncoding.EncodeToString(salt[:]), "nonce": base64.StdEncoding.EncodeToString(nonce[:]), "ciphertext": base64.StdEncoding.EncodeToString(aead.Seal(nil, nonce[:], priv, nil))}
	return json.MarshalIndent(payload, "", "  ")
}

func LoadEncrypted(path string, passphrase []byte) (ed25519.PrivateKey, error) {
	data, err := readPrivate(path)
	if err != nil {
		return nil, err
	}
	return decryptKey(data, passphrase)
}

func decryptKey(data, passphrase []byte) (ed25519.PrivateKey, error) {
	payload, salt, nonce, ciphertext, err := decodePayload(data)
	if err != nil {
		return nil, err
	}
	key := argon2.IDKey(passphrase, salt, 1, 64*1024, 4, chacha20poly1305.KeySize)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	if len(plain) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid decrypted key length")
	}
	private := ed25519.PrivateKey(plain)
	if !private.Equal(ed25519.NewKeyFromSeed(private.Seed())) {
		return nil, errors.New("private key public component mismatch")
	}
	if payload.PublicKey != "" && payload.PublicKey != base64.StdEncoding.EncodeToString(private.Public().(ed25519.PublicKey)) {
		return nil, errors.New("encrypted public-key mismatch")
	}
	return private, nil
}

func WriteManifest(path, kind string, pub ed25519.PublicKey) error {
	if len(pub) != ed25519.PublicKeySize || (kind != "device" && kind != "user") {
		return errors.New("invalid manifest identity")
	}
	manifest := Manifest{Kind: kind, PublicKey: base64.StdEncoding.EncodeToString(pub), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Fingerprint: Fingerprint(pub)}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	{
		err := os.MkdirAll(filepath.Dir(path), 0o700)
		if err != nil {
			return err
		}
	}
	return atomicWrite(path, data, 0o644)
}

func RepairManifest(dir string, passphrase []byte) error {
	priv, err := LoadEncrypted(filepath.Join(dir, "device.key.json"), passphrase)
	if err != nil {
		return err
	}
	return WriteManifest(filepath.Join(dir, "device.manifest.json"), "device", priv.Public().(ed25519.PublicKey))
}

func RotateDevice(dir string, passphrase []byte) (Manifest, error) {
	pair, err := Generate()
	if err != nil {
		return Manifest{}, err
	}
	{
		err := SaveEncrypted(filepath.Join(dir, "device.key.json"), pair.Private, passphrase)
		if err != nil {
			return Manifest{}, err
		}
	}
	{
		err := WriteManifest(filepath.Join(dir, "device.manifest.json"), "device", pair.Public)
		if err != nil {
			return Manifest{}, err
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "device.manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	return manifest, json.Unmarshal(data, &manifest)
}

func RevokeManifest(path string) error {
	manifest, err := readManifest(path)
	if err != nil {
		return err
	}
	manifest.RevokedAt = time.Now().UTC().Format(time.RFC3339Nano)
	updated, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, updated, 0o644)
}

func Classify(dir string) Status {
	gen, selected, err := current(dir)
	if err != nil {
		return StatusInconsistent
	}
	dir = selected
	if gen.Mode == "plain" {
		key, err := loadPair(dir, "plain", nil)
		clear(key)
		if err != nil {
			return StatusInconsistent
		}
		return StatusReady
	}
	_, privErr := os.Lstat(filepath.Join(dir, "device.key.json"))
	_, manifestErr := os.Lstat(filepath.Join(dir, "device.manifest.json"))
	if privErr == nil && manifestErr == nil {
		data, err := readPrivate(filepath.Join(dir, "device.key.json"))
		if err != nil {
			return StatusInconsistent
		}
		payload, _, _, _, err := decodePayload(data)
		if err != nil {
			return StatusInconsistent
		}
		manifest, err := readManifest(filepath.Join(dir, "device.manifest.json"))
		if err != nil || manifest.RevokedAt != "" || manifest.Kind != "device" {
			return StatusInconsistent
		}
		if payload.PublicKey != "" && payload.PublicKey != manifest.PublicKey {
			return StatusInconsistent
		}
		return StatusReady
	}
	if gen.Directory == "" && os.IsNotExist(privErr) && os.IsNotExist(manifestErr) {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return StatusInconsistent
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "pair-") {
				return StatusInconsistent
			}
		}
		return StatusMissing
	}
	return StatusInconsistent
}

func ArchivedReset(dir string) (string, error) {
	archive := filepath.Join(filepath.Dir(dir), "identity-archive-"+time.Now().UTC().Format("20060102150405"))
	{
		err := os.Rename(dir, archive)
		if err != nil {
			return "", err
		}
	}
	return archive, nil
}
