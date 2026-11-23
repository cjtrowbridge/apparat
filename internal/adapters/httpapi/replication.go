// Batches one reader's policy snapshot and pool queue cursors over authenticated TLS.
// The no-change response is an empty JSON object; queue pages are bounded per request.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (a *workerAuth) reconcile(w http.ResponseWriter, r *http.Request, access persistence.Access, keys *identity.Session) {
	var cursor persistence.SyncCursor
	if r.Method != http.MethodPost || !boundedJSON(w, r, &cursor) || len(cursor.Queues) > 32 {
		respond(w, 400, map[string]string{"error": "invalid_sync_request"})
		return
	}
	view, err := a.db.ReadReplica(r.Context(), access, identity.Fingerprint(keys.Public()), cursor.Revision, r.Header.Get("Apparat-Device-Directory") == "1")
	if err != nil {
		respond(w, 403, map[string]string{"error": "access_denied"})
		return
	}
	page := persistence.SyncPage{}
	if view != nil {
		body, err := json.Marshal(view)
		if err != nil {
			respond(w, 500, map[string]string{"error": "snapshot_unavailable"})
			return
		}
		sig, err := keys.Sign(body)
		if err != nil {
			respond(w, 503, map[string]string{"error": "worker_locked"})
			return
		}
		page.Snapshot = &persistence.SignedReplica{View: *view, Key: keys.Public(), Signature: sig}
	} else {
		page.Queues = map[string][]persistence.QueueChange{}
		for pool, after := range cursor.Queues {
			changes, err := a.db.QueueAfter(r.Context(), access, pool, after)
			if err != nil {
				respond(w, 403, map[string]string{"error": "access_denied"})
				return
			}
			if len(changes) > 0 {
				page.Queues[pool] = changes
			}
		}
	}
	respond(w, 200, page)
}

func (c *WorkerClient) Sync(ctx context.Context, db *persistence.DB) (int, error) {
	r, cursor, err := db.LocalReplica(ctx, c.Worker, c.Access)
	if err != nil {
		return 0, err
	}
	pools := append([]string(nil), r.View.Pools...)
	sort.Strings(pools)
	for start := 0; start == 0 || start < len(pools); start += 32 {
		batch := persistence.SyncCursor{Revision: cursor.Revision, Queues: map[string]int64{}}
		if c.directory && len(r.View.Devices) == 0 {
			batch.Revision = 0
		}
		for _, pool := range pools[start:min(start+32, len(pools))] {
			batch.Queues[pool] = cursor.Queues[pool]
		}
		var page persistence.SyncPage
		err = c.call(ctx, "POST", "/v1/sync", batch, &page)
		if err != nil {
			return 0, err
		}
		err = db.ApplyReplica(ctx, c.Worker, c.Access, cursor, page)
		if err != nil {
			return 0, err
		}
		if page.Snapshot != nil {
			return 0, nil
		}
		for _, changes := range page.Queues {
			if len(changes) == 64 {
				return 0, nil
			}
		}
	}
	if r.View.Admin && r.View.Governance != nil {
		start, err := db.ReplicaPolicyCursor(ctx, c.Worker, c.Access)
		if err != nil {
			return 0, err
		}
		if start != r.View.Governance.Head {
			page, err := c.Policy(ctx, start)
			if err != nil {
				return 0, err
			}
			err = db.ApplyReceiptPage(ctx, c.Worker, c.Access, r, start, page)
			if err != nil {
				return 0, err
			}
			return 0, nil
		}
	}
	return r.View.PollSeconds, nil
}
