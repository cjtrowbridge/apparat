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
	projects, err := n.ProjectPage(ctx, "local", "")
	if err != nil {
		return nil, err
	}
	return append(projects, n.remoteProjects(ctx)...), nil
}
func (n *Runtime) ProjectPage(ctx context.Context, host, before string) ([]domain.Project, error) {
	if host != "local" {
		return n.remoteProjectPage(ctx, host, before)
	}
	projects, err := n.database.RegistryPage(ctx, before, false, "", true)
	if err != nil {
		return nil, err
	}
	for i, p := range projects {
		projects[i].Threads, err = n.database.ChatThreads(ctx, p.ID)
		if err != nil {
			return nil, err
		}
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
	return n.SetupLocalProject(ctx, domain.LocalProjectSetup{Mode: "empty", Parent: parent, Name: name}, nil)
}
func (n *Runtime) CloneProject(ctx context.Context, parent, repo, account string) error {
	_, err := n.SetupLocalProject(ctx, domain.LocalProjectSetup{Mode: "clone", Parent: parent, Name: repo, Account: account}, nil)
	return err
}
func (n *Runtime) SetupLocalProject(ctx context.Context, req domain.LocalProjectSetup, progress func(string)) (string, error) {
	report := func(message string) {
		if progress != nil {
			progress(message)
		}
	}
	var p domain.Project
	var err error
	tree := strings.HasPrefix(req.Parent, "content://")
	if tree && n.documents == nil {
		return "", errors.New("folder access unavailable")
	}
	switch req.Mode {
	case "empty":
		report("Creating the project folder and initializing Git on main…")
		if tree {
			p, err = n.documents.CreateTree(ctx, req.Parent, req.Name)
		} else {
			p, err = projectgit.Create(ctx, n.config.RootDir, req.Parent, req.Name)
		}
	case "clone":
		report("Preparing the GitHub checkout destination…")
		if tree {
			if req.Account != "" {
				return "", errors.New("GitHub login is unavailable for this folder provider")
			}
			report("Downloading the repository and exporting its working tree to the selected folder…")
			p, err = n.documents.CloneTree(ctx, req.Parent, req.Name)
		} else {
			p, err = projectgit.CloneWithProgress(ctx, n.config.RootDir, req.Parent, req.Name, req.Account, progress)
		}
	case "import":
		report("Inspecting the selected Git working tree…")
		p, err = n.inspectProject(req.Name)
		p.Account = req.Account
	default:
		return "", errors.New("unknown local repository source")
	}
	if err != nil {
		return "", err
	}
	if !projectgit.ValidAccount(p.Account) {
		return "", errors.New("invalid GitHub account")
	}
	report("Saving the checkout registration in the local project database…")
	err = n.registerProject(ctx, p)
	if err != nil {
		return "", errors.New("could not save the registration; the checkout can be added using Existing folder")
	}
	return p.Location, nil
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
