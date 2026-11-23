// Tests canonical durable identifiers and deterministic intent validation.
// Uses injected entropy and time without storage or external services.
package domain

import (
	"bytes"
	"testing"
	"time"
)

func TestULIDCanonicalAndTimeOrder(t *testing.T) {
	first, err := NewID(time.UnixMilli(0), bytes.NewReader(make([]byte, 10)))
	if err != nil || first != "00000000000000000000000000" {
		t.Fatal(first, err)
	}
	next, err := NewID(time.UnixMilli(1), bytes.NewReader(make([]byte, 10)))
	if err != nil || next != "00000000010000000000000000" || next <= first {
		t.Fatal(next, err)
	}
	for _, id := range []string{"80000000000000000000000000", "0000000000000000000000000i", "abc"} {
		if ValidID(id) {
			t.Fatal("accepted", id)
		}
	}
	{
		_, err = NewID(time.UnixMilli(-1), nil)
		if err == nil {
			t.Fatal("negative timestamp")
		}
	}
	{
		_, err = NewID(time.UnixMilli(1<<48), nil)
		if err == nil {
			t.Fatal("timestamp overflow")
		}
	}
	{
		_, err = NewID(time.UnixMilli(0), bytes.NewReader(nil))
		if err == nil {
			t.Fatal("entropy failure hidden")
		}
	}
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id, err := NewID(time.UnixMilli(1), nil)
		if err != nil || seen[id] || !ValidID(id) {
			t.Fatal(id, err)
		}
		seen[id] = true
	}
}

func TestDeviceDisplayName(t *testing.T) {
	for _, host := range []string{"localhost", "LOCALHOST.localdomain", "127.0.0.1", "::1", "android-deadbeef", ""} {
		d := PoolDevice{Host: host, ID: "abcdefgh123"}
		if d.DisplayName("") != "App device abcdefgh" {
			t.Fatal(host)
		}
	}
	w := ClusterDevice{Fingerprint: "worker-id", Host: "localhost", Roster: []PoolDevice{{ID: "worker-id", Worker: true, Host: "lab-server"}}}
	if w.DisplayName() != "lab-server" {
		t.Fatal("lost worker rename")
	}
}
