// Detects named checkout providers in the executing worker's service environment.
// Installation uses fixed package IDs and arguments; caller input never selects an executable.
package projectgit

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func termux() bool {
	return os.Getenv("PREFIX") == "/data/data/com.termux/files/usr" && (runtime.GOOS == "android" || runtime.GOOS == "linux")
}

func installCommand(ctx context.Context) ([]string, string) {
	if termux() {
		p := filepath.Join(os.Getenv("PREFIX"), "bin", "pkg")
		_, err := exec.LookPath(p)
		if err == nil {
			return []string{p, "install", "-y", "gh"}, ""
		}
		return nil, "The Termux package manager is unavailable in this worker."
	}
	switch runtime.GOOS {
	case "linux":
		for _, args := range [][]string{{"apt-get", "install", "-y", "gh"}, {"dnf", "install", "-y", "gh"}, {"zypper", "--non-interactive", "install", "gh"}} {
			path, err := exec.LookPath(args[0])
			if err != nil {
				continue
			}
			args[0] = path
			if os.Geteuid() == 0 {
				return args, ""
			}
			sudo, err := exec.LookPath("sudo")
			if err == nil {
				check := exec.CommandContext(ctx, sudo, append([]string{"-n", "-l", "--"}, args...)...)
				if check.Run() == nil {
					return append([]string{sudo, "-n", "--"}, args...), ""
				}
			}
			return nil, "Install the gh package locally or authorize this service account to install that package."
		}
	case "windows":
		path, err := exec.LookPath("winget.exe")
		if err == nil {
			return []string{path, "install", "--exact", "--id", "GitHub.cli", "--source", "winget", "--silent", "--disable-interactivity", "--accept-package-agreements", "--accept-source-agreements"}, ""
		}
		return nil, "WinGet is unavailable to the Apparat service account. Install GitHub CLI locally."
	case "darwin":
		path, err := exec.LookPath("brew")
		if err == nil {
			return []string{path, "install", "gh"}, ""
		}
		return nil, "Install GitHub CLI locally or make Homebrew available to this account."
	case "android", "ios":
		return nil, "This app cannot install system command-line tools. Select a desktop or Termux worker."
	}
	return nil, "No supported package manager is available to this worker. Install GitHub CLI locally."
}

func githubProvider(ctx context.Context) domain.GitProvider {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	p := domain.GitProvider{ID: "github", Name: "GitHub", Tool: "GitHub CLI"}
	_, err := exec.LookPath("gh")
	if err != nil {
		cmd, reason := installCommand(ctx)
		p.CanInstall = len(cmd) > 0
		p.Reason = reason
		return p
	}
	p.Installed = true
	help, err := exec.CommandContext(ctx, "gh", "auth", "token", "--help").Output()
	if err != nil || !strings.Contains(string(help), "--user") {
		p.Reason = "Update GitHub CLI to a version with multi-account support."
		return p
	}
	p.Accounts, err = SavedAccounts()
	if err != nil {
		p.Reason = "GitHub CLI configuration is unreadable."
		return p
	}
	p.CanEnroll = enrollmentAvailable(ctx)
	if !p.CanEnroll {
		p.Reason = "Remote enrollment requires an unlocked Linux Secret Service and a private memory-backed staging directory. You can add an account locally with gh auth login."
	}
	return p
}

func installGitHub(ctx context.Context) error {
	_, err := exec.LookPath("gh")
	if err == nil {
		return nil
	}
	args, reason := installCommand(ctx)
	if len(args) == 0 {
		return errors.New(reason)
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	err = cmd.Run()
	if err != nil {
		return errors.New("installation failed; check the host package source and service account permissions")
	}
	_, err = exec.LookPath("gh")
	if err != nil {
		return errors.New("installation finished but gh is unavailable on the service PATH")
	}
	return nil
}

// ProviderDriver declares the operations a trusted local adapter supports.
type ProviderDriver struct {
	ID       string
	Inspect  func(context.Context) domain.GitProvider
	Install  func(context.Context) error
	Enroll   func(context.Context, func() bool, func(string, string)) (string, error)
	Accounts func() ([]string, error)
	Clone    func(context.Context, string, string, string, string, string, func(string)) (domain.Project, error)
}

var providerDrivers = []ProviderDriver{
	{ID: "github", Inspect: githubProvider, Install: installGitHub, Enroll: Enroll, Accounts: SavedAccounts, Clone: cloneCheckout},
}

func Provider(id string) (ProviderDriver, bool) {
	for _, d := range providerDrivers {
		if d.ID == id {
			return d, true
		}
	}
	return ProviderDriver{}, false
}

func Providers(ctx context.Context) []domain.GitProvider {
	out := make([]domain.GitProvider, 0, len(providerDrivers))
	for _, d := range providerDrivers {
		out = append(out, d.Inspect(ctx))
	}
	return out
}

func InstallProvider(ctx context.Context, id string) error {
	d, ok := Provider(id)
	if !ok {
		return errors.New("unknown provider")
	}
	return d.Install(ctx)
}
