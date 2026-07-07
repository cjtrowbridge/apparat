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
func Sign(priv ed25519.PrivateKey, message []byte) []byte { return ed25519.Sign(priv, message) }
func Verify(pub ed25519.PublicKey, message, sig []byte) bool {
	return ed25519.Verify(pub, message, sig)
}
func Fingerprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func SaveEncrypted(path string, priv ed25519.PrivateKey, passphrase []byte) error {
	if len(passphrase) == 0 {
		return errors.New("passphrase required")
	}
	var salt [16]byte
	var nonce [chacha20poly1305.NonceSizeX]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return err
	}
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	key := argon2.IDKey(passphrase, salt[:], 1, 64*1024, 4, chacha20poly1305.KeySize)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return err
	}
	payload := map[string]string{"salt": base64.StdEncoding.EncodeToString(salt[:]), "nonce": base64.StdEncoding.EncodeToString(nonce[:]), "ciphertext": base64.StdEncoding.EncodeToString(aead.Seal(nil, nonce[:], priv, nil))}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func LoadEncrypted(path string, passphrase []byte) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload map[string]string
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	salt, _ := base64.StdEncoding.DecodeString(payload["salt"])
	nonce, _ := base64.StdEncoding.DecodeString(payload["nonce"])
	ciphertext, _ := base64.StdEncoding.DecodeString(payload["ciphertext"])
	key := argon2.IDKey(passphrase, salt, 1, 64*1024, 4, chacha20poly1305.KeySize)
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	plain, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return ed25519.PrivateKey(plain), nil
}

func WriteManifest(path, kind string, pub ed25519.PublicKey) error {
	manifest := Manifest{Kind: kind, PublicKey: base64.StdEncoding.EncodeToString(pub), CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Fingerprint: Fingerprint(pub)}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
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
	if err := SaveEncrypted(filepath.Join(dir, "device.key.json"), pair.Private, passphrase); err != nil {
		return Manifest{}, err
	}
	if err := WriteManifest(filepath.Join(dir, "device.manifest.json"), "device", pair.Public); err != nil {
		return Manifest{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "device.manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	return manifest, json.Unmarshal(data, &manifest)
}

func RevokeManifest(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	manifest.RevokedAt = time.Now().UTC().Format(time.RFC3339Nano)
	updated, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, updated, 0o644)
}

func Classify(dir string) Status {
	_, privErr := os.Stat(filepath.Join(dir, "device.key.json"))
	_, manifestErr := os.Stat(filepath.Join(dir, "device.manifest.json"))
	if privErr == nil && manifestErr == nil {
		return StatusReady
	}
	if os.IsNotExist(privErr) && os.IsNotExist(manifestErr) {
		return StatusMissing
	}
	return StatusInconsistent
}

func ArchivedReset(dir string) (string, error) {
	archive := filepath.Join(filepath.Dir(dir), "identity-archive-"+time.Now().UTC().Format("20060102150405"))
	if err := os.Rename(dir, archive); err != nil {
		return "", err
	}
	return archive, nil
}
