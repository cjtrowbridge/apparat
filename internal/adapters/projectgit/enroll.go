// Relays only the GitHub user code from a noninteractive login. Linux staging lives in tmpfs.
// The system keyring receives credentials only after approval and a fresh authorization check.
package projectgit

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var deviceCode = regexp.MustCompile(`^[A-Z0-9]{4}-[A-Z0-9]{4}$`)

const verificationURL = "https://github.com/login/device"

func enrollmentProblem(ctx context.Context) string {
	if runtime.GOOS != "linux" {
		return "Remote account enrollment is supported on Linux workers. Run gh auth login on this host to add an account."
	}
	out, err := exec.CommandContext(ctx, "stat", "-f", "-c", "%T", "/dev/shm").Output()
	if err != nil || strings.TrimSpace(string(out)) != "tmpfs" {
		return "The worker could not verify /dev/shm as memory-backed staging. Configure tmpfs there or run gh auth login on the worker."
	}
	cmd := exec.CommandContext(ctx, "/usr/bin/python3", "-I", "-c", `import gi
gi.require_version("Secret","1")
from gi.repository import Secret
s=Secret.Service.get_sync(Secret.ServiceFlags.LOAD_COLLECTIONS,None)
c=Secret.Collection.for_alias_sync(s,"default",Secret.CollectionFlags.NONE,None)
if c is None: raise SystemExit(20)
raise SystemExit(21 if c.get_locked() else 0)`)
	err = cmd.Run()
	if err == nil {
		return ""
	}
	if ctx.Err() != nil {
		return "The worker did not finish checking secure storage. Refresh credential providers to retry."
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		switch exit.ExitCode() {
		case 20:
			return "The worker has no default keyring. Create one for the worker service user and refresh, or run gh auth login on the worker."
		case 21:
			return "The worker’s default keyring is locked. Unlock it in the worker service user’s session and refresh, or run gh auth login on the worker."
		}
	}
	return "The worker could not check its default keyring through Linux Secret Service. Check the service user’s keyring session or run gh auth login on the worker."
}

func stagedGH(ctx context.Context, dir string, isolated bool, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "gh", args...)
	env := os.Environ()
	for _, v := range []string{"GH_TOKEN", "GITHUB_TOKEN", "GH_DEBUG", "DEBUG", "GH_CONFIG_DIR", "GH_HOST", "GH_FORCE_TTY", "GH_PROMPT_DISABLED", "GH_BROWSER", "BROWSER"} {
		env = withoutVariable(env, v)
	}
	if isolated {
		env = withoutVariable(env, "DBUS_SESSION_BUS_ADDRESS")
		env = append(env, "DBUS_SESSION_BUS_ADDRESS=unix:path=/dev/null")
	}
	cmd.Env = append(env, "GH_CONFIG_DIR="+dir, "GH_PROMPT_DISABLED=1", "GH_NO_UPDATE_NOTIFIER=1", "NO_COLOR=1")
	return cmd
}

func Enroll(ctx context.Context, authorized func() bool, progress func(string, string)) (string, error) {
	problem := enrollmentProblem(ctx)
	if problem != "" {
		return "", errors.New(problem)
	}
	return enrollIn(ctx, "/dev/shm", authorized, progress)
}

func enrollIn(ctx context.Context, ramDir string, authorized func() bool, progress func(string, string)) (string, error) {
	stage, err := os.MkdirTemp(ramDir, "apparat-gh-")
	if err != nil {
		return "", errors.New("private staging unavailable")
	}
	defer os.RemoveAll(stage)
	cmd := stagedGH(ctx, stage, true, "auth", "login", "--web", "--hostname", "github.com", "--insecure-storage")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	err = cmd.Start()
	if err != nil {
		return "", errors.New("GitHub CLI unavailable")
	}
	scan := bufio.NewScanner(stderr)
	code := ""
	shown := false
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		_, candidate, ok := strings.Cut(line, "First copy your one-time code: ")
		if ok && deviceCode.MatchString(candidate) {
			code = candidate
		}
		if strings.HasSuffix(line, " "+verificationURL) && code != "" {
			progress("Waiting for GitHub approval", code)
			shown = true
		}
	}
	err = cmd.Wait()
	if err != nil || !shown {
		return "", errors.New("GitHub login failed, expired, or returned an unsupported response")
	}
	progress("Verifying account on the worker", "")
	token, err := stagedGH(ctx, stage, true, "auth", "token", "--hostname", "github.com").Output()
	if err != nil {
		return "", errors.New("GitHub did not return an enrolled account")
	}
	defer clear(token)
	if !authorized() {
		return "", errors.New("administrator authorization expired")
	}
	return saveEnrollment(ctx, stage, token, authorized)
}

func saveEnrollment(ctx context.Context, stage string, token []byte, authorized func() bool) (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return "", errors.New("configuration unavailable")
	}
	path := filepath.Join(dir, "hosts.yml")
	prior, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", errors.New("configuration unavailable")
	}
	var hosts map[string]struct{ User string }
	if yaml.Unmarshal(prior, &hosts) != nil {
		return "", errors.New("invalid existing configuration")
	}
	copyDir := filepath.Join(stage, "system")
	err = os.Mkdir(copyDir, 0700)
	if err != nil {
		return "", err
	}
	scratch := filepath.Join(copyDir, "hosts.yml")
	err = os.WriteFile(scratch, prior, 0600)
	if err != nil {
		return "", err
	}
	cmd := stagedGH(ctx, copyDir, false, "auth", "login", "--with-token", "--hostname", "github.com")
	cmd.Stdin = bytes.NewReader(token)
	previous := hosts["github.com"].User
	published := false
	if previous != "" {
		defer func() {
			if published {
				return
			}
			current, readErr := os.ReadFile(path)
			if readErr != nil || !bytes.Equal(current, prior) {
				return
			}
			restore, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// The original file is unchanged on failure; restore its keyring selection too.
			_ = stagedGH(restore, dir, false, "auth", "switch", "--hostname", "github.com", "--user", previous).Run()
		}()
	}
	if cmd.Run() != nil {
		return "", errors.New("could not store GitHub account")
	}
	data, err := os.ReadFile(scratch)
	if err != nil || bytes.Contains(data, bytes.TrimSpace(token)) || yaml.Unmarshal(data, &hosts) != nil {
		return "", errors.New("secure credential storage failed")
	}
	user := hosts["github.com"].User
	if user == "" || !ValidAccount(user) {
		return "", errors.New("invalid enrolled username")
	}
	if previous != "" && previous != user {
		if stagedGH(ctx, copyDir, false, "auth", "switch", "--hostname", "github.com", "--user", previous).Run() != nil {
			return "", errors.New("could not preserve the active account")
		}
	}
	data, err = os.ReadFile(scratch)
	if err != nil || bytes.Contains(data, bytes.TrimSpace(token)) {
		return "", errors.New("secure credential storage failed")
	}
	proof, checkErr := stagedGH(ctx, copyDir, false, "auth", "token", "--hostname", "github.com", "--user", user).Output()
	valid := checkErr == nil && bytes.Equal(bytes.TrimSpace(proof), bytes.TrimSpace(token))
	clear(proof)
	if !valid || !authorized() {
		return "", errors.New("account verification or authorization failed")
	}
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) || !bytes.Equal(prior, current) {
		return "", errors.New("host account configuration changed; retry")
	}
	tmp, err := os.CreateTemp(dir, ".apparat-accounts-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(data)
	err = errors.Join(err, tmp.Close())
	if err != nil {
		return "", errors.New("unable to save account metadata")
	}
	err = os.Rename(tmp.Name(), path)
	published = err == nil
	return user, err
}
