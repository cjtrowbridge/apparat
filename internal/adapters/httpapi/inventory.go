// Exposes read-only worker observations and administrator-only loopback configuration.
// Uses the existing per-request membership check and device-bound TLS bearer.
package httpapi

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"net/http"
	"time"
)

type InventorySource interface {
	Snapshot() domain.WorkerInventory
	Refresh(context.Context, int) (domain.WorkerInventory, error)
}
type probeReply struct {
	Inventory domain.WorkerInventory
	Error     string
}

func (a *workerAuth) inventoryRoute(w http.ResponseWriter, r *http.Request, admin bool) {
	if r.URL.Path == "/v1/inventory" {
		if r.Method != http.MethodGet {
			respond(w, 405, map[string]string{"error": "method_not_allowed"})
			return
		}
		if a.inventory == nil {
			respond(w, 503, map[string]string{"error": "inventory_unavailable"})
			return
		}
		v := a.inventory.Snapshot()
		if !admin {
			v.Port = 0
			if v.Problem != "" {
				v.Problem = "Ollama metadata unavailable"
			}
		}
		respond(w, 200, v)
		return
	}
	if !admin {
		respond(w, 403, map[string]string{"error": "admin_required"})
		return
	}
	if r.Method != http.MethodPost {
		respond(w, 405, map[string]string{"error": "method_not_allowed"})
		return
	}
	var body struct{ Port int }
	if !boundedJSON(w, r, &body) || body.Port < 1 || body.Port > 65535 {
		respond(w, 400, map[string]string{"error": "invalid_loopback_port"})
		return
	}
	if a.inventory == nil {
		respond(w, 503, map[string]string{"error": "inventory_unavailable"})
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(25 * time.Second))
	v, err := a.inventory.Refresh(r.Context(), body.Port)
	reply := probeReply{Inventory: v}
	if err != nil {
		reply.Error = err.Error()
	}
	respond(w, 200, reply)
}
func (c *WorkerClient) Inventory(ctx context.Context) (domain.WorkerInventory, error) {
	var v domain.WorkerInventory
	err := c.call(ctx, "GET", "/v1/inventory", nil, &v)
	return v, err
}
func (c *WorkerClient) TryOllama(ctx context.Context, port int) (domain.WorkerInventory, error) {
	var reply probeReply
	slow := *c
	slow.http = &http.Client{Transport: c.http.Transport, CheckRedirect: c.http.CheckRedirect, Timeout: 25 * time.Second}
	err := slow.call(ctx, "POST", "/v1/admin/inventory/ollama", map[string]int{"Port": port}, &reply)
	if err == nil && reply.Error != "" {
		err = errors.New(reply.Error)
	}
	return reply.Inventory, err
}
