// Retains a fixed ring of payload-free workspace observations.
// Snapshotting copies bounded memory; encoding and disk writes run on the caller's background worker.
package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Event struct {
	Seq         uint64 `json:"seq"`
	At          int64  `json:"utc_ms"`
	Operation   string `json:"operation"`
	Elapsed     int64  `json:"duration_us"`
	Received    int    `json:"input_bytes"`
	Sent        int    `json:"output_bytes"`
	Rows        int    `json:"rows"`
	Failed      bool   `json:"failed"`
	Correlation string `json:"correlation,omitempty"`
}

const capacity = 1024

type Ring struct {
	mu     sync.Mutex
	events [capacity]Event
	next   uint64
}
type Capture struct {
	Schema int     `json:"schema"`
	Role   string  `json:"role"`
	Build  string  `json:"build"`
	Lost   uint64  `json:"overwritten"`
	Events []Event `json:"-"`
}

func (r *Ring) Observe(e Event) {
	switch e.Operation {
	case "http-other", "http-workspace", "http-auth", "http-sync", "http-catalog", "list", "read", "save", "tasks", "run", "directory-status", "terminal", "sync", "catalog", "draft", "subscribe":
	default:
		return
	}
	e.At = time.Now().UTC().UnixMilli()
	r.mu.Lock()
	r.next++
	e.Seq = r.next
	r.events[(r.next-1)%capacity] = e
	r.mu.Unlock()
}
func (r *Ring) Snapshot(role, build string, after uint64) Capture {
	r.mu.Lock()
	defer r.mu.Unlock()
	first := max(after+1, 1)
	oldest := uint64(1)
	if r.next >= capacity {
		oldest = r.next - capacity + 1
	}
	c := Capture{Schema: 1, Role: role, Build: build}
	if first < oldest {
		c.Lost = oldest - first
		first = oldest
	}
	for seq := first; seq <= r.next; seq++ {
		c.Events = append(c.Events, r.events[(seq-1)%capacity])
	}
	return c
}
func (c Capture) Write(out io.Writer) error {
	buffer := bufio.NewWriter(out)
	enc := json.NewEncoder(buffer)
	err := enc.Encode(c)
	for _, e := range c.Events {
		if err != nil {
			return err
		}
		err = enc.Encode(e)
	}
	return errors.Join(err, buffer.Flush())
}
func Save(ctx context.Context, dir string, c Capture) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "capture-*.partial")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	err = c.Write(f)
	err = errors.Join(err, f.Sync(), f.Close(), ctx.Err())
	if err != nil {
		return "", err
	}
	path := f.Name()[:len(f.Name())-len(".partial")] + ".ndjson"
	return path, os.Rename(f.Name(), path)
}
func ReadPage(path string, offset int64) ([]string, int64, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	_, err = f.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, 0, err
	}
	reader := bufio.NewReaderSize(f, 4096)
	var rows []string
	for len(rows) < 24 {
		line, err := reader.ReadSlice('\n')
		if err == io.EOF {
			return rows, 0, nil
		}
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, string(line[:len(line)-1]))
		offset += int64(len(line))
	}
	return rows, offset, nil
}

// SavePair publishes a local capture and a verified worker's bounded capture as one NDJSON file.
func SavePair(ctx context.Context, dir string, local Capture, worker string) (string, error) {
	if len(worker) > 512<<10 {
		return "", errors.New("worker capture exceeded limit")
	}
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "paired-*.partial")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	err = local.Write(f)
	if err == nil {
		_, err = io.WriteString(f, worker)
	}
	err = errors.Join(err, ctx.Err(), f.Sync(), f.Close())
	if err != nil {
		return "", err
	}
	path := f.Name()[:len(f.Name())-len(".partial")] + ".ndjson"
	return path, os.Rename(f.Name(), path)
}
