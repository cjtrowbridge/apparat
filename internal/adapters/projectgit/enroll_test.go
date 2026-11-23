// Checks enrollment through a synthetic GH executable without contacting GitHub.
// The fixture verifies code relay, preservation of existing users, and rejection of insecure storage.
package projectgit

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStagedGitHubEnrollment(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux staging adapter")
	}
	for _, mode := range []string{"secure", "fallback", "revoked"} {
		t.Run(mode, func(t *testing.T) {
			bin := t.TempDir()
			config := t.TempDir()
			ram := t.TempDir()
			t.Setenv("GH_CONFIG_DIR", config)
			t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
			t.Setenv("FIXTURE_MODE", mode)
			original := "github.com:\n  user: old\n  users:\n    old: {}\n"
			err := os.WriteFile(filepath.Join(config, "hosts.yml"), []byte(original), 0600)
			if err != nil {
				t.Fatal(err)
			}
			const script = `#!/bin/sh
case "$*" in
"auth login --web --hostname github.com --insecure-storage")
 echo '! First copy your one-time code: A123-B456' >&2
 echo 'Open this URL to continue in your web browser: https://github.com/login/device' >&2 ;;
"auth token --hostname github.com"|"auth token --hostname github.com --user new") echo synthetic-token ;;
"auth login --with-token --hostname github.com")
 read t
 [ "$t" = synthetic-token ] || exit 1
 printf 'github.com:\n  user: new\n  users:\n    old: {}\n    new: {}\n' > "$GH_CONFIG_DIR/hosts.yml"
 if [ "$FIXTURE_MODE" = fallback ]; then echo '  oauth_token: synthetic-token' >> "$GH_CONFIG_DIR/hosts.yml"; fi ;;
"auth switch --hostname github.com --user old") sed -i 's/user: new/user: old/' "$GH_CONFIG_DIR/hosts.yml" ;;
*) exit 2 ;;
esac
`
			err = os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0700)
			if err != nil {
				t.Fatal(err)
			}
			gotCode := false
			who, err := enrollIn(context.Background(), ram, func() bool {
				return mode != "revoked"
			}, func(message, code string) {
				if code != "" {
					gotCode = code == "A123-B456"
				}
			})
			data, readErr := os.ReadFile(filepath.Join(config, "hosts.yml"))
			if !gotCode || readErr != nil || strings.Contains(string(data), "synthetic-token") {
				t.Fatal("unsafe enrollment")
			}
			if mode == "secure" {
				if who != "new" || err != nil || !strings.Contains(string(data), "user: old") || !strings.Contains(string(data), "new:") {
					t.Fatal("account enrollment failed", err)
				}
			} else if err == nil || string(data) != original {
				t.Fatal("unauthorized or insecure enrollment saved")
			}
			files, err := os.ReadDir(ram)
			if err != nil || len(files) != 0 {
				t.Fatal("staging retained")
			}
		})
	}
}
