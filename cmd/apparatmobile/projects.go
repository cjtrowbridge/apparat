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
	host := currentDocuments()
	if host == nil {
		return domain.Project{}, errors.New("folder access unavailable")
	}
	temp, err := os.MkdirTemp(nodeCache, "clone-")
	if err != nil {
		return domain.Project{}, errors.New("clone storage unavailable")
	}
	defer func() {
		_ = os.RemoveAll(temp)
	}()
	p, err := projectgit.Clone(ctx, "", temp, repo, "")
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
