// Verifies the member/admin disclosure boundary and bounded port trial contract.
// Authentication and device-binding cases run through the worker route tests.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"net/http/httptest"
	"strings"
	"testing"
)

type inventoryFixture struct{ attempts int }

func (f *inventoryFixture) Snapshot() domain.WorkerInventory {
	return domain.WorkerInventory{Port: 12345, Problem: "sensitive local port 12345"}
}
func (f *inventoryFixture) Refresh(context.Context, int) (domain.WorkerInventory, error) {
	f.attempts++
	return f.Snapshot(), errors.New("test metadata failure")
}
func TestInventoryAuthorization(t *testing.T) {
	fixture := &inventoryFixture{}
	auth := workerAuth{inventory: fixture}
	for _, tc := range []struct {
		path, method, body string
		admin              bool
		want               int
	}{
		{"/v1/inventory", "GET", "", false, 200},
		{"/v1/admin/inventory/ollama", "POST", `{"Port":11434}`, false, 403},
		{"/v1/admin/inventory/ollama", "POST", `{"Port":65536}`, true, 400},
		{"/v1/admin/inventory/ollama", "POST", `{"Port":11434,"URL":"elsewhere"}`, true, 400},
		{"/v1/admin/inventory/ollama", "POST", `{"Port":11434}`, true, 200},
	} {
		r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		auth.inventoryRoute(w, r, tc.admin, "worker")
		if w.Code != tc.want {
			t.Fatal(tc, w.Code)
		}
		if !tc.admin && strings.Contains(w.Body.String(), "12345") {
			t.Fatal("private port returned")
		}
		if tc.admin && tc.want == 200 {
			var reply probeReply
			err := json.Unmarshal(w.Body.Bytes(), &reply)
			if err != nil || reply.Error != "test metadata failure" {
				t.Fatal(reply, err)
			}
		}
	}
	if fixture.attempts != 1 {
		t.Fatal("unexpected probe count", fixture.attempts)
	}
}
