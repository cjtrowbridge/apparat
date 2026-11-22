// Exercises identity continuity through wrong credentials, recovery, replacement and restart.
// Uses disposable directories and signing challenges without emitting credentials or key material.
package identity

import (
	"bytes"
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
