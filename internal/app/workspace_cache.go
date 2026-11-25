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
	"golang.org/x/sync/singleflight"
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

// All callers for a device share its short-lived credential, including concurrent sessions.
type ownerConnection struct {
	mu      sync.Mutex
	entries map[string]cachedCredential
	open    singleflight.Group
}
type cachedCredential struct {
	client *httpapi.WorkerClient
	until  time.Time
}

func (n *Runtime) workspaceClient(ctx context.Context, host string) (*httpapi.WorkerClient, error) {
	return n.cachedWorker(ctx, host, func() (*httpapi.WorkerClient, error) {
		return n.projectHost(ctx, host)
	})
}
func (n *Runtime) cachedWorker(ctx context.Context, host string, connect func() (*httpapi.WorkerClient, error)) (*httpapi.WorkerClient, error) {
	user := n.Personal.Fingerprint()
	if user == "" || n.Keys.State() != "unlocked" {
		return nil, httpapi.ErrWorkspaceLocked
	}
	key := user + "/" + identity.Fingerprint(n.Keys.Public()) + "/" + host
	cache := &n.workspaceConnection
	cache.mu.Lock()
	v := cache.entries[key]
	cache.mu.Unlock()
	if v.client != nil && time.Now().Before(v.until) {
		return v.client, nil
	}
	value, err, _ := cache.open.Do(key, func() (any, error) {
		cache.mu.Lock()
		old := cache.entries[key]
		cache.mu.Unlock()
		if old.client != nil && time.Now().Before(old.until) {
			return old.client, nil
		}
		c, err := connect()
		if err != nil {
			return nil, err
		}
		if c.Access.User != user {
			return nil, httpapi.ErrWorkspaceLocked
		}
		cache.mu.Lock()
		if cache.entries == nil {
			cache.entries = map[string]cachedCredential{}
		}
		for k, r := range cache.entries {
			if !time.Now().Before(r.until) {
				delete(cache.entries, k)
			}
		}
		cache.entries[key] = cachedCredential{c, time.Now().Add(4 * time.Minute)}
		cache.mu.Unlock()
		return c, nil
	})
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return value.(*httpapi.WorkerClient), nil
}
