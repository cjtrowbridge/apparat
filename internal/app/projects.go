// Coordinates owner-local registration and Git observations through the node.
// Android document-tree checkouts use the persisted picker grant.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
	"time"
)

type ProjectDocuments interface {
	InspectTree(string) (domain.Project, error)
	CreateTree(context.Context, string, string) (domain.Project, error)
	CloneTree(context.Context, string, string) (domain.Project, error)
}

func (n *Runtime) inspectProject(path string) (domain.Project, error) {
	if strings.HasPrefix(path, "content://") {
		if n.documents == nil {
			return domain.Project{}, errors.New("checkout permission unavailable")
		}
		return n.documents.InspectTree(path)
	}
	return projectgit.Inspect(path)
}
func (n *Runtime) Projects(ctx context.Context) ([]domain.Project, error) {
	projects, err := n.database.Projects(ctx)
	if err != nil {
		return nil, err
	}
	for i, p := range projects {
		current, err := n.inspectProject(p.Location)
		if err == nil {
			projects[i].Actions = current.Actions
			projects[i].Entrypoints = current.Entrypoints
			projects[i].Branch = current.Branch
			projects[i].Revision = current.Revision
			projects[i].Available = true
		}
	}
	return projects, nil
}
func (n *Runtime) StoreProject(ctx context.Context, id, location, account string) error {
	n.workspaceMu.Lock()
	defer n.workspaceMu.Unlock()
	if !projectgit.ValidAccount(account) {
		return errors.New("invalid GitHub account")
	}
	p, err := n.inspectProject(location)
	if err != nil {
		return err
	}
	p.Account = account
	if id != "" {
		return n.database.UpdateProject(ctx, id, p.Location, p.Name, p.Account)
	}
	return n.registerProject(ctx, p)
}
func (n *Runtime) registerProject(ctx context.Context, p domain.Project) error {
	id, err := domain.NewID(time.Now(), nil)
	if err != nil {
		return err
	}
	p.ID = id
	return n.database.RegisterProject(ctx, p)
}
func (n *Runtime) CreateProject(ctx context.Context, parent, name string) (string, error) {
	var p domain.Project
	var err error
	if strings.HasPrefix(parent, "content://") {
		if n.documents == nil {
			return "", errors.New("folder access unavailable")
		}
		p, err = n.documents.CreateTree(ctx, parent, name)
	} else {
		p, err = projectgit.Create(ctx, n.config.RootDir, parent, name)
	}
	if err != nil {
		return "", err
	}
	err = n.registerProject(ctx, p)
	if err != nil {
		return "", errors.New("repository created; choose Import local repo to finish registration")
	}
	return p.Location, nil
}
func (n *Runtime) CloneProject(ctx context.Context, parent, repo, account string) error {
	var p domain.Project
	var err error
	if strings.HasPrefix(parent, "content://") && n.documents != nil {
		p, err = n.documents.CloneTree(ctx, parent, repo)
	} else {
		p, err = projectgit.Clone(ctx, n.config.RootDir, parent, repo, account)
	}
	if err != nil {
		return err
	}
	return n.registerProject(ctx, p)
}
func (n *Runtime) RemoveProject(ctx context.Context, id string) error {
	n.workspaceMu.Lock()
	defer n.workspaceMu.Unlock()
	return n.database.RemoveProject(ctx, id)
}
func (n *Runtime) ProjectDefault(ctx context.Context) (string, error) {
	return n.database.ProjectDefault(ctx)
}
func (n *Runtime) SetProjectDefault(ctx context.Context, name string) error {
	if !projectgit.ValidAccount(name) {
		return errors.New("invalid GitHub account")
	}
	return n.database.SetProjectDefault(ctx, name)
}
func (n *Runtime) SetProjectDocuments(d ProjectDocuments) {
	n.documents = d
}
