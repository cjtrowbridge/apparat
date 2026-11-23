// Reads GitHub CLI account labels from the current OS user's host configuration.
// Only github.com usernames enter the result. Credential material never enters errors or responses.
package projectgit

import (
	"context"
	"errors"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
)

func SavedAccounts() ([]string, error) {
	dir := os.Getenv("GH_CONFIG_DIR")
	if dir == "" {
		dir = os.Getenv("XDG_CONFIG_HOME")
		if dir == "" && runtime.GOOS == "windows" {
			dir = os.Getenv("AppData")
		}
		if dir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, errors.New("GitHub CLI configuration unavailable")
			}
			dir = filepath.Join(home, ".config")
		}
		dir = filepath.Join(dir, "gh")
	}
	f, err := os.Open(filepath.Join(dir, "hosts.yml"))
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, errors.New("GitHub CLI configuration unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return nil, errors.New("invalid GitHub CLI configuration")
	}
	var hosts struct {
		GitHub struct {
			User  string
			Users map[string]struct{}
		} `yaml:"github.com"`
	}
	body, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if err != nil || len(body) > 1<<20 || yaml.Unmarshal(body, &hosts) != nil {
		return nil, errors.New("invalid GitHub CLI configuration")
	}
	users := hosts.GitHub.Users
	if users == nil {
		users = map[string]struct{}{}
	}
	users[hosts.GitHub.User] = struct{}{}
	names := []string{}
	for name := range users {
		if name != "" && len(name) <= 39 && ValidAccount(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// AddLocalAccount delegates interactive authorization to the host terminal.
func AddLocalAccount(ctx context.Context) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.CommandContext(ctx, "x-terminal-emulator", "-e", "gh", "auth", "login", "--hostname", "github.com", "--web")
	case "darwin":
		cmd = exec.CommandContext(ctx, "osascript", "-e", `tell application "Terminal" to do script "gh auth login --hostname github.com --web"`)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd.exe", "/d", "/c", "start", "", "gh", "auth", "login", "--hostname", "github.com", "--web")
	default:
		return errors.New("host terminal unavailable")
	}
	return cmd.Run()
}
