//go:build linux && !android

// Exercises the real download/install boundary using an isolated HTTP server and executable.
// Failures must preserve installed bytes and remove every unpromoted temporary download.
package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const revision = "1234567890123456789012345678901234567890"

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func fixture(t *testing.T, data []byte, mutate func(map[string]any)) (*Client, string) {
	t.Helper()
	path := "releases/linux/amd64/apparat/latest"
	m := map[string]any{"accepted": true, "edition": "ordinary", "ref": "refs/heads/main", "delivery_schema": 1,
		"targets": []map[string]any{{"name": "linux/amd64/apparat", "output": path, "state": "passed", "sha256": checksum(data)}}}
	if mutate != nil {
		mutate(m)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/commits/main":
			_ = json.NewEncoder(w).Encode(map[string]string{"sha": revision})
		case "/" + revision + "/releases/verified.json":
			_ = json.NewEncoder(w).Encode(m)
		case "/" + revision + "/" + path:
			_, _ = w.Write(data)
		default:
			t.Errorf("unpinned or unexpected request: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	client := &Client{HTTP: server.Client(), API: server.URL, Raw: server.URL, Arch: "amd64"}
	target := filepath.Join(t.TempDir(), "apparat")
	err := os.WriteFile(target, []byte("installed original"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	return client, target
}

func testELF(t *testing.T) []byte {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestQualifiedUpdatePinnedAndInstalledAtomically(t *testing.T) {
	data := testELF(t)
	c, target := fixture(t, data, nil)
	r, err := c.Latest(context.Background())
	if err != nil || r.Revision != revision {
		t.Fatalf("release = %#v, %v", r, err)
	}
	candidate, err := c.Download(context.Background(), r, target)
	if err != nil {
		t.Fatal(err)
	}
	defer candidate.Discard()
	old, _ := os.Open(target)
	defer old.Close()
	err = candidate.Install(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	digest, _ := Digest(target)
	if digest != checksum(data) {
		t.Fatal("installed bytes differ from qualified artifact")
	}
	previous, _ := os.ReadFile(target + ".previous")
	if string(previous) != "installed original" {
		t.Fatal("previous executable not preserved")
	}
	info, _ := os.Stat(target)
	if info.Mode().Perm() != 0755 {
		t.Fatal("executable permissions lost")
	}
	// An already-open file keeps its original inode through the rename.
	buffer := make([]byte, 18)
	_, _ = old.Read(buffer)
	if string(buffer) != "installed original" {
		t.Fatal("active executable was modified in place")
	}
}

func TestMetadataRejectsUnqualifiedTargets(t *testing.T) {
	cases := map[string]func(map[string]any){
		"not accepted": func(m map[string]any) {
			m["accepted"] = false
		},
		"private edition": func(m map[string]any) {
			m["edition"] = "other"
		},
		"wrong branch": func(m map[string]any) {
			m["ref"] = "refs/heads/other"
		},
		"wrong schema": func(m map[string]any) {
			m["delivery_schema"] = 2
		},
		"wrong target": func(m map[string]any) {
			m["targets"].([]map[string]any)[0]["name"] = "linux/arm64/apparat"
		},
		"traversal": func(m map[string]any) {
			m["targets"].([]map[string]any)[0]["output"] = "../../other"
		},
		"failed": func(m map[string]any) {
			m["targets"].([]map[string]any)[0]["state"] = "failed"
		},
		"bad digest": func(m map[string]any) {
			m["targets"].([]map[string]any)[0]["sha256"] = "not-a-digest"
		},
		"duplicate": func(m map[string]any) {
			m["targets"] = append(m["targets"].([]map[string]any), m["targets"].([]map[string]any)[0])
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c, _ := fixture(t, []byte("test"), mutate)
			_, err := c.Latest(context.Background())
			if err == nil {
				t.Fatal("unsafe metadata accepted")
			}
		})
	}
}

func TestRejectsCorruptionInvalidELFAndCancellation(t *testing.T) {
	for _, kind := range []string{"digest", "ELF", "cancel", "arch"} {
		t.Run(kind, func(t *testing.T) {
			data := testELF(t)
			if kind == "ELF" {
				data = []byte("not an executable")
			}
			c, target := fixture(t, data, nil)
			r, err := c.Latest(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if kind == "digest" {
				r.SHA256 = strings.Repeat("0", 64)
			}
			if kind == "arch" {
				c.Arch = "arm64"
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if kind == "cancel" {
				cancel()
			}
			_, err = c.Download(ctx, r, target)
			if err == nil {
				t.Fatal("invalid update accepted")
			}
			unchanged, _ := os.ReadFile(target)
			if string(unchanged) != "installed original" {
				t.Fatal("failed download changed installed executable")
			}
			staged, _ := filepath.Glob(filepath.Join(filepath.Dir(target), ".apparat-update-*"))
			if len(staged) != 0 {
				t.Fatal("failed candidate was not cleaned up")
			}
		})
	}
}

func TestInstallRejectsChangedFiles(t *testing.T) {
	for _, change := range []string{"installed", "staged", "cancelled"} {
		t.Run(change, func(t *testing.T) {
			c, target := fixture(t, testELF(t), nil)
			r, _ := c.Latest(context.Background())
			candidate, err := c.Download(context.Background(), r, target)
			if err != nil {
				t.Fatal(err)
			}
			defer candidate.Discard()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch change {
			case "installed":
				_ = os.WriteFile(target, []byte("external update"), 0755)
			case "staged":
				_ = os.WriteFile(candidate.Path, []byte("tampered"), 0755)
			case "cancelled":
				cancel()
			}
			before, _ := Digest(target)
			err = candidate.Install(ctx)
			after, _ := Digest(target)
			if err == nil || before != after {
				t.Fatal("failed promotion changed installed bytes")
			}
		})
	}
}
