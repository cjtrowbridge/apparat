//go:build android && gui

// Connects system folder selection to durable Android document-tree registrations.
// Clones to a granted folder pass through a private temporary working tree.
package apparatmobile

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"path/filepath"
	"sync"
)

type Documents interface {
	GitHubLogin(string)
	ExportCapture(string) bool
	ChooseFolder() string
	Inspect(string) string
	Access(string, string, string, string) string
	Export(string, string, string, Transfer) string
}

// Transfer allows the document provider copy loop to honor a cancelled clone.
type Transfer interface{ Cancelled() bool }
type transferContext struct{ ctx context.Context }

func (t transferContext) Cancelled() bool {
	return t.ctx.Err() != nil
}

var documents struct {
	sync.RWMutex
	host Documents
}

type projectDocuments struct{}

func RegisterDocuments(host Documents) {
	documents.Lock()
	defer documents.Unlock()
	documents.host = host
}
func currentDocuments() Documents {
	documents.RLock()
	defer documents.RUnlock()
	return documents.host
}
func (projectDocuments) GitHubLogin(code string) {
	host := currentDocuments()
	if host != nil {
		host.GitHubLogin(code)
	}
}
func (projectDocuments) ChooseFolder() string {
	host := currentDocuments()
	if host == nil {
		return ""
	}
	return host.ChooseFolder()
}
func (projectDocuments) InspectTree(uri string) (domain.Project, error) {
	var info struct{ Name, Head, Revision, Tasks, Entrypoints string }
	host := currentDocuments()
	if host == nil || json.Unmarshal([]byte(host.Inspect(uri)), &info) != nil || info.Name == "" {
		return domain.Project{}, errors.New("cannot read Git metadata; reselect the working tree or use app-private storage")
	}
	return domain.Project{Name: info.Name, Branch: info.Head, Entrypoints: projectgit.EnabledTasks(info.Entrypoints), Revision: info.Revision, Actions: projectgit.TaskNames([]byte(info.Tasks)), Location: uri, Available: true}, nil
}
func (d projectDocuments) CloneTree(ctx context.Context, parent, repo string) (domain.Project, error) {
	return d.exportTree(ctx, parent, func(temp string) (domain.Project, error) {
		return projectgit.Clone(ctx, "", temp, repo, "")
	})
}
func (d projectDocuments) CreateTree(ctx context.Context, parent, name string) (domain.Project, error) {
	return d.exportTree(ctx, parent, func(temp string) (domain.Project, error) {
		return projectgit.Create(ctx, "", temp, name)
	})
}
func (d projectDocuments) exportTree(ctx context.Context, parent string, prepare func(string) (domain.Project, error)) (domain.Project, error) {
	host := currentDocuments()
	if host == nil {
		return domain.Project{}, errors.New("folder access unavailable")
	}
	temp, err := os.MkdirTemp(nodeCache, "clone-")
	if err != nil {
		return domain.Project{}, errors.New("temporary checkout storage unavailable")
	}
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	p, err := prepare(temp)
	if err != nil {
		return p, err
	}
	err = ctx.Err()
	if err != nil {
		return p, err
	}
	uri := host.Export(parent, filepath.Base(p.Location), p.Location, transferContext{ctx})
	if uri == "" {
		return domain.Project{}, errors.New("folder cannot preserve this checkout or destination is occupied; try app-private storage")
	}
	return d.InspectTree(uri)
}

var nodeCache string

func (projectDocuments) ExportCapture(path string) error {
	h := currentDocuments()
	if h == nil || !h.ExportCapture(path) {
		return errors.New("capture export cancelled or unavailable")
	}
	return nil
}
