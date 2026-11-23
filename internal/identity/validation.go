// Validates bounded identity files and replaces them atomically.
// Ready is local bootstrap validity; decryption and remote trust are separate checks.
package identity

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type encryptedPayload struct {
	Salt       string `json:"salt"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
	Algorithm  string `json:"algorithm,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
}

func strictJSON(data []byte, target any) error {
	if len(data) > 16384 {
		return errors.New("identity payload too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(target)
	if err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return errors.New("identity contains trailing JSON")
	}
	return nil
}
func readPrivate(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16384 {
		return nil, errors.New("invalid private-key file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("private-key permissions must exclude group/other access")
	}
	return os.ReadFile(path)
}
func decodePayload(data []byte) (encryptedPayload, []byte, []byte, []byte, error) {
	var payload encryptedPayload
	err := strictJSON(data, &payload)
	if err != nil {
		return payload, nil, nil, nil, err
	}
	if payload.Algorithm != "" && payload.Algorithm != "argon2id-xchacha20poly1305" {
		return payload, nil, nil, nil, errors.New("unsupported key algorithm")
	}
	salt, e1 := base64.StdEncoding.DecodeString(payload.Salt)
	nonce, e2 := base64.StdEncoding.DecodeString(payload.Nonce)
	cipher, e3 := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if e1 != nil || e2 != nil || e3 != nil || len(salt) != 16 || len(nonce) != 24 || len(cipher) != 80 {
		return payload, nil, nil, nil, errors.New("invalid encrypted-key encoding or lengths")
	}
	return payload, salt, nonce, cipher, nil
}
func readManifest(path string) (Manifest, error) {
	var manifest Manifest
	info, err := os.Lstat(path)
	if err != nil {
		return manifest, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16384 {
		return manifest, errors.New("invalid manifest file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest, err
	}
	err = strictJSON(data, &manifest)
	if err != nil {
		return manifest, err
	}
	public, err := base64.StdEncoding.DecodeString(manifest.PublicKey)
	if err != nil || len(public) != ed25519.PublicKeySize || manifest.Fingerprint != Fingerprint(public) {
		return manifest, errors.New("invalid manifest key or fingerprint")
	}
	if manifest.Kind != "device" && manifest.Kind != "user" {
		return manifest, errors.New("invalid identity kind")
	}
	_, err = time.Parse(time.RFC3339Nano, manifest.CreatedAt)
	if err != nil {
		return manifest, errors.New("invalid identity creation time")
	}
	if manifest.RevokedAt != "" {
		_, err = time.Parse(time.RFC3339Nano, manifest.RevokedAt)
	}
	return manifest, err
}
func atomicWrite(path string, data []byte, mode os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".identity-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()
	err = file.Chmod(mode)
	if err == nil {
		_, err = file.Write(data)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	err = os.Rename(file.Name(), path)
	if err != nil {
		return err
	}
	directory, err := os.Open(filepath.Dir(path))
	if err == nil {
		_ = directory.Sync()
		_ = directory.Close()
	}
	return nil
}

// DeviceFingerprint reads the active public manifest without unlocking the signer.
func DeviceFingerprint(dir string) (string, error) {
	_, selected, err := current(dir)
	if err != nil {
		return "", err
	}
	m, err := readManifest(filepath.Join(selected, "device.manifest.json"))
	if err != nil || m.Kind != "device" || m.RevokedAt != "" {
		return "", errors.New("no active device manifest")
	}
	return m.Fingerprint, nil
}
