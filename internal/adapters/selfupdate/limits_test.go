//go:build linux && !android

// Verifies early rejection of oversized transfers without allocating the payload.
// Exercises the same download cleanup path as truncated or interrupted bodies.
package selfupdate

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type responseTransport func(*http.Request) (*http.Response, error)

func (f responseTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestOversizedTransferKeepsInstalledBinary(t *testing.T) {
	c, target := fixture(t, []byte("ignored"), nil)
	r, err := c.Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	c.HTTP = &http.Client{Transport: responseTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, ContentLength: MaxBinaryBytes + 1, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	_, err = c.Download(context.Background(), r, target)
	if err == nil {
		t.Fatal("oversized download accepted")
	}
	data, _ := os.ReadFile(target)
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(target), ".apparat-update-*"))
	if string(data) != "installed original" || len(files) != 0 {
		t.Fatal("oversized transfer changed installation or left staging bytes")
	}
}
