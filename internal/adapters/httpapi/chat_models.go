// Exchanges scoped model choices over the worker's authenticated, device-bound API.
// Only current cluster administrators may change or replicate a scope default.
package httpapi

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"net/http"
)

type ModelCommand struct {
	Action, Project, Session string
	Edit                     domain.ModelEdit
	Replica                  domain.ModelDefault
	Fallback                 domain.ModelRef
}

func (c *WorkerClient) ModelCommand(ctx context.Context, q ModelCommand) (s domain.ModelSettings, err error) {
	err = c.call(ctx, "POST", "/v1/chat/models", q, &s)
	if errors.Is(err, ErrWorkspaceConflict) {
		err = domain.ErrModelConflict
	}
	return
}
func (a *workerAuth) chatModels(w http.ResponseWriter, r *http.Request, access persistence.Access, worker string) {
	var q ModelCommand
	if r.Method != "POST" || !boundedJSON(w, r, &q) || len(q.Session) > 256 {
		respond(w, 400, "invalid model request")
		return
	}
	cluster := access.Cluster
	if q.Project != "" {
		if !domain.ValidID(q.Project) || !a.db.OwnerConfirmed(r.Context(), worker) {
			respond(w, 403, "project unavailable")
			return
		}
		_, err := a.db.WorkspaceLocation(r.Context(), q.Project, access.User, "read")
		if err != nil {
			respond(w, 403, "project unavailable")
			return
		}
		cluster = ""
	}
	var s domain.ModelSettings
	var err error
	switch q.Action {
	case "read":
		s, err = a.db.ChatModels(r.Context(), cluster, q.Project, q.Session)
	case "bind":
		if q.Project == "" {
			respond(w, 400, "expected project")
			return
		}
		s, err = a.db.BindChatModel(r.Context(), cluster, q.Project, q.Session, q.Fallback)
	case "save":
		if (q.Edit.Default || q.Project == "") && !access.Admin {
			respond(w, 403, "administrator required")
			return
		}
		if q.Project == "" && !q.Edit.Default {
			respond(w, 400, "expected global default")
			return
		}
		if q.Project == "" {
			d := domain.ModelDefault{Model: q.Edit.Model, Revision: q.Edit.Expected.Revision + 1, Operation: q.Edit.Operation, Proof: q.Edit.Proof}
			u, dev, err := identity.ModelDefaultAuthor(cluster, d)
			if err != nil || u != access.User || dev != access.Device {
				respond(w, 403, "invalid administrator proof")
				return
			}
		}
		q.Edit.Project = q.Project
		q.Edit.Session = q.Session
		s, err = a.db.SaveChatModel(r.Context(), cluster, q.Edit)
	case "merge":
		u, dev, verifyErr := identity.ModelDefaultAuthor(cluster, q.Replica)
		if q.Project != "" || verifyErr != nil || !a.db.ModelAdmin(r.Context(), cluster, u, dev) {
			respond(w, 403, "administrator required")
			return
		}
		err = a.db.MergeGlobalModel(r.Context(), cluster, q.Replica)
	default:
		respond(w, 400, "unknown model operation")
		return
	}
	if errors.Is(err, domain.ErrModelConflict) {
		respond(w, 412, "default changed")
		return
	}
	if err != nil {
		respond(w, 409, "model settings changed or unavailable")
		return
	}
	if (q.Action == "save" && q.Edit.Default) || q.Action == "merge" {
		a.modelPulse()
	}
	s.CanDefault = access.Admin
	respond(w, 200, s)
}
