// Exercises independent worker databases, real pinned TLS, the scheduler and Ollama transport.
// An optional forwarded loopback port replaces the fixture provider for live qualification.
package app

import (
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/inventory"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type inferenceFixture struct {
	p    *inferencePool
	c    *httpapi.WorkerClient
	peer domain.InferencePeer
}

func mustInference(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func workerFixture(t *testing.T, owner *identity.Personal, device *identity.Session, port int) inferenceFixture {
	t.Helper()
	ctx := t.Context()
	key := identity.NewSession(t.TempDir())
	mustInference(t, key.Install(nil, "", "plain", "", "", true))
	path := filepath.Join(t.TempDir(), "queue.db")
	db, err := persistence.Open(ctx, path)
	mustInference(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	mustInference(t, db.Migrate(ctx))
	self := identity.Fingerprint(key.Public())
	mustInference(t, db.ConfirmOwner(ctx, self, owner.Fingerprint()))
	monitor := inventory.New(port, db.SaveOllamaPort)
	_, err = monitor.Refresh(ctx, port)
	mustInference(t, err)
	p, err := newInferencePool(ctx, db, key, monitor)
	mustInference(t, err)
	p.runtime = &Runtime{database: db, inference: p, Keys: key, state: "ready"}
	p.runtime.config.ArtifactsDir = t.TempDir()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	mustInference(t, err)
	peer := domain.InferencePeer{ID: self, Address: l.Addr().String()}
	server := httpapi.ServeWorkerListener(l, t.TempDir(), db, key, httpapi.WorkerSurface{Inference: p, Workspace: p.runtime.serveWorkspace})
	t.Cleanup(func() {
		_ = server.Close()
	})
	client, err := httpapi.OpenWorker(ctx, peer.Address, self, owner, device)
	mustInference(t, err)
	return inferenceFixture{p, client, peer}
}
func TestInferenceAgreement(t *testing.T) {
	var starts atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_, _ = fmt.Fprint(w, `{"version":"fixture"}`)
		case "/api/tags":
			_, _ = fmt.Fprint(w, `{"models":[{"name":"fixture","digest":"fixture-digest"}]}`)
		case "/api/ps":
			_, _ = fmt.Fprint(w, `{"models":[]}`)
		case "/api/show":
			_, _ = fmt.Fprint(w, `{"capabilities":["completion"]}`)
		case "/v1/chat/completions":
			starts.Add(1)
			_, _ = fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()
	port := provider.Listener.Addr().(*net.TCPAddr).Port
	var peerPort int
	model := "fixture"
	live := os.Getenv("APPARAT_QUALIFY_OLLAMA_PORT")
	if live != "" {
		var err error
		port, err = strconv.Atoi(live)
		mustInference(t, err)
		model = os.Getenv("APPARAT_QUALIFY_OLLAMA_MODEL")
		peerPort = port
	} else {
		other := httptest.NewServer(provider.Config.Handler)
		defer other.Close()
		peerPort = other.Listener.Addr().(*net.TCPAddr).Port
	}
	ctx := t.Context()
	user := identity.NewPersonal(t.TempDir())
	mustInference(t, user.Create("Fixture", []byte("test password")))
	device := identity.NewSession(t.TempDir())
	mustInference(t, device.Install(nil, "", "plain", "", "", true))
	a := workerFixture(t, user, device, port)
	b := workerFixture(t, user, device, peerPort)
	nodes := []inferenceFixture{a, b}
	r := domain.InferenceRoster{Cluster: a.c.Access.Cluster, Peers: []domain.InferencePeer{a.peer, b.peer}}
	for _, w := range nodes {
		_, err := w.c.Inference(ctx, domain.InferenceCommand{Operation: "status", Roster: r})
		mustInference(t, err)
	}
	ref := domain.ModelRef{Provider: "ollama", Name: model}
	for _, item := range a.p.inventory.Snapshot().Models {
		if item.Name == model {
			ref.Digest = item.Digest
		}
	}
	if ref.Digest == "" {
		t.Fatal("requested model unavailable")
	}
	var jobs []domain.InferenceJob
	for i := range 8 {
		id, err := domain.NewID(time.Now().Add(time.Duration(i)*time.Millisecond), nil)
		mustInference(t, err)
		pool := domain.DefaultPools()[3-i%4].Name
		j := domain.InferenceJob{Revision: r.Revision(), Owner: a.c.Access.Device, User: a.c.Access.User, Pool: pool, Packet: domain.ChatRequest{ChatAddress: domain.ChatAddress{Session: id, Turn: id}, Job: id, Model: ref, Messages: []domain.ChatMessage{{Role: "user", Text: "Reply with only OK."}}, MaxOutputTokens: 256}}
		jobs = append(jobs, j)
		_, err = a.c.Inference(ctx, domain.InferenceCommand{Operation: "submit", Job: j})
		mustInference(t, err)
	}
	if a.p.dispatchNext(ctx) {
		t.Fatal("started before peer admitted any job")
	}
	for _, j := range jobs {
		_, err := b.c.Inference(ctx, domain.InferenceCommand{Operation: "submit", Job: j})
		mustInference(t, err)
	}
	domain.OrderInferenceJobs(jobs)
	for i, j := range jobs {
		if !nodes[i%2].p.dispatchNext(ctx) {
			t.Fatal("no progress at priority", j.Pool)
		}
		for _, node := range nodes {
			v, err := node.c.Inference(ctx, domain.InferenceCommand{Operation: "result", ID: j.Packet.Job})
			mustInference(t, err)
			if v.State != "finished" || v.Output == nil || v.Output.Finish != "stop" || v.Output.Text == "" {
				t.Fatal("missing provider completion at priority", j.Pool, v.State, v.Output)
			}
		}
		if live != "" {
			t.Log("live ordered completion", j.Pool)
			break
		}
	}
	if live != "" {
		return
	}
	if starts.Load() != 8 {
		t.Fatal("unexpected provider start count", starts.Load())
	}
	id, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	j := jobs[0]
	j.Packet.Job = id
	for _, w := range nodes {
		_, err = w.c.Inference(ctx, domain.InferenceCommand{Operation: "submit", Job: j})
		mustInference(t, err)
	}
	for range 10 {
		var group sync.WaitGroup
		for _, w := range nodes {
			group.Go(func() {
				w.p.dispatchNext(ctx)
			})
		}
		group.Wait()
		if starts.Load() > 8 {
			break
		}
	}
	if starts.Load() != 9 {
		t.Fatal("contending providers failed exactly-one start", starts.Load())
	}
	for _, w := range nodes {
		recovered, e := newInferencePool(ctx, w.p.db, w.p.keys, w.p.inventory)
		mustInference(t, e)
		if recovered.dispatchNext(ctx) {
			t.Fatal("dispatched completed job after restart")
		}
	}
	checkInferenceRecovery(t, nodes, r, jobs[0])
	exerciseProjectQueue(t, nodes, ref)
	writer := r.GlobalWriter(user.Fingerprint(), id)
	idx := slices.Index(r.IDs(), writer)
	q := domain.InferenceCommand{Operation: "global-post", Global: domain.GlobalRequest{Session: id, Turn: id, Title: "Qualification", Input: "Reply with OK.", Model: ref}}
	_, err = nodes[idx].c.Inference(ctx, q)
	mustInference(t, err)
	work, err := nodes[idx].p.db.ChatWork(ctx)
	mustInference(t, err)
	if len(work) != 1 || work[0].Pool != "Chat Pool" {
		t.Fatal("global turn not queued")
	}
	n := nodes[idx].p.runtime
	n.reconcileChatWork(ctx, work[0])
	if !nodes[1-idx].p.dispatchNext(ctx) {
		t.Fatal("global request unavailable to peer")
	}
	n.reconcileChatWork(ctx, work[0])
	nodes[idx].p.publishGlobals(ctx)
	v, err := nodes[1-idx].c.Inference(ctx, domain.InferenceCommand{Operation: "global-get", Global: domain.GlobalRequest{Session: id}})
	mustInference(t, err)
	if v.Artifact == nil || v.Artifact.Session.Turns[0].Reply != "OK" {
		t.Fatal("global reply failed replication")
	}
	checkPersonalResume(t, nodes[idx], nodes[1-idx], user, id, ref)
	checkGlobalChunks(t, nodes[idx], nodes[1-idx], user.Fingerprint(), id)
	for _, node := range nodes {
		node.p.runtime.mirrorGlobalMarkdown(ctx)
		name, e := projectgit.ChatFile(id)
		mustInference(t, e)
		b, e := os.ReadFile(filepath.Join(node.p.runtime.config.ArtifactsDir, "global", user.Fingerprint(), ".apparat", "chats", name))
		mustInference(t, e)
		if !strings.Contains(string(b), "OK") {
			t.Fatal("missing global Markdown reply")
		}
	}

}

func exerciseProjectQueue(t *testing.T, nodes []inferenceFixture, model domain.ModelRef) {
	t.Helper()
	a, b := nodes[0], nodes[1]
	ctx := t.Context()
	project, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	repo := t.TempDir()
	mustInference(t, a.p.db.RegisterProject(ctx, domain.Project{ID: project, Name: "Acceptance project", Location: repo}))
	id, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	q := domain.WorkspaceRequest{Path: ".", ProjectID: project, Action: "agent-post", ID: id, Thread: id, Argument: "Review project. Reply with OK", Model: model}
	for range 2 {
		view, err := a.c.Workspace(ctx, q)
		mustInference(t, err)
		if view.Detail != "Review project." || len(view.Choices) != 2 {
			t.Fatal("first submission or retry lost the title or duplicated the turn")
		}
	}
	work, err := a.p.db.ChatWork(ctx)
	mustInference(t, err)
	if len(work) != 1 || work[0].Pool != "Project Pool" {
		t.Fatal("missing project task")
	}
	a.p.runtime.reconcileChatWork(ctx, work[0])
	if !b.p.dispatchNext(ctx) {
		t.Fatal("peer failed to dispatch the project turn")
	}
	a.p.runtime.reconcileChatWork(ctx, work[0])
	a.p.runtime.reconcileChatWork(ctx, work[0])
	q.Action = "agent-chat"
	view, err := a.c.Workspace(ctx, q)
	mustInference(t, err)
	if len(view.Choices) != 2 || view.Choices[1].State != "assistant" || view.Choices[1].Label != "OK" {
		t.Fatal("project transcript missing provider reply")
	}
}
