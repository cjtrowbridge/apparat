// Exercises identity continuity through wrong credentials, recovery, replacement and restart.
// Uses disposable directories and signing challenges without emitting credentials or key material.
package identity

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupContinuity(t *testing.T) {
	root := t.TempDir()
	s := NewSession(root)
	s.Setup(func(string) string {
		return "null"
	})
	if s.State() != "missing" {
		t.Fatal("cancel created an identity")
	}
	err := s.Install(nil, "", "password", "sample-password", "", true)
	if err != nil {
		t.Fatal(err)
	}
	_, first, _ := current(root)
	key, err := os.ReadFile(filepath.Join(first, "device.key.json"))
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("local challenge")
	want, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	s.Lock()
	_, err = s.Sign(msg)
	if err == nil || s.Unlock("incorrect") == nil || s.State() != "password" {
		t.Fatal("wrong password or locked signing accepted")
	}
	restarted := NewSession(root)
	if restarted.Unlock("sample-password") != nil {
		t.Fatal("restart lost key")
	}
	got, _ := restarted.Sign(msg)
	if !bytes.Equal(want, got) {
		t.Fatal("restart changed identity")
	}
	restarted.Lock()
	err = os.Remove(filepath.Join(first, "device.key.json"))
	if err != nil || restarted.State() != "recovery" {
		t.Fatal("partial key pair was not recognized")
	}
	other := NewSession(t.TempDir())
	err = other.Install(nil, "", "password", "other", "", true)
	if err != nil {
		t.Fatal(err)
	}
	_, wrongDir, _ := current(other.dir)
	wrong, _ := os.ReadFile(filepath.Join(wrongDir, "device.key.json"))
	if restarted.Install(wrong, "other", "password", "new", "", false) == nil {
		t.Fatal("accepted unrelated backup")
	}
	err = restarted.Install(key, "sample-password", "system", "system-secret", "reference", false)
	if err != nil {
		t.Fatal(err)
	}
	got, _ = restarted.Sign(msg)
	if !bytes.Equal(want, got) {
		t.Fatal("restore changed identity")
	}
	restarted.Lock()
	if restarted.State() != "system" || restarted.Reference() != "reference" || restarted.Unlock("system-secret") != nil {
		t.Fatal("native credential did not unlock")
	}
	err = restarted.Install(nil, "", "plain", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	got, _ = restarted.Sign(msg)
	if bytes.Equal(want, got) {
		t.Fatal("confirmed replacement did not change identity")
	}
	fresh := NewSession(root)
	if fresh.State() != "plain" || fresh.Unlock("") != nil {
		t.Fatal("unattended worker did not reopen identity")
	}
}

func TestStartupUsesSavedProtection(t *testing.T) {
	for _, mode := range []string{"system", "password", "plain"} {
		dir := t.TempDir()
		created := NewSession(dir)
		err := created.Install(nil, "", mode, "test secret", "reference", true)
		if err != nil {
			t.Fatal(err)
		}
		for _, answer := range []string{"null", `"incorrect"`, `"test secret"`} {
			restarted := NewSession(dir)
			restarted.Setup(func(raw string) string {
				var req request
				if json.Unmarshal([]byte(raw), &req) != nil {
					t.Fatal("invalid request")
				}
				if req.Op == "notice" && answer == `"incorrect"` {
					return "null"
				}
				if !(mode == "system" && req.Op == "load" && req.Value == "reference" || mode == "password" && req.Op == "password") {
					t.Fatalf("%s startup unexpectedly asked %s", mode, req.Op)
				}
				return answer
			})
			unlocked := restarted.State() == "unlocked"
			if unlocked != (mode == "plain" || answer == `"test secret"`) {
				t.Fatal("incorrect startup state", mode)
			}
			if unlocked && !bytes.Equal(restarted.Public(), created.Public()) {
				t.Fatal("startup replaced saved identity")
			}
		}
	}
}

func TestMissingPointerOnlyOffersRecovery(t *testing.T) {
	dir := t.TempDir()
	s := NewSession(dir)
	err := s.Install(nil, "", "plain", "", "", true)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(filepath.Join(dir, "active.json"))
	if err != nil {
		t.Fatal(err)
	}
	restarted := NewSession(dir)
	if restarted.State() != "recovery" {
		t.Fatal("orphaned identity mistaken for a new install")
	}
	restarted.Setup(func(raw string) string {
		var req request
		if json.Unmarshal([]byte(raw), &req) != nil || len(req.Options) != 1 || req.Options[0] != "Find existing keys" {
			t.Fatal("expected recovery only")
		}
		return "null"
	})
}

func TestDeviceNameSetupPersists(t *testing.T) {
	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "hostname"), []byte("localhost"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	attempts := 0
	ok := NewSession(root).SetupHostname(func(string) string {
		attempts++
		if attempts == 1 {
			return `"LOCALHOST"`
		}
		return "null"
	})
	if ok || attempts != 2 {
		t.Fatal("invalid input should produce one notice and leave setup")
	}
	ok = NewSession(root).SetupHostname(func(string) string {
		return `"Pixel 10 XL"`
	})
	name, err := NewSession(root).Hostname()
	if !ok || attempts != 2 || name != "Pixel-10-XL" || err != nil {
		t.Fatal("name setup did not persist", err)
	}
	NewSession(root).SetupHostname(func(string) string {
		t.Fatal("existing name prompted again")
		return "null"
	})
}
