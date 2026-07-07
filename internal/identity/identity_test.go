package identity

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSignEncryptRoundTrip(t *testing.T) {
	pair, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("device authorization")
	if !Verify(pair.Public, msg, Sign(pair.Private, msg)) {
		t.Fatal("signature did not verify")
	}
	path := filepath.Join(t.TempDir(), "device.key.json")
	if err := SaveEncrypted(path, pair.Private, []byte("pass")); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadEncrypted(path, []byte("pass"))
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(loaded.Public().(ed25519.PublicKey), msg, Sign(loaded, msg)) {
		t.Fatal("loaded key failed")
	}
	if _, err := LoadEncrypted(path, []byte("wrong")); err == nil {
		t.Fatal("expected wrong passphrase failure")
	}
}

func TestManifestStatusAndArchivedReset(t *testing.T) {
	dir := t.TempDir()
	if Classify(dir) != StatusMissing {
		t.Fatal("expected missing")
	}
	pair, _ := Generate()
	if err := SaveEncrypted(filepath.Join(dir, "device.key.json"), pair.Private, []byte("pass")); err != nil {
		t.Fatal(err)
	}
	if Classify(dir) != StatusInconsistent {
		t.Fatal("expected inconsistent")
	}
	if err := WriteManifest(filepath.Join(dir, "device.manifest.json"), "device", pair.Public); err != nil {
		t.Fatal(err)
	}
	if Classify(dir) != StatusReady {
		t.Fatal("expected ready")
	}
	archive, err := ArchivedReset(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatal(err)
	}
}

func TestRepairRotateAndRevoke(t *testing.T) {
	dir := t.TempDir()
	pair, _ := Generate()
	if err := SaveEncrypted(filepath.Join(dir, "device.key.json"), pair.Private, []byte("pass")); err != nil {
		t.Fatal(err)
	}
	if err := RepairManifest(dir, []byte("pass")); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(dir, "device.manifest.json")
	if Classify(dir) != StatusReady {
		t.Fatal("expected repaired identity")
	}
	rotated, err := RotateDevice(dir, []byte("pass"))
	if err != nil {
		t.Fatal(err)
	}
	if rotated.Fingerprint == Fingerprint(pair.Public) {
		t.Fatal("rotation did not change device key")
	}
	if err := RevokeManifest(manifestPath); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "revoked_at") {
		t.Fatal("revocation marker missing")
	}
}
