// Runs private workspace captures away from the UI and the SQLite write path.
// The Performance page reads bounded portions of completed captures from the node's data directory.
package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/telemetry"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type captureSession struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	message string
}

func (n *Runtime) captureDir() string {
	return filepath.Join(n.config.RootDir, "telemetry")
}
func (n *Runtime) capture(after uint64) telemetry.Capture {
	return n.observations.Snapshot(string(n.mode), domain.ReleaseVersion(), after)
}
func (n *Runtime) stopCapture() {
	r := &n.recording
	r.mu.Lock()
	cancel, done := r.cancel, r.done
	r.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
}
func (n *Runtime) startCapture() error {
	r := &n.recording
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		return errors.New("recording already active")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	r.cancel = cancel
	r.done = make(chan struct{})
	r.message = "Recording for up to 15 minutes or 4 MiB"
	go n.recordWindow(ctx, cancel, r.done)
	return nil
}
func (n *Runtime) recordWindow(ctx context.Context, cancel context.CancelFunc, done chan struct{}) {
	defer cancel()
	message := "Recording complete"
	defer func() {
		n.recording.mu.Lock()
		n.recording.message = message
		n.recording.cancel = nil
		close(done)
		n.recording.mu.Unlock()
	}()
	dir := n.captureDir()
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		message = "Could not create capture directory"
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var cursor uint64
	var segment bytes.Buffer
	total := 0
	flush := func() error {
		if segment.Len() == 0 {
			return nil
		}
		f, err := os.CreateTemp(dir, "session-*.partial")
		if err != nil {
			return err
		}
		defer os.Remove(f.Name())
		_, err = segment.WriteTo(f)
		err = errors.Join(err, f.Sync(), f.Close())
		if err != nil {
			return err
		}
		return os.Rename(f.Name(), strings.TrimSuffix(f.Name(), ".partial")+".ndjson")
	}
	for {
		c := n.capture(cursor)
		var batch bytes.Buffer
		err = c.Write(&batch)
		if total+batch.Len() > 4<<20 {
			break
		}
		if segment.Len()+batch.Len() > 1<<20 {
			err = flush()
		}
		if err != nil {
			message = "Recording failed"
			return
		}
		total += batch.Len()
		segment.Write(batch.Bytes())
		if len(c.Events) > 0 {
			cursor = c.Events[len(c.Events)-1].Seq
		}
		if ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
	if flush() != nil {
		message = "Recording could not be finalized"
	}
}
func (n *Runtime) performance(ctx context.Context, r domain.ClusterRoute) (domain.ClusterPage, error) {
	p := domain.ClusterPage{Title: "Performance"}
	dir := n.captureDir()
	if r.Target != "" {
		if filepath.Base(r.Target) != r.Target || !strings.HasSuffix(r.Target, ".ndjson") {
			return p, errors.New("unknown capture")
		}
		offset, _ := strconv.ParseInt(r.Pool, 10, 64)
		rows, next, err := telemetry.ReadPage(filepath.Join(dir, r.Target), offset)
		exporter, ok := n.documents.(interface{ ExportCapture(string) error })
		if ok {
			p.Options = append(p.Options, domain.ClusterOption{Label: "Export capture", Run: func(context.Context, string) error {
				return exporter.ExportCapture(filepath.Join(dir, r.Target))
			}})
		}
		p.Title = r.Target
		p.Detail = strings.Join(rows, "\n")
		if next > 0 {
			r.Pool = strconv.FormatInt(next, 10)
			p.Options = append(p.Options, domain.ClusterOption{Label: "Next records", Route: r})
		}
		return p, err
	}
	n.recording.mu.Lock()
	running := n.recording.cancel != nil
	p.Detail = n.recording.message
	n.recording.mu.Unlock()
	p.Options = append(p.Options, domain.ClusterOption{Label: "Save recent telemetry", Run: func(ctx context.Context, _ string) error {
		_, err := telemetry.Save(ctx, dir, n.capture(0))
		return err
	}})
	workers, _ := n.KnownWorkers(ctx)
	for _, worker := range workers {
		if !worker.Admin {
			continue
		}
		p.Options = append(p.Options, domain.ClusterOption{Label: "Save capture with " + worker.DisplayName(), Run: func(ctx context.Context, _ string) error {
			local := n.capture(0)
			client, err := n.openSavedWorker(ctx, worker)
			var peer domain.WorkspaceView
			if err == nil {
				peer, err = client.Workspace(ctx, domain.WorkspaceRequest{Action: "performance-capture", Path: "."})
			}
			if err != nil {
				_, saveErr := telemetry.Save(ctx, dir, local)
				if saveErr != nil {
					return saveErr
				}
				return errors.New("Saved the local capture. Worker capture unavailable.")
			}
			_, err = telemetry.SavePair(ctx, dir, local, peer.Detail)
			return err
		}})
	}
	label := "Record a diagnostic session"
	if running {
		label = "Stop recording"
	}
	p.Options = append(p.Options, domain.ClusterOption{Label: label, Run: func(context.Context, string) error {
		if running {
			n.stopCapture()
			return nil
		}
		return n.startCapture()
	}})
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return p, nil
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ndjson") && e.Type().IsRegular() {
			p.Options = append(p.Options, domain.ClusterOption{Label: e.Name(), Route: domain.ClusterRoute{Page: "performance", Target: e.Name()}})
		}
	}
	p.Detail += fmt.Sprintf("\nCapture directory: %s", dir)
	return p, err
}
