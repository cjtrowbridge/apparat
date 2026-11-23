// Adapts the scoped owner workspace command to a bounded authenticated JSON route.
// Only a confirmed worker can dispatch; the runtime independently checks project grants.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"io"
	"net/http"
)

type OwnerWorkspace func(context.Context, persistence.Access, domain.WorkspaceRequest) (domain.WorkspaceView, error)

var ErrWorkspaceConflict = errors.New("checkout changed; compare the saved draft with the current file")

func (a *workerAuth) ownerWorkspace(w http.ResponseWriter, r *http.Request, access persistence.Access, worker string) {
	if r.Method != "POST" {
		respond(w, 405, map[string]string{"error": "method_not_allowed"})
		return
	}
	if a.workspace == nil || !a.db.OwnerConfirmed(r.Context(), worker) {
		respond(w, 403, map[string]string{"error": "access_denied"})
		return
	}
	var q domain.WorkspaceRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512<<10))
	decoder.DisallowUnknownFields()
	if r.Header.Get("Content-Type") != "application/json" || decoder.Decode(&q) != nil || decoder.Decode(new(any)) != io.EOF {
		respond(w, 400, map[string]string{"error": "invalid_request"})
		return
	}
	v, err := a.workspace(r.Context(), access, q)
	switch {
	case errors.Is(err, ErrDenied):
		respond(w, 403, map[string]string{"error": "access_denied"})
	case errors.Is(err, ErrWorkspaceConflict):
		respond(w, 412, map[string]string{"error": "base_changed"})
	case err != nil:
		respond(w, 409, map[string]string{"error": "workspace_operation_failed"})
	default:
		respond(w, 200, v)
	}
}
func (c *WorkerClient) Workspace(ctx context.Context, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	var v domain.WorkspaceView
	err := c.call(ctx, "POST", "/v1/workspace", q, &v)
	return v, err
}
