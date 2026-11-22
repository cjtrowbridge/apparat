// Exercises portable personal identities across two independent data roots.
// Covers wrong credentials, signed profile edits and role-confused backups.
package identity

import (
	"bytes"
	"encoding/json"
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
