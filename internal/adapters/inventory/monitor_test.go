// Tests installed and loaded inventory, failed reconfiguration and expiry.
// Only metadata URLs are permitted by the loopback fixture.
package inventory

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestOllamaInventory(t *testing.T) {
	var shows atomic.Int32
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			fmt.Fprint(w, `{"version":"0.14.1"}`)
		case "/api/tags":
			fmt.Fprint(w, `{"models":[{"name":"test-model","size":100,"digest":"123"},{"name":"embed","size":45,"digest":"456"}]}`)
		case "/api/ps":
			fmt.Fprint(w, `{"models":[{"name":"test-model"}]}`)
		case "/api/show":
			shows.Add(1)
			if r.Method != "POST" {
				t.Error("unexpected method")
			}
			var b [128]byte
			n, _ := r.Body.Read(b[:])
			if strings.Contains(string(b[:n]), "embed") {
				fmt.Fprint(w, `{"capabilities":["embedding"]}`)
			} else {
				fmt.Fprint(w, `{"capabilities":["completion"]}`)
			}
		default:
			t.Error("unexpected request", r.URL.Path)
		}
	}))
	defer fixture.Close()
	port, err := strconv.Atoi(strings.Split(fixture.URL, ":")[2])
	if err != nil {
		t.Fatal(err)
	}
	saved := 0
	m := New(0, func(_ context.Context, value int) error {
		saved = value
		return nil
	})
	v, err := m.Refresh(t.Context(), port)
	if err != nil || saved != port || !v.HasText(time.Now()) || v.HasText(time.Now().Add(3*time.Minute)) || len(v.Models) != 2 || v.Models[0].Text || !v.Models[1].Loaded || shows.Load() != 2 {
		t.Fatal(v, err)
	}
	v, err = m.Refresh(t.Context(), 0)
	if err != nil || shows.Load() != 2 || v.Resources.CPUs < 1 {
		t.Fatal(v, err)
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fixture.URL+r.URL.Path, http.StatusFound)
	}))
	defer bad.Close()
	badPort, _ := strconv.Atoi(strings.Split(bad.URL, ":")[2])
	v, err = m.Refresh(t.Context(), badPort)
	if err == nil || !strings.Contains(err.Error(), "HTTP 302") || v.Port != port || saved != port || !v.HasText(time.Now()) {
		t.Fatal("lost existing configuration", v, err)
	}
	fixture.Close()
	v, err = m.Refresh(t.Context(), 0)
	if err == nil || v.HasText(time.Now()) || len(v.Models) != 0 || v.Problem == "" {
		t.Fatal("stale inventory after outage", v, err)
	}
}
