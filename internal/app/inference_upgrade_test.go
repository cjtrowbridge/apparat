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
	entered := make(chan struct{}, 2)
	firstChunk := make(chan struct{})
	finish := make(chan struct{})
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
		case "/api/chat":
			entered <- struct{}{}
			select {
			case <-firstChunk:
			case <-time.After(6 * time.Second):
				return
			}
			fmt.Fprintln(w, `{"message":{"role":"assistant","content":"one "},"done":false}`)
			w.(http.Flusher).Flush()
			select {
			case <-finish:
			case <-time.After(6 * time.Second):
				return
			}
			fmt.Fprint(w, `{"message":{"role":"assistant","content":"reply"},"done_reason":"stop","done":true,"eval_count":2,"eval_duration":100000000}`)
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
	session, err := node.p.runtime.Chat(ctx, work[0].Project, id)
	mustInference(t, err)
	job := session.Turns[0].PendingJob()
	dispatched := make(chan bool, 1)
	go func() {
		dispatched <- node.p.dispatchNext(ctx)
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("provider not started")
	}
	checkLivePreview(t, gui, user, job, firstChunk, finish)
	if !<-dispatched {
		t.Fatal("singleton did not dispatch")
	}
	node.p.runtime.reconcileChatWork(ctx, work[0])
	gui.syncPersonalPlanning(ctx)
	cached, err := db.GlobalArtifact(ctx, user, id)
	mustInference(t, err)
	if cached.Session.Turns[0].Reply != "one reply" {
		t.Fatal("reply missing from originating device with chat closed")
	}
	measurements := cached.Session.Turns[0].Measurements()
	if len(measurements) != 1 || measurements[0].Metrics["eval_count"] != 2 {
		t.Fatal("provider measurements missing from originating device")
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

func checkLivePreview(t *testing.T, gui *Runtime, user, job string, generate, finish chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	watch, leave := context.WithCancel(ctx)
	defer leave()
	frames := make(chan domain.PreviewFrame, 8)
	done := make(chan error, 1)
	start := func(watch context.Context) {
		go func() {
			done <- gui.WatchChatPreview(watch, user, job, func(f domain.PreviewFrame) {
				frames <- f
			})
		}()
	}
	next := func() domain.PreviewFrame {
		select {
		case f := <-frames:
			return f
		case e := <-done:
			t.Fatalf("preview ended early: %v", e)
		case <-ctx.Done():
			t.Fatal("preview timed out")
		}
		return domain.PreviewFrame{}
	}
	start(watch)
	f := next()
	if f.Stage != "Evaluating prompt" || f.Text != "" {
		t.Fatal("incorrect initial stage", f.Stage)
	}
	close(generate)
	f = next()
	if f.Stage != "Generating reply" || f.Text != "one " {
		t.Fatal("missing partial output")
	}
	leave()
	<-done
	start(ctx)
	f = next()
	if f.Text != "one " {
		t.Fatal("late subscriber missing prefix")
	}
	close(finish)
	err := <-done
	mustInference(t, err)
}
