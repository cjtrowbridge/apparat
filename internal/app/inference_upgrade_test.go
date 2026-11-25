// Exercises a pre-upgrade GUI outbox against an automatically discovered singleton.
// Verifies immediate queue visibility, retry identity and maintenance handback.
package app

import (
	"context"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestAutomaticWorkerAndStrandedTurn(t *testing.T) {
	ctx := t.Context()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			fmt.Fprint(w, `{"version":"local"}`)
		case "/api/tags":
			fmt.Fprint(w, `{"models":[{"name":"fixture","digest":"test"}]}`)
		case "/api/ps":
			fmt.Fprint(w, `{"models":[]}`)
		case "/api/show":
			fmt.Fprint(w, `{"capabilities":["completion"]}`)
		case "/v1/chat/completions":
			fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"one reply"},"finish_reason":"stop"}]}`)
		}
	}))
	defer provider.Close()
	port := provider.Listener.Addr().(*net.TCPAddr).Port
	owner := identity.NewPersonal(t.TempDir())
	mustInference(t, owner.Create("owner", []byte("test-password")))
	device := identity.NewSession(t.TempDir())
	mustInference(t, device.Install(nil, "", "plain", "", "", true))
	node := workerFixture(t, owner, device, port)
	db, err := persistence.Open(ctx, filepath.Join(t.TempDir(), "client.db"))
	mustInference(t, err)
	defer db.Close()
	mustInference(t, db.Migrate(ctx))
	mustInference(t, db.SaveWorkerAccess(ctx, "worker", node.peer.Address, node.peer.ID, node.c.Access))
	gui := &Runtime{database: db, Keys: device, Personal: owner, state: "ready"}
	user := owner.Fingerprint()
	id, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	g := domain.GlobalRequest{Session: id, Turn: id, Title: "Earlier chat", Input: "Resume the original turn", Model: domain.ModelRef{Provider: "ollama", Name: "fixture", Digest: "test"}}
	mustInference(t, gui.SavePlanning(ctx, user, "global-post", g))
	entries, err := db.PlanningEntries(ctx, user)
	mustInference(t, err)
	r, clients, _ := gui.planningClients(ctx, user)
	if len(r.Peers) != 1 {
		t.Fatal("ordinary registration did not initialize singleton")
	}
	for range 2 {
		gui.syncPlanning(ctx, user, entries, r, clients)
	}
	page, err := node.c.Inference(ctx, domain.InferenceCommand{Operation: "tasks", Pool: "Chat Pool"})
	mustInference(t, err)
	if len(page.Tasks) != 1 || page.Tasks[0].ID != id || page.Tasks[0].State != "queued" {
		t.Fatal("lost or duplicated durable receipt")
	}
	node.p.applyReceipts(ctx)
	work, err := node.p.db.ChatWork(ctx)
	mustInference(t, err)
	if len(work) != 1 {
		t.Fatal("expected one owner turn")
	}
	node.p.runtime.reconcileChatWork(ctx, work[0])
	gui.syncPersonalPlanning(ctx)
	entries, err = db.PlanningEntries(ctx, user)
	mustInference(t, err)
	if len(entries) != 0 {
		t.Fatal("worker did not acknowledge the pending turn")
	}
	if !node.p.dispatchNext(ctx) {
		t.Fatal("singleton did not dispatch")
	}
	node.p.runtime.reconcileChatWork(ctx, work[0])
	gui.syncPersonalPlanning(ctx)
	cached, err := db.GlobalArtifact(ctx, user, id)
	mustInference(t, err)
	if cached.Session.Turns[0].Reply != "one reply" {
		t.Fatal("reply missing from originating device with chat closed")
	}
	view, err := gui.PlanningView(ctx, user, id)
	mustInference(t, err)
	if view.Session == nil || len(view.Session.Turns) != 1 || view.Session.Turns[0].Reply != "one reply" || !view.Acknowledged {
		t.Fatal("missing original turn result")
	}
	peer := workerFixture(t, owner, device, port)
	mustInference(t, db.SaveWorkerAccess(ctx, "second", peer.peer.Address, peer.peer.ID, peer.c.Access))
	r, _, _ = gui.planningClients(ctx, user)
	if len(r.Peers) != 2 {
		t.Fatal("joining worker did not participate automatically")
	}
	if !peer.p.runtime.PrepareInferenceUpdate(ctx) {
		t.Fatal("could not withdraw idle worker")
	}
	checkMaintenanceDispatch(t, ctx, node, peer, r, g)
}

func checkMaintenanceDispatch(t *testing.T, ctx context.Context, node, peer inferenceFixture, r domain.InferenceRoster, g domain.GlobalRequest) {
	t.Helper()
	id, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	j := domain.InferenceJob{Revision: r.Revision(), Owner: node.c.Access.Device, User: node.c.Access.User, Pool: "Chat Pool", Packet: domain.ChatRequest{ChatAddress: domain.ChatAddress{Session: g.Session, Turn: g.Turn}, Job: id, Model: g.Model, MaxOutputTokens: 128}}
	_, err = node.c.Inference(ctx, domain.InferenceCommand{Operation: "submit", Job: j})
	mustInference(t, err)
	if peer.p.dispatchNext(ctx) || !node.p.dispatchNext(ctx) {
		t.Fatal("remaining worker did not serve during update")
	}
	peer.p.runtime.CancelInferenceUpdate(ctx)
	if len(peer.p.votingRoster().Peers) != 2 {
		t.Fatal("worker did not rejoin")
	}
	record, err := peer.p.db.InferenceRecord(ctx, id)
	mustInference(t, err)
	if record.State != "finished" || peer.p.dispatchNext(ctx) {
		t.Fatal("missing completion fence after rejoin")
	}
}
