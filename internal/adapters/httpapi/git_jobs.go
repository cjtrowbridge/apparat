// Runs a bounded set of administrator-initiated host operations behind mutual TLS.
// Progress is scoped to the initiating device. Only the host can choose executables and destinations.
package httpapi

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type gitJob struct {
	owner   persistence.Access
	request domain.GitRequest
	status  domain.GitStatus
	cancel  context.CancelFunc
	until   time.Time
}
type gitJobs struct {
	sync.Mutex
	jobs map[string]*gitJob
}

func (a *workerAuth) gitRoute(w http.ResponseWriter, r *http.Request, access persistence.Access, worker string) {
	if r.Method == "GET" && r.URL.Path == "/v1/admin/git/projects" {
		rows, err := a.workerProjects(r.Context())
		if err != nil {
			respond(w, 503, map[string]string{"error": "unavailable"})
			return
		}
		respond(w, 200, rows)
		return
	}
	if r.Method == "GET" && r.URL.Path == "/v1/admin/git/providers" {
		respond(w, 200, projectgit.Providers(r.Context()))
		return
	}
	var req domain.GitRequest
	if r.URL.Path != "/v1/admin/git/operation" || r.Method != "POST" || !boundedJSON(w, r, &req) || !domain.ValidID(req.ID) {
		respond(w, 400, map[string]string{"error": "invalid_operation"})
		return
	}
	_, known := projectgit.Provider(req.Provider)
	if !known {
		respond(w, 400, map[string]string{"error": "unknown_provider"})
		return
	}
	jobs := &a.git
	jobs.Lock()
	defer jobs.Unlock()
	if jobs.jobs == nil {
		jobs.jobs = map[string]*gitJob{}
	}
	now := time.Now()
	busy := false
	for id, j := range jobs.jobs {
		if j.status.State == "pending" {
			busy = true
		} else if now.After(j.until) {
			delete(jobs.jobs, id)
		}
	}
	j := jobs.jobs[req.ID]
	if j != nil {
		if j.owner.User != access.User || j.owner.Device != access.Device || j.owner.Cluster != access.Cluster {
			respond(w, 403, map[string]string{"error": "access_denied"})
			return
		}
		switch req.Operation {
		case "cancel":
			j.cancel()
		case "status":
		default:
			if req != j.request {
				respond(w, 409, map[string]string{"error": "operation_conflict"})
				return
			}
		}
		respond(w, 200, j.status)
		return
	}
	if req.Operation != "install" && req.Operation != "enroll" && req.Operation != "clone" {
		respond(w, 404, map[string]string{"error": "operation_unavailable"})
		return
	}
	if busy || len(jobs.jobs) >= 128 {
		respond(w, 409, map[string]string{"error": "worker_busy"})
		return
	}
	ctx, cancel := context.WithTimeout(a.gitContext, 16*time.Minute)
	j = &gitJob{owner: access, request: req, status: domain.GitStatus{State: "pending", Message: "Starting on the worker"}, cancel: cancel}
	jobs.jobs[req.ID] = j
	go a.runGit(ctx, j, worker)
	respond(w, 200, j.status)
}

func (a *workerAuth) runGit(ctx context.Context, j *gitJob, worker string) {
	defer j.cancel()
	authorized := func() bool {
		current, err := a.db.CurrentAccess(ctx, j.owner.Cluster, j.owner.User, j.owner.Device)
		return err == nil && current.Admin && a.db.OwnerConfirmed(ctx, worker)
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-t.C:
				if !authorized() {
					j.cancel()
					return
				}
			}
		}
	}()
	progress := func(msg, code string) {
		a.git.Lock()
		j.status.Message = msg
		j.status.Code = code
		a.git.Unlock()
	}
	result := domain.GitStatus{State: "completed"}
	driver, _ := projectgit.Provider(j.request.Provider)
	var err error
	switch {
	case !authorized():
		err = errors.New("administrator access required")
	case j.request.Operation == "install":
		progress("Installing GitHub CLI", "")
		err = projectgit.InstallProvider(ctx, j.request.Provider)
		if err == nil {
			progress("Verifying GitHub CLI", "")
			if !driver.Inspect(ctx).Installed {
				err = errors.New("installed tool unavailable")
			}
		}
	case j.request.Operation == "enroll":
		result.Username, err = driver.Enroll(ctx, authorized, progress)
	case j.request.Operation == "clone":
		result.Project, err = a.cloneOnWorker(ctx, j, authorized, progress)
	}
	if err != nil {
		result.State = "failed"
		result.Message = err.Error()
	}
	if ctx.Err() != nil {
		result = domain.GitStatus{State: "failed", Message: "Operation cancelled or timed out"}
	}
	a.git.Lock()
	j.status = result
	j.until = time.Now().Add(15 * time.Minute)
	a.git.Unlock()
}

func (a *workerAuth) cloneOnWorker(ctx context.Context, j *gitJob, authorized func() bool, progress func(string, string)) (*domain.Project, error) {
	req := j.request
	driver, _ := projectgit.Provider(req.Provider)
	name, err := projectgit.RepoName(req.Repo)
	if err != nil || !projectgit.ValidAccount(req.Account) {
		return nil, errors.New("invalid checkout request")
	}
	if req.Account != "" {
		names, err := driver.Accounts()
		if err != nil || !slices.Contains(names, req.Account) {
			return nil, errors.New("selected account unavailable")
		}
	}
	base := filepath.Dir(filepath.Dir(a.artifacts))
	dest, err := projectgit.Destination(base, "", name)
	if err != nil {
		return nil, err
	}
	receipt := identity.ClaimDigest(strings.Join([]string{j.owner.User, j.owner.Device, req.ID, req.Provider, name, req.Account}, "/"))
	id := "remote-" + receipt
	rows, err := a.db.Projects(ctx)
	if err != nil {
		return nil, errors.New("registry unavailable")
	}
	for _, p := range rows {
		if p.ID == id && p.Location == dest && p.Account == req.Account {
			current, err := projectgit.Inspect(dest)
			current.ID = p.ID
			current.Actions = nil
			current.Entrypoints = nil
			return &current, err
		}
		if p.ID == id || p.Location == dest {
			return nil, errors.New("worker checkout already registered")
		}
	}
	p, err := driver.Clone(ctx, base, "", name, req.Account, receipt, func(msg string) {
		progress(msg, "")
	})
	if err != nil {
		return nil, err
	}
	if !authorized() {
		return nil, errors.New("authorization expired before registration")
	}
	p.ID = id
	err = a.db.RegisterProject(ctx, p)
	if err != nil {
		return nil, errors.New("checkout created but registration failed")
	}
	p.Account = ""
	p.Actions = nil
	p.Entrypoints = nil
	return &p, nil
}

func (a *workerAuth) workerProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := a.db.Projects(ctx)
	if err != nil {
		return nil, err
	}
	out := []domain.Project{}
	for _, p := range rows {
		if !strings.HasPrefix(p.ID, "remote-") {
			continue
		}
		v, err := projectgit.Inspect(p.Location)
		if err != nil {
			v.Name = p.Name
			v.Location = p.Location
		}
		out = append(out, domain.Project{ID: p.ID, Name: v.Name, Location: v.Location, Branch: v.Branch, Revision: v.Revision, Available: v.Available})
	}
	return out, nil
}
