// Fault-injects unreachable peers and a process loss after the provider-start receipt.
// Verifies recovery and large global transcripts against the same independent TLS workers.
package app

import (
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"net"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func checkInferenceRecovery(t *testing.T, nodes []inferenceFixture, r domain.InferenceRoster, j domain.InferenceJob) {
	t.Helper()
	a, b := nodes[0], nodes[1]
	ctx := t.Context()
	job, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	j.Packet.Job = job
	for _, w := range nodes {
		_, err = w.c.Inference(ctx, domain.InferenceCommand{Operation: "submit", Job: j})
		mustInference(t, err)
	}
	unreachable, err := net.Listen("tcp", "127.0.0.1:0")
	mustInference(t, err)
	address := unreachable.Addr().String()
	mustInference(t, unreachable.Close())
	partition := domain.InferenceRoster{Cluster: r.Cluster, Peers: slices.Clone(r.Peers)}
	partition.Peers[1].Address = address
	mustInference(t, a.p.install(ctx, partition))
	if a.p.dispatchNext(ctx) {
		t.Fatal("dispatched during partition")
	}
	status, err := a.c.Inference(ctx, domain.InferenceCommand{Operation: "status"})
	mustInference(t, err)
	if status.Status.Peers[b.peer.ID].Received || status.Status.Timing.Incomplete == 0 || status.Status.BlockedSince == 0 {
		t.Fatal("missing partition telemetry")
	}
	mustInference(t, a.p.install(ctx, r))
	attempt, err := domain.NewID(time.Now().Add(time.Second), nil)
	mustInference(t, err)
	c := domain.InferenceClaim{Job: job, Revision: j.Revision, Digest: domain.InferenceDigest(j), Executor: a.peer.ID, Attempt: attempt}
	for _, op := range []string{"confirm", "own"} {
		ans := a.p.fanout(ctx, r, domain.InferenceCommand{Operation: op, Claim: c})
		if len(ans) != 2 {
			t.Fatal("handoff failed")
		}
	}
	mustInference(t, a.p.db.StartInference(ctx, c))
	reopened, err := newInferencePool(ctx, a.p.db, a.p.keys, a.p.inventory)
	mustInference(t, err)
	if !reopened.dispatchNext(ctx) {
		t.Fatal("did not publish interrupted dispatch")
	}
	for _, w := range nodes {
		v, e := w.c.Inference(ctx, domain.InferenceCommand{Operation: "result", ID: job})
		mustInference(t, e)
		if v.Output == nil || v.Output.Finish != "indeterminate" {
			t.Fatal("wrong crash outcome", v)
		}
	}
}
func checkGlobalChunks(t *testing.T, writer, peer inferenceFixture, user, id string) {
	t.Helper()
	ctx := t.Context()
	a, err := writer.p.db.GlobalArtifact(ctx, user, id)
	mustInference(t, err)
	for range 40 {
		turn, e := domain.NewID(time.Now(), nil)
		mustInference(t, e)
		_, e = a.Session.Apply(turn, strings.Repeat("large fixture ", 2200), false)
		mustInference(t, e)
		_, e = a.Session.Apply(turn, "OK", true)
		mustInference(t, e)
	}
	a.Version++
	mustInference(t, writer.p.db.PutGlobalArtifact(ctx, a, true))
	writer.p.publishGlobals(ctx)
	v, err := peer.c.Inference(ctx, domain.InferenceCommand{Operation: "global-get", Global: domain.GlobalRequest{Session: id}})
	mustInference(t, err)
	if v.Artifact == nil || domain.InferenceDigest(v.Artifact.Session) != domain.InferenceDigest(a.Session) {
		t.Fatal("incomplete long global transcript")
	}
}

func checkPersonalResume(t *testing.T, writer, replica inferenceFixture, user *identity.Personal, id string, model domain.ModelRef) {
	t.Helper()
	ctx := t.Context()
	key := identity.NewSession(t.TempDir())
	mustInference(t, key.Install(nil, "", "plain", "", "", true))
	db, err := persistence.Open(ctx, filepath.Join(t.TempDir(), "phone.db"))
	mustInference(t, err)
	defer func() {
		_ = db.Close()
	}()
	mustInference(t, db.Migrate(ctx))
	mustInference(t, db.SaveWorkerAccess(ctx, "replica", replica.peer.Address, replica.peer.ID, replica.c.Access))
	gui := &Runtime{database: db, Keys: key, Personal: user, state: "ready"}
	view, err := gui.PlanningView(ctx, user.Fingerprint(), id)
	mustInference(t, err)
	if view.Session == nil || view.Session.Turns[0].Reply != "OK" || !view.Acknowledged {
		t.Fatal("second device missed global reply")
	}
	turn, err := domain.NewID(time.Now(), nil)
	mustInference(t, err)
	q := domain.GlobalRequest{Session: id, Turn: turn, Title: view.Session.Description, Input: "list projects", Model: model}
	mustInference(t, gui.SavePlanning(ctx, user.Fingerprint(), "global-post", q))
	entries, err := db.PlanningEntries(ctx, user.Fingerprint())
	mustInference(t, err)
	roster, clients := gui.planningClients(ctx, user.Fingerprint())
	gui.syncPlanning(ctx, user.Fingerprint(), entries, roster, clients)
	entries, err = db.PlanningEntries(ctx, user.Fingerprint())
	mustInference(t, err)
	if len(entries) != 0 {
		t.Fatal("could not continue session through its enrolled writer")
	}
	writer.p.publishGlobals(ctx)
	view, err = gui.PlanningView(ctx, user.Fingerprint(), id)
	mustInference(t, err)
	last := view.Session.Turns[len(view.Session.Turns)-1]
	if last.ID != turn || !strings.Contains(last.Reply, "Acceptance project") || len(last.Steps) != 0 {
		t.Fatal("deterministic resumed turn failed")
	}
}
