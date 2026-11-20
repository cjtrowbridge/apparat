// Reads local Git checkout metadata and clones explicit GitHub repositories.
// Repository task definitions are observed as data and never executed.
package projectgit

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var repositoryName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*/[a-zA-Z0-9_.-]+$`)
var accountName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)

func ValidAccount(name string) bool {
	return name == "" || accountName.MatchString(name)
}

func RepoName(input string) (string, error) {
	name := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "https://github.com/"), "/"), ".git")
	if !repositoryName.MatchString(name) || filepath.Base(name) == ".." || filepath.Base(name) == "." {
		return "", errors.New("enter a GitHub owner/repo or HTTPS repository URL")
	}
	return name, nil
}

func Inspect(path string) (domain.Project, error) {
	var p domain.Project
	if !filepath.IsAbs(path) {
		return p, errors.New("choose an absolute checkout path")
	}
	root, err := filepath.EvalSymlinks(path)
	if err != nil {
		return p, errors.New("checkout unavailable; choose its current folder")
	}
	metadata, err := os.Stat(filepath.Join(root, ".git"))
	if err != nil || !metadata.IsDir() && (!metadata.Mode().IsRegular() || metadata.Size() > 4096) {
		return p, errors.New("choose a Git working tree")
	}
	repo, err := git.PlainOpenWithOptions(root, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return p, errors.New("choose the root of an existing Git working tree")
	}
	_, err = repo.Worktree()
	if err != nil {
		return p, errors.New("bare repositories have no working tree")
	}
	p.Name = filepath.Base(root)
	p.Location = root
	p.Available = true
	ref, err := repo.Reference(plumbing.HEAD, false)
	if err == nil && ref.Target().IsBranch() {
		p.Branch = ref.Target().Short()
	}
	head, err := repo.Head()
	if err == nil {
		p.Revision = head.Hash().String()[:12]
	}
	sandbox, err := os.OpenRoot(root)
	if err != nil {
		return p, err
	}
	defer func() {
		_ = sandbox.Close()
	}()
	definition, _ := ReadText(sandbox, ".apparat/tasks.json")
	p.Entrypoints = EnabledTasks(definition)
	info, err := sandbox.Stat(".vscode/tasks.json")
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return p, nil
	}
	f, err := sandbox.Open(".vscode/tasks.json")
	if err == nil {
		defer func() {
			_ = f.Close()
		}()
		info, err := f.Stat()
		if err == nil && info.Mode().IsRegular() && info.Size() <= 1<<20 {
			data, err := io.ReadAll(io.LimitReader(f, 1<<20))
			if err == nil {
				p.Actions = TaskNames(data)
			}
		}
	}
	return p, nil
}

func TaskNames(data []byte) []string {
	var spec struct{ Tasks []struct{ Label string } }
	if json.Unmarshal(data, &spec) != nil {
		return nil
	}
	var names []string
	for _, task := range spec.Tasks {
		name := strings.TrimSpace(task.Label)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func Destination(base, parent, name string) (string, error) {
	repo, err := RepoName(name)
	if err != nil {
		return "", err
	}
	if parent == "" {
		return filepath.Join(base, "data", "repos", filepath.FromSlash(repo)), nil
	}
	if !filepath.IsAbs(parent) {
		return "", errors.New("choose an absolute parent folder")
	}
	return filepath.Join(parent, filepath.Base(repo)), nil
}

func Clone(ctx context.Context, base, parent, input, account string) (domain.Project, error) {
	var p domain.Project
	name, err := RepoName(input)
	if err != nil {
		return p, err
	}
	dest, err := Destination(base, parent, name)
	if err != nil {
		return p, err
	}
	opts := &git.CloneOptions{URL: "https://github.com/" + name + ".git", Depth: 1}
	if account != "" {
		if runtime.GOOS == "android" || !ValidAccount(account) {
			return p, errors.New("account unavailable on this device")
		}
		opts.Auth, err = accountAuth(ctx, account)
		if err != nil {
			return p, err
		}
	}
	if parent == "" {
		err = os.MkdirAll(filepath.Dir(dest), 0700)
		if err != nil {
			return p, errors.New("cannot create default clone directory")
		}
	}
	_, err = os.Lstat(dest)
	if !errors.Is(err, os.ErrNotExist) {
		return p, errors.New("destination unavailable or occupied; use Import local repo for a checkout")
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".apparat-clone-")
	if err != nil {
		return p, errors.New("cannot write to selected clone parent")
	}
	defer func() {
		_ = os.RemoveAll(stage)
	}()
	_, err = git.PlainCloneContext(ctx, stage, false, opts)
	if err != nil {
		return p, errors.New("clone incomplete; check the connection and the selected account’s repository access")
	}
	err = installClone(stage, dest)
	if err != nil {
		return p, errors.New("cannot publish clone; verify the destination is free and supports atomic rename")
	}
	p, err = Inspect(dest)
	p.Account = account
	return p, err
}
func withoutVariable(env []string, key string) []string {
	filtered := make([]string, 0, len(env))
	for _, s := range env {
		if !strings.HasPrefix(s, key+"=") {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func accountAuth(ctx context.Context, account string) (*http.BasicAuth, error) {
	cmd := exec.CommandContext(ctx, "gh", "auth", "token", "--hostname", "github.com", "--user", account)
	cmd.Env = os.Environ()
	for _, key := range []string{"GH_TOKEN", "GITHUB_TOKEN"} {
		cmd.Env = withoutVariable(cmd.Env, key)
	}
	token, err := cmd.Output()
	if err != nil {
		return nil, errors.New("check the selected account with system gh auth login")
	}
	return &http.BasicAuth{Username: "x-access-token", Password: strings.TrimSpace(string(token))}, nil
}
