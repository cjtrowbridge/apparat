// Covers interrupted checkout publication and account isolation for Linux clones.
// All filesystem fixtures are local and existing destination contents must survive.
package projectgit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFailedClonePreservesDestination(t *testing.T) {
	parent := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Clone(ctx, "", parent, "example/project", "")
	if err == nil {
		t.Fatal("cancelled transfer succeeded")
	}
	files, err := os.ReadDir(parent)
	if err != nil || len(files) != 0 {
		t.Fatal("cancelled transfer left a checkout or temporary data", err)
	}
	checkout := filepath.Join(parent, "project")
	err = os.Mkdir(checkout, 0700)
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(checkout, "my-uncommitted-file")
	err = os.WriteFile(marker, []byte("local edit"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Clone(ctx, "", parent, "example/project", "")
	if err == nil {
		t.Fatal("occupied destination accepted")
	}
	bytes, err := os.ReadFile(marker)
	if err != nil || string(bytes) != "local edit" {
		t.Fatal("local edit lost")
	}
}

func TestOnlyOneCloneCanPublish(t *testing.T) {
	parent := t.TempDir()
	dest := filepath.Join(parent, "project")
	completed := make(chan error, 2)
	for _, name := range []string{"first", "second"} {
		stage, err := os.MkdirTemp(parent, name)
		if err != nil {
			t.Fatal(err)
		}
		go func() {
			completed <- installClone(stage, dest)
		}()
	}
	first, second := <-completed, <-completed
	if (first == nil) == (second == nil) {
		t.Fatal("expected exactly one completed installation", first, second)
	}
}

func TestAccountLookupIsScopedToEachTransfer(t *testing.T) {
	// A deterministic CLI double accepts only the exact lookup for each user.
	dir := t.TempDir()
	helper := filepath.Join(dir, "gh")
	script := "#!/bin/sh\n" +
		"test -z \"$GH_TOKEN$GITHUB_TOKEN\" || exit 2\n" +
		"test \"$1 $2 $3 $4 $5\" = 'auth token --hostname github.com --user' || exit 3\n" +
		"case \"$6\" in personal|work) printf 'test-only-%s' \"$6\";; *) exit 4;; esac\n"
	err := os.WriteFile(helper, []byte(script), 0700)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("GH_TOKEN", "unrelated-ambient-test-value")
	t.Setenv("GITHUB_TOKEN", "another-ambient-test-value")
	results := make(chan bool, 2)
	for _, name := range []string{"personal", "work"} {
		go func() {
			auth, err := accountAuth(context.Background(), name)
			results <- err == nil && auth.Password == "test-only-"+name
		}()
	}
	first, second := <-results, <-results
	if !first || !second {
		t.Fatal("concurrent lookups used the wrong credentials")
	}
	_, err = accountAuth(context.Background(), "unavailable")
	if err == nil {
		t.Fatal("missing account accepted")
	}
}
