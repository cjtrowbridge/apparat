// Validates per-user queue views, re-leasing, active fencing and artifact references.
// Exercises the same production schema and transition methods used by a worker.
package persistence

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestHashTaskFencing(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, t.TempDir()+"/queue.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Migrate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.sql.ExecContext(ctx, `
	INSERT INTO worker_cluster(singleton,cluster,worker,owner_user,authority_device) VALUES(1,'c','w','u','d');
	INSERT INTO cluster_users VALUES('c','u',X'00','approved',1),('c','friend',X'00','approved',0);
	INSERT INTO cluster_devices_authorized(cluster,user_id,device_id,public_binding) VALUES('c','u','d',X'00'),('c','friend','fd',X'00');
	INSERT INTO cluster_pools VALUES('c','p');
	INSERT INTO pool_services VALUES('c','p','w','hash-text');`)
	if err != nil {
		t.Fatal(err)
	}
	owner := Access{Cluster: "c", User: "u", Device: "d"}
	friend := Access{Cluster: "c", User: "friend", Device: "fd"}
	if db.SubmitHash(ctx, friend, "p", "unexpected", "") == nil {
		t.Fatal("friend submitted without grant")
	}
	err = db.SubmitHash(ctx, owner, "p", "id", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if db.SubmitHash(ctx, owner, "p", "id", "different") == nil {
		t.Fatal("duplicate submission changed prompt")
	}
	now := time.Now()
	first, err := db.LeaseHash(ctx, "c", "p", "w", now)
	if err != nil {
		t.Fatal(err)
	}
	next := now.Add(2 * time.Minute)
	second, err := db.LeaseHash(ctx, "c", "p", "w", next)
	if err != nil || second.Fence <= first.Fence {
		t.Fatal("expired job was not re-leased", err)
	}
	artifact := fmt.Sprintf("%x\n", sha256.Sum256([]byte("hello")))
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(artifact)))
	if db.CompleteHash(ctx, first, "w", hash, next) == nil || db.CompleteHash(ctx, second, "w", "invalid", next) == nil {
		t.Fatal("accepted stale or invalid result")
	}
	err = db.CompleteHash(ctx, second, "w", hash, next)
	if err != nil {
		t.Fatal(err)
	}
	got, err := db.ResultReference(ctx, owner, "p", "id")
	if err != nil || got != hash {
		t.Fatal("missing result", err)
	}
	_, err = db.ResultReference(ctx, friend, "p", "id")
	if err == nil {
		t.Fatal("result leaked to second user")
	}
	changes, err := db.QueueAfter(ctx, owner, "p", 0)
	if err != nil || len(changes) != 4 || changes[3].State != "succeeded" {
		t.Fatal("missing task history", err)
	}
	changes, err = db.QueueAfter(ctx, friend, "p", 0)
	if err == nil || len(changes) != 0 {
		t.Fatal("task history leaked", err)
	}

	err = db.CompleteHash(ctx, second, "w", hash, next)
	if err != nil {
		t.Fatal("exact result retry failed", err)
	}
	_, err = db.sql.ExecContext(ctx, "INSERT INTO pool_limits VALUES('c','p',1)")
	if err != nil {
		t.Fatal(err)
	}
	err = db.SubmitHash(ctx, owner, "p", "quota-test", "first")
	if err != nil {
		t.Fatal(err)
	}
	if db.SubmitHash(ctx, owner, "p", "too-many", "second") == nil {
		t.Fatal("pool quota exceeded")
	}
	_, err = db.sql.ExecContext(ctx, "INSERT INTO local_runtime VALUES('worker_update_installing','1',0)")
	if err != nil {
		t.Fatal(err)
	}
	err = db.SubmitHash(ctx, owner, "p", "quota-test", "first")
	if err != nil {
		t.Fatal("draining broke exact retry", err)
	}
	_, err = db.LeaseHash(ctx, "c", "p", "w", next)
	if err == nil {
		t.Fatal("new lease during worker update")
	}
	err = db.loadQueueHeads(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.sql.ExecContext(ctx, "ALTER TABLE pool_queue_events RENAME TO hidden_events")
	if err != nil {
		t.Fatal(err)
	}
	changes, err = db.QueueAfter(ctx, owner, "p", 5)
	if err != nil || len(changes) != 0 {
		t.Fatal("unchanged cursor read the event table", err)
	}

}

func TestLocalPoolTaskLatestVisibleState(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	for _, row := range []struct{ worker, user, device, pool, body string }{
		{"w", "u", "d", "p", `{"ID":"one","State":"queued"}`},
		{"w", "u", "d", "p", `{"ID":"one","State":"succeeded"}`},
		{"w", "u", "d", "p", `{"ID":"two","State":"running"}`},
		{"w", "other", "d", "p", `{"ID":"hidden"}`},
		{"w", "u", "other", "p", `{"ID":"hidden"}`},
		{"other", "u", "d", "p", `{"ID":"hidden"}`},
		{"w", "u", "d", "other", `{"ID":"hidden"}`},
	} {
		_, err := db.sql.ExecContext(ctx, "INSERT INTO replica_queue SELECT ?,?,?,?,COUNT(*)+1,? FROM replica_queue", row.worker, row.user, row.device, row.pool, []byte(row.body))
		if err != nil {
			t.Fatal(err)
		}
	}
	rows, err := db.LocalPoolTasks(ctx, "w", Access{User: "u", Device: "d"}, "p")
	if err != nil || len(rows) != 2 {
		t.Fatalf("expected two visible tasks, got %v, %v", rows, err)
	}
	got := map[string]string{}
	for _, r := range rows {
		got[r.ID] = r.State
	}
	if got["one"] != "succeeded" || got["two"] != "running" {
		t.Fatal(got)
	}
}
