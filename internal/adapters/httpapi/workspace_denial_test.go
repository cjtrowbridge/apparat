// Checks denial guidance received through the same worker client used by Files and Editor.
// Covers both current reason codes and older worker replies.
package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func TestWorkspaceDenialGuidance(t *testing.T) {
	for _, tc := range []struct {
		status     int
		body, want string
	}{
		{403, `{"error":"workspace_grant_required","scope":"read"}`, "approved cluster member"},
		{403, `{"error":"owner_confirmation_required"}`, "trust-host"},
		{403, `{"error":"membership_required"}`, "check approval"},
		{401, `{}`, "authenticated session"},
		{403, `{"error":"access_denied"}`, "verify the checkout permissions"},
		{403, `{"error":"workspace_grant_required","scope":"PRIVATE"}`, "verify the checkout permissions"},
	} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			io.WriteString(w, tc.body)
		}))
		c := WorkerClient{http: s.Client(), base: s.URL}
		c.Access.User = "user"
		_, err := c.Workspace(context.Background(), domain.WorkspaceRequest{ProjectID: "remote-test", Action: "list"})
		s.Close()
		if !errors.Is(err, ErrDenied) || !strings.Contains(err.Error(), "Cannot list this folder") || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "PRIVATE") {
			t.Fatalf("%s: %v", tc.body, err)
		}
	}
}
