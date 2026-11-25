// Exercises two device-bound clients and a reopened worker database and listener.
// The second client enters Cluster using the ordinary background discovery path.
package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/inventory"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func TestInventoryRestoredAcrossDevices(t *testing.T) {
	ctx := t.Context()
	meta := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			fmt.Fprint(w, `{"version":"fixture"}`)
		case "/api/tags":
			fmt.Fprint(w, `{"models":[{"name":"shared","digest":"abc","size":100}]}`)
		case "/api/ps":
			fmt.Fprint(w, `{"models":[]}`)
		case "/api/show":
			fmt.Fprint(w, `{"capabilities":["completion"]}`)
		default:
			t.Error("unexpected metadata request", r.URL.Path)
		}
	}))
	defer meta.Close()
	port := meta.Listener.Addr().(*net.TCPAddr).Port
	root := t.TempDir()
	key := fixtureDevice(t)
	pin := identity.Fingerprint(key.Public())
	actor := fixtureActor(t, "Owner")
	open := func() (*persistence.DB, *inventory.Monitor, *httpapi.Server, string) {
		db, err := persistence.Open(ctx, filepath.Join(root, "worker.db"))
		requireWorkspaceOK(t, err)
		requireWorkspaceOK(t, db.Migrate(ctx))
		saved, err := db.OllamaPort(ctx)
		requireWorkspaceOK(t, err)
		monitor := inventory.New(saved, db.SaveOllamaPort, db.CPUBaseline)
		l, err := net.Listen("tcp", "127.0.0.1:0")
		requireWorkspaceOK(t, err)
		server := httpapi.ServeWorkerListener(l, root, db, key, httpapi.WorkerSurface{Inventory: monitor})
		return db, monitor, server, l.Addr().String()
	}
	db, _, server, addr := open()
	requireWorkspaceOK(t, db.ConfirmOwner(ctx, pin, actor.person.Fingerprint()))
	first, err := httpapi.OpenWorker(ctx, addr, pin, actor.person, actor.device)
	requireWorkspaceOK(t, err)
	firstID := identity.Fingerprint(actor.device.Public())
	_, err = first.MemberHardware(ctx, domain.HardwareCapacity{Threads: 24})
	requireWorkspaceOK(t, err)
	_, err = first.TryOllama(ctx, port)
	requireWorkspaceOK(t, err)
	requireWorkspaceOK(t, server.Close())
	requireWorkspaceOK(t, db.Close())

	db, monitor, server, addr := open()
	defer db.Close()
	defer server.Close()
	if !monitor.Snapshot().Configured || monitor.Snapshot().Port != port {
		t.Fatal("lost saved port before first probe")
	}
	_, err = monitor.Refresh(ctx, 0)
	requireWorkspaceOK(t, err)
	secondDB, err := persistence.Open(ctx, filepath.Join(t.TempDir(), "client.db"))
	requireWorkspaceOK(t, err)
	defer secondDB.Close()
	requireWorkspaceOK(t, secondDB.Migrate(ctx))
	second := &Runtime{database: secondDB, Personal: actor.person, Keys: fixtureDevice(t), inventoryHint: make(chan struct{}, 1)}
	poll, cancel := context.WithCancel(ctx)
	second.discoveryWorkers.Add(1)
	go second.refreshInventoryPeers(poll)
	defer func() {
		cancel()
		second.discoveryWorkers.Wait()
	}()
	secondClient, err := httpapi.OpenWorker(ctx, addr, pin, actor.person, second.Keys)
	requireWorkspaceOK(t, err)
	reports, err := secondClient.MemberHardware(ctx, domain.HardwareCapacity{Threads: 8})
	requireWorkspaceOK(t, err)
	if reports[firstID].Threads != 24 || reports[secondClient.Access.Device].Threads != 8 {
		t.Fatal("paired device capacities missing after restart", reports)
	}
	requireWorkspaceOK(t, secondDB.SaveWorkerAccess(ctx, "worker", addr, pin, secondClient.Access))
	second.RequestSync()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		rows, err := second.KnownWorkers(ctx)
		requireWorkspaceOK(t, err)
		if len(rows) == 1 && rows[0].Inventory != nil {
			v := rows[0].Inventory
			if rows[0].Hardware[firstID].Threads == 24 && rows[0].Hardware[pin].Threads > 0 && v.Configured && v.Port == port && v.HasText(time.Now()) && v.Models[0].Name == "shared" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("second device failed to discover the configured worker on entry")
}
