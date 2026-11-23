// Exercises portable personal identities across two independent data roots.
// Covers wrong credentials, signed profile edits and role-confused backups.
package identity

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPortableUserProfile(t *testing.T) {
	first := NewPersonal(t.TempDir())
	password := []byte("fixture password")
	err := first.Create("Owner", password)
	if err != nil {
		t.Fatal(err)
	}
	first.Lock()
	if first.Rename("Changed") == nil || first.Unlock([]byte("wrong")) == nil {
		t.Fatal("locked personal identity allowed modification or wrong password")
	}
	err = first.Unlock(password)
	if err != nil {
		t.Fatal(err)
	}
	before := bytes.Clone(first.key)
	defer clear(before)
	err = first.Rename("Renamed user")
	if err != nil {
		t.Fatal(err)
	}
	data, err := readPrivate(first.path)
	if err != nil || bytes.Contains(data, before) {
		t.Fatal("invalid portable backup")
	}
	var bundle personalBundle
	err = json.Unmarshal(data, &bundle)
	if err != nil || !bundle.Profile.Valid() || bundle.Profile.Profile.Revision != 2 {
		t.Fatal("missing valid revised signature")
	}
	second := NewPersonal(t.TempDir())
	if second.Import(bundle.Encrypted, password) == nil {
		t.Fatal("accepted a bare node key")
	}
	err = second.Import(data, password)
	if err != nil || !bytes.Equal(second.key, before) {
		t.Fatal("restore changed user identity", err)
	}
	if second.Create("Duplicate", password) == nil {
		t.Fatal("replaced existing user")
	}
	bundle.Profile.Profile.Name = "Forged"
	if bundle.Profile.Valid() {
		t.Fatal("accepted a forged profile")
	}
}

func TestJoiningFromAnotherDeviceRequestsDiscovery(t *testing.T) {
	personal := NewPersonal(t.TempDir())
	answers := []string{"Join my existing cluster", "Pair with another device"}
	openDiscovery := personal.Setup(func(raw string) string {
		var r request
		if json.Unmarshal([]byte(raw), &r) != nil || r.Op != "choose" || len(answers) == 0 || !slices.Contains(r.Options, answers[0]) {
			t.Fatal("unexpected join setup step")
		}
		answer, _ := json.Marshal(answers[0])
		answers = answers[1:]
		return string(answer)
	}, nil)
	if !openDiscovery || len(answers) != 0 || personal.Fingerprint() != "" {
		t.Fatal("joining should open discovery with no new personal key")
	}
}

func TestLaunchUnlockMigration(t *testing.T) {
	user := NewPersonal(t.TempDir())
	password := "portable fixture password"
	err := user.Create("Original", []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	err = user.Rename("Updated")
	if err != nil || user.ScreenName() != "Updated" {
		t.Fatal("rename not reflected")
	}
	device := NewSession(t.TempDir())
	err = device.Install(nil, "", "system", "fixture device secret", "fixture reference", true)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(filepath.Dir(user.path), "unlock.ref")
	err = atomicWrite(path, []byte("legacy reference"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	for launch := 0; launch < 2; launch++ {
		user.Lock()
		device.Lock()
		ref, _ := readPrivate(path)
		_, ok := device.wrapLocal(password)
		if ok {
			t.Fatal("locked device can wrap credentials")
		}
		loads := 0
		dialog := func(raw string) string {
			var r request
			if json.Unmarshal([]byte(raw), &r) != nil {
				t.Fatal("invalid dialog")
			}
			switch r.Op {
			case "load":
				loads++
				if r.Value == "fixture reference" {
					return `"fixture device secret"`
				}
				if r.Value == "legacy reference" {
					return `"portable fixture password"`
				}
			case "remove":
				return `""`
			}
			t.Fatal("unexpected credential request", r.Op)
			return "null"
		}
		device.Setup(dialog)
		user.Setup(dialog, device)
		if user.ScreenName() != "Updated" || loads != 2-launch {
			t.Fatal("wrong launch identity or prompt count", loads)
		}
		if launch == 1 && !strings.HasPrefix(string(ref), launchPrefix) {
			t.Fatal("legacy reference did not migrate")
		}
		device.Setup(dialog)
		user.Setup(dialog, device)
		if loads != 2-launch {
			t.Fatal("unlocked session prompted again")
		}
	}
}
