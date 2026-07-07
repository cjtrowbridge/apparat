package cluster

import (
	"context"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/database"
)

func TestDirectoryPersistsProfilesAndChanges(t *testing.T) {
	db, err := database.Open(context.Background(), t.TempDir()+"/apparat.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	dir := New(db.SQL)
	if err := dir.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	profile := DeviceProfile{ID: "device-1", Signature: "sig", Roles: []string{"service-host"}, Capabilities: []Capability{{WorkloadClass: "text_generation", Runtime: "mock", Healthy: true}}, Reachable: true}
	if err := dir.PutDevice(context.Background(), profile); err != nil {
		t.Fatal(err)
	}
	got, err := dir.GetDevice(context.Background(), "device-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Capabilities[0].WorkloadClass != "text_generation" {
		t.Fatalf("profile = %+v", got)
	}
	changes, err := dir.ChangesAfter(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].DeviceID != "device-1" {
		t.Fatalf("changes = %+v", changes)
	}
}
