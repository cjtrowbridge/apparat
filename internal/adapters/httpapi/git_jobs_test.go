// Exercises account-scoped GitHub clones against a private local Git fixture.
// The production worker handler, checkout path, and SQLite registration participate in every request.
package httpapi

import (
	"bytes"
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"io"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixtureTransport struct{ server *httptest.Server }

func (f fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Host != "github.com" {
		return nil, errors.New("unexpected Git fixture host")
	}
	request := r.Clone(r.Context())
	request.URL.Scheme = "http"
	request.URL.Host = f.server.Listener.Addr().String()
	res, err := http.DefaultTransport.RoundTrip(request)
	if res != nil {
		res.Request = r
	}
	return res, err
}

func exerciseWorkerGit(t *testing.T, admin, other, pending *WorkerClient) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	repo := filepath.Join(root, "org", "secret.git")
	err := os.MkdirAll(repo, 0700)
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-b", "main", repo}, {"-C", repo, "-c", "user.name=fixture", "-c", "user.email=f@test.local", "commit", "--allow-empty", "-m", "fixture"}} {
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatal(string(out), err)
		}
	}
	gitPath, err := exec.Command("git", "--exec-path").Output()
	if err != nil {
		t.Fatal(err)
	}
	handler := cgi.Handler{Path: filepath.Join(strings.TrimSpace(string(gitPath)), "git-http-backend"), Env: []string{"GIT_PROJECT_ROOT=" + root, "GIT_HTTP_EXPORT_ALL=1"}}
	fixture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pw, ok := r.BasicAuth()
		if !ok || pw != "fixture-only" {
			w.WriteHeader(401)
			return
		}
		if r.Method == "POST" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(400)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			r.TransferEncoding = nil
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		for k, vs := range rec.Header() {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(rec.Body.Bytes())

	}))
	defer fixture.Close()
	original := client.Protocols["https"]
	client.InstallProtocol("https", githttp.NewClient(&http.Client{Transport: fixtureTransport{fixture}}))
	defer client.InstallProtocol("https", original)
	bin := t.TempDir()
	err = os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\ncase \"$*\" in\n'auth token --hostname github.com --user alice') echo fixture-only;;\n'auth token --help') echo --user;;\n*) exit 1;;\nesac\n"), 0700)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, op := range []string{"install", "enroll", "clone", "status", "cancel"} {
		_, err := pending.GitOperation(ctx, domain.GitRequest{Operation: op, ID: "01HF7YAT000000000000000000", Provider: "github"})
		if !errors.Is(err, ErrDenied) {
			t.Fatal("denied session accepted", op, err)
		}
	}
	for _, account := range []string{"", "bob", "alice"} {
		id, err := domain.NewID(time.Now(), nil)
		if err != nil {
			t.Fatal(err)
		}
		req := domain.GitRequest{Provider: "github", Operation: "clone", Repo: "org/secret", Account: account, ID: id}
		status, err := admin.GitOperation(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		poll := domain.GitRequest{Provider: "github", Operation: "status", ID: id}
		_, err = other.GitOperation(ctx, poll)
		if !errors.Is(err, ErrDenied) {
			t.Fatal("different device read operation")
		}
		for i := 0; i < 300 && status.State == "pending"; i++ {
			time.Sleep(10 * time.Millisecond)
			status, err = admin.GitOperation(ctx, poll)
			if err != nil {
				t.Fatal(err)
			}
		}
		if account != "alice" {
			if status.State != "failed" {
				t.Fatal("private clone permitted", account, status)
			}
			continue
		}
		if status.State != "completed" || status.Project == nil || status.Project.Revision == "" {
			t.Fatal("private clone failed", status)
		}
		again, err := admin.GitOperation(ctx, req)
		if err != nil || again.Project == nil || again.Project.ID != status.Project.ID {
			t.Fatal("retry lost registration")
		}
		rows, err := admin.RemoteProjects(ctx)
		if err != nil || len(rows) != 1 || rows[0].Branch != "main" || rows[0].Account != "" {
			t.Fatal("incorrect remote catalog", rows, err)
		}
	}
}
