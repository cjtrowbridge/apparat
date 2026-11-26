//go:build gui

// Presents host-scoped credential installation, device authorization and worker clone progress.
// A pending operation owns its inputs until completion or explicit cancellation.
package gui

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type projectProviders interface {
	GitProviders(context.Context, string) ([]domain.GitProvider, error)
	GitSession(context.Context, string) (func(context.Context, domain.GitRequest) (domain.GitStatus, error), error)
}

func (g *Game) projectHostName(host string) string {
	for _, h := range g.projectHosts() {
		if h.id == host {
			return h.label
		}
	}
	return "selected device"
}

func (g *Game) refreshProviders(host string) {
	w := &g.projects.wizard
	source, ok := g.runtimeInfo.Queries.(projectProviders)
	if !ok {
		return
	}
	w.providers = nil
	w.providerError = ""
	w.enrollmentHelp = ""
	g.projectWork(func(ctx context.Context) projectResult {
		providers, err := source.GitProviders(ctx, host)
		return projectResult{applyField: func() {
			w.accountHost = host
			w.providers = providers
			if err != nil {
				w.providerError = "Provider status unavailable. Check the connection and admin access, then retry."
			}
		}}
	})
}

func (g *Game) startHostOperation(host string, req domain.GitRequest, value *string) {
	f := &g.projects
	w := &f.wizard
	source, ok := g.runtimeInfo.Queries.(projectProviders)
	if !ok || f.pending != nil {
		return
	}
	previous := w.retryRequest
	previous.ID = ""
	if previous == req {
		req.ID = w.retryRequest.ID
	}
	if req.ID == "" {
		id, err := domain.NewID(time.Now(), nil)
		if err != nil {
			return
		}
		req.ID = id
	}
	w.retryRequest = req
	queue := make(chan domain.GitStatus, 1)
	w.jobUpdates = queue
	w.jobStatus = domain.GitStatus{Message: "Connecting to " + g.projectHostName(host)}
	g.projectWork(func(ctx context.Context) projectResult {
		out, err := waitHostOperation(ctx, source, host, req, queue)
		if err != nil {
			return projectResult{message: err.Error()}
		}
		if out.State != "completed" {
			return projectResult{message: out.Message, applyField: func() {
				w.retryRequest = domain.GitRequest{}
			}}
		}
		if req.Operation == "clone" && out.Project != nil {
			rows, err := f.store.Projects(ctx)
			if err != nil {
				return projectResult{message: "Clone registered. Reopen Projects to refresh."}
			}
			id := host + "/" + out.Project.ID
			return projectResult{applyField: func() {
				f.rows = rows
				f.wizard = newProjectState{}
				for i, p := range rows {
					if p.ID == id {
						g.selectSection(hud.TabProjects, len(projectEntrySections)+i)
						return
					}
				}
				f.message = "Clone registered. Reopen Projects to refresh."
			}}
		}
		providers, err := source.GitProviders(ctx, host)
		if err != nil {
			return projectResult{message: "Operation completed. Refresh the account list."}
		}
		return projectResult{applyField: func() {
			w.providers = providers
			w.accountHost = host
			w.retryRequest = domain.GitRequest{}
			if out.Username != "" {
				*value = out.Username
				w.preview = domain.LocalProjectSetup{}
			}
		}}
	})
}

func waitHostOperation(ctx context.Context, source projectProviders, host string, req domain.GitRequest, queue chan domain.GitStatus) (domain.GitStatus, error) {
	call, err := source.GitSession(ctx, host)
	if err != nil {
		return domain.GitStatus{}, errors.New("Cannot connect to the selected host.")
	}
	status, err := call(ctx, req)
	poll := domain.GitRequest{ID: req.ID, Provider: req.Provider, Operation: "status"}
	for err == nil && status.State == "pending" {
		select {
		case queue <- status:
		default:
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(2 * time.Second):
			status, err = call(ctx, poll)
		}
	}
	if ctx.Err() != nil && host != "local" {
		stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		poll.Operation = "cancel"
		_, _ = call(stop, poll)
		return domain.GitStatus{}, errors.New("Cancellation requested")
	}
	if err != nil {
		return domain.GitStatus{}, errors.New("Host operation unavailable. Check administrator access and connectivity, then retry.")
	}
	return status, nil
}

func (g *Game) hostJobControls(c *widget.Container) {
	w := &g.projects.wizard
	c.AddChild(g.detailText(w.jobStatus.Message))
	code := w.jobStatus.Code
	if code != "" {
		c.AddChild(g.detailText("Enter " + code + " at https://github.com/login/device"))
		g.projectButton(c, "Copy code and open GitHub", func() {
			opener, ok := g.runtimeInfo.ProjectPicker.(interface{ GitHubLogin(string) })
			if ok {
				opener.GitHubLogin(code)
				return
			}
			url := "https://github.com/login/device"
			program := "xdg-open"
			args := []string{url}
			if runtime.GOOS == "darwin" {
				program = "open"
			} else if runtime.GOOS == "windows" {
				program = "rundll32.exe"
				args = []string{"url.dll,FileProtocolHandler", url}
			}
			go func() {
				var clipboard *exec.Cmd
				switch runtime.GOOS {
				case "windows":
					clipboard = exec.Command("clip.exe")
				case "darwin":
					clipboard = exec.Command("pbcopy")
				default:
					clipboard = exec.Command("wl-copy", "--trim-newline")
				}
				clipboard.Stdin = strings.NewReader(code)
				_ = clipboard.Run()
				_ = exec.Command(program, args...).Run()
			}()
		})
	}
	g.projectButton(c, "Cancel", g.projects.cancel)
}

func matchesProjectHost(host setupChoice, query string) bool {
	return strings.Contains(strings.ToLower(host.label), strings.ToLower(strings.TrimSpace(query)))
}

func searchLine(value string) string {
	return searchBreaks.Replace(value)
}
func compactSearchControl(control *widget.Widget, font int) {
	height := font + 12
	control.MinHeight = height
	control.LayoutData = widget.GridLayoutData{MaxHeight: height, VerticalPosition: widget.GridLayoutPositionCenter}
}

func (g *Game) remoteCloneOption() *accountOption {
	f := &g.projects
	w := &f.wizard
	if w.mode != "clone" || !strings.HasPrefix(w.host, "worker-") {
		return nil
	}
	return &accountOption{"clone", "Clone on " + g.projectHostName(w.host), func() {
		_, err := projectgit.RepoName(f.repo)
		if err != nil {
			f.message = "Enter a GitHub owner/repository or repository URL."
			g.layoutDirty = true
			return
		}
		g.startHostOperation(w.host, domain.GitRequest{Provider: "github", Operation: "clone", Repo: f.repo, Account: w.remoteAccount}, &w.remoteAccount)
	}}
}
