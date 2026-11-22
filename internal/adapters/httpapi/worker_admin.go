// Accepts ordered policy proposals and the explicit first-owner handoff to a worker.
// Both transport and database authorization must still hold when a change commits.
package httpapi

import (
	"net/http"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (a *workerAuth) admin(w http.ResponseWriter, r *http.Request, access persistence.Access, keys *identity.Session) {
	if !access.Admin || r.Method != http.MethodPost {
		respond(w, 403, map[string]string{"error": "admin_required"})
		return
	}
	var event persistence.PolicyEvent
	switch r.URL.Path {
	case "/v1/admin/propose":
		var proposal identity.Proposal
		if !boundedJSON(w, r, &proposal) || !sameActor(access, proposal) {
			respond(w, 403, map[string]string{"error": "invalid_actor"})
			return
		}
		ordered, err := a.db.CommitPolicy(r.Context(), proposal, keys)
		if err != nil {
			respond(w, 409, map[string]string{"error": "policy_conflict"})
			return
		}
		event = ordered
	case "/v1/admin/order-on-worker":
		if !boundedJSON(w, r, &event) || !sameActor(access, event.Proposal) || event.Proposal.Edit.Kind != "order_on_worker" || event.Proposal.Edit.Target != identity.Fingerprint(keys.Public()) {
			respond(w, 403, map[string]string{"error": "invalid_handoff"})
			return
		}
		err := a.db.ApplyPolicy(r.Context(), event, event.Proposal.Device.Binding.Device)
		if err != nil {
			respond(w, 409, map[string]string{"error": "handoff_rejected"})
			return
		}
	}
	respond(w, 200, event)
}

func sameActor(a persistence.Access, p identity.Proposal) bool {
	return p.Edit.Cluster == a.Cluster && identity.Fingerprint(p.Device.Binding.User) == a.User && identity.Fingerprint(p.Device.Binding.Device) == a.Device
}
