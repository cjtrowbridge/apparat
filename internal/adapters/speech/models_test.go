// Checks that cancelled or corrupt model downloads preserve the installed bytes.
// The cases also check temporary file cleanup after each failure.
package speech

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model")
	m := Model{Bytes: 4, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte("good")))}
	err := os.WriteFile(path, []byte("good"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"bad!", "extra", "cut", "cancel"} {
		ctx, cancel := context.WithCancel(context.Background())
		if input == "cancel" {
			cancel()
		}
		_, err := installStream(ctx, path, m, strings.NewReader(input), nil)
		cancel()
		if err == nil {
			t.Fatalf("accepted %q", input)
		}
		kept, err := os.ReadFile(path)
		files, listErr := os.ReadDir(dir)
		if err != nil || string(kept) != "good" || listErr != nil || len(files) != 1 {
			t.Fatal("previous model or temp cleanup failed")
		}
	}
}

type fixtureTransport struct{}

func (fixtureTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("good"))}, nil
}
func TestDownloadIntoEmptyDirectory(t *testing.T) {
	previous := http.DefaultTransport
	http.DefaultTransport = fixtureTransport{}
	defer func() {
		http.DefaultTransport = previous
	}()
	m := Model{ID: "test", File: "test.bin", Bytes: 4, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte("good")))}
	path, err := Install(context.Background(), t.TempDir(), m, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "good" {
		t.Fatal("download missing", err)
	}
}
