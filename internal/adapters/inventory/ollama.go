// Probes only the worker's IPv4 loopback Ollama metadata API.
// Redirects, proxies and inference endpoints are excluded from discovery.
package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"
)

var client = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 2, IdleConnTimeout: 30 * time.Second}, Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}}

func metadata(ctx context.Context, port int, path string, body []byte, out any) error {
	method := "GET"
	if body != nil {
		method = "POST"
	}
	addr := fmt.Sprintf("http://127.0.0.1:%d/api/%s", port, path)
	req, err := http.NewRequestWithContext(ctx, method, addr, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w. Check that Ollama is listening on this worker's loopback port", path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Ollama %s returned HTTP %d on port %d; check the port and server version", path, resp.StatusCode, port)
	}
	const maxBytes = 4 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil || len(data) > maxBytes {
		return fmt.Errorf("Ollama %s returned an incomplete or oversized inventory", path)
	}
	if !json.Valid(data) {
		return fmt.Errorf("Ollama %s returned invalid JSON on port %d; another service may be listening", path, port)
	}
	return json.Unmarshal(data, out)
}

type taggedModel struct {
	Name, Digest string
	Size         int64
}
type tags struct{ Models *[]taggedModel }

func safeName(s string, limit int) bool {
	return s != "" && len(s) <= limit && strings.IndexFunc(s, unicode.IsControl) < 0
}
func probe(ctx context.Context, port int, cache map[string]bool) (domain.WorkerInventory, error) {
	v := domain.WorkerInventory{Port: port}
	var version struct{ Version string }
	err := metadata(ctx, port, "version", nil, &version)
	if err != nil {
		return v, err
	}
	if !safeName(version.Version, 80) {
		return v, errors.New("missing Ollama version in /api/version; verify this port serves Ollama")
	}
	var installed, running tags
	err = metadata(ctx, port, "tags", nil, &installed)
	if err != nil {
		return v, err
	}
	err = metadata(ctx, port, "ps", nil, &running)
	if err != nil {
		return v, err
	}
	if installed.Models == nil || running.Models == nil || len(*installed.Models) > 2048 {
		return v, errors.New("invalid or oversized Ollama model inventory; expected models arrays from /api/tags and /api/ps")
	}
	loaded := map[string]bool{}
	for _, m := range *running.Models {
		loaded[m.Name] = true
	}
	next := map[string]bool{}
	for _, m := range *installed.Models {
		if !safeName(m.Name, 256) || !safeName(m.Digest, 128) || m.Size < 0 {
			return v, errors.New("invalid model name, size or digest in Ollama inventory")
		}
		text, known := cache[m.Digest]
		if !known {
			args, _ := json.Marshal(map[string]string{"model": m.Name})
			var info struct{ Capabilities []string }
			err = metadata(ctx, port, "show", args, &info)
			if err != nil {
				return v, fmt.Errorf("checking capabilities for %s: %w", m.Name, err)
			}
			for _, kind := range info.Capabilities {
				if kind == "completion" {
					text = true
				}
			}
		}
		cache[m.Digest] = text
		next[m.Digest] = text
		v.Models = append(v.Models, domain.InventoryModel{Name: m.Name, Digest: m.Digest, Size: m.Size, Loaded: loaded[m.Name], Text: text})
	}
	clear(cache)
	for k, b := range next {
		cache[k] = b
	}
	sort.Slice(v.Models, func(i, j int) bool {
		return v.Models[i].Name < v.Models[j].Name
	})
	v.Version = version.Version
	return v, nil
}
