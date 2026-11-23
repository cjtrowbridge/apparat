// Keeps a bounded client cache of recently verified owner reads.
// A conditional read still requires fresh authorization before cached bytes are returned.
package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"sync"
	"time"
)

type cacheKey struct{ user, project, path, action string }
type cacheRow struct {
	view  domain.WorkspaceView
	stamp uint64
}
type workspaceCache struct {
	mu    sync.Mutex
	clock uint64
	rows  map[cacheKey]cacheRow
}

func (c *workspaceCache) get(k cacheKey) (domain.WorkspaceView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	row, ok := c.rows[k]
	if ok {
		c.clock++
		row.stamp = c.clock
		c.rows[k] = row
	}
	return row.view, ok
}
func (c *workspaceCache) put(k cacheKey, v domain.WorkspaceView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rows == nil {
		c.rows = map[cacheKey]cacheRow{}
	}
	if len(c.rows) >= 64 {
		var oldest cacheKey
		stamp := ^uint64(0)
		for key, r := range c.rows {
			if r.stamp < stamp {
				oldest = key
				stamp = r.stamp
			}
		}
		delete(c.rows, oldest)
	}
	c.clock++
	c.rows[k] = cacheRow{v, c.clock}
}
func conditional(q domain.WorkspaceRequest, v domain.WorkspaceView, err error) (domain.WorkspaceView, error) {
	if err != nil || v.Cursor != "" {
		return v, err
	}
	if q.Action == "list" {
		raw, _ := json.Marshal(v.Entries)
		v.Base = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	if q.Known != "" && v.Base == q.Known {
		return domain.WorkspaceView{Unchanged: true, Base: v.Base}, nil
	}
	return v, nil
}

// Active workspace operations share a short-lived device-bound credential for the selected host.
type ownerConnection struct {
	mu     sync.Mutex
	client *httpapi.WorkerClient
	until  time.Time
}

func (n *Runtime) workspaceClient(ctx context.Context, host string) (*httpapi.WorkerClient, error) {
	c := &n.workspaceConnection
	c.mu.Lock()
	defer c.mu.Unlock()
	user := n.Personal.Fingerprint()
	if user == "" || n.Keys.State() != "unlocked" {
		return nil, httpapi.ErrWorkspaceLocked
	}
	if c.client != nil && "worker-"+c.client.Worker == host && c.client.Access.User == user && c.client.Access.Device == identity.Fingerprint(n.Keys.Public()) && time.Now().Before(c.until) {
		return c.client, nil
	}
	client, err := n.projectHost(ctx, host)
	if err == nil {
		c.client = client
		c.until = time.Now().Add(4 * time.Minute)
	}
	return client, err
}
