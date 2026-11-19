// Coordinates one in-memory PTT utterance and one loaded model per GUI.
// Capture stops on release even while the decoder is producing a preview.
package speech

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Selection interface {
	SpeechModel(context.Context) (string, error)
	SetSpeechModel(context.Context, string) error
}
type Engine struct {
	mu      sync.Mutex
	dir     string
	store   Selection
	active  string
	loaded  *decoder
	busy    bool
	closed  bool
	workers sync.WaitGroup
	message string
}

func New(root string, store Selection) *Engine {
	id, err := store.SpeechModel(context.Background())
	if err != nil || id == "" {
		id = Default
	}
	return &Engine{dir: filepath.Join(root, "data", "whisper"), store: store, active: id}
}
func (e *Engine) Status() (string, string, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.active, e.message, e.busy
}
func (e *Engine) reserve() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.busy || e.closed {
		return false
	}
	e.workers.Add(1)
	e.busy = true
	return true
}
func (e *Engine) done(message string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.busy = false
	e.message = message
	e.workers.Done()
}

// Close follows cancellation of the GUI lifetime context.
func (e *Engine) Close() {
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()
	e.workers.Wait()
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.loaded != nil {
		e.loaded.close()
		e.loaded = nil
	}
}
func (e *Engine) prepare(ctx context.Context, id string) error {
	m, err := lookup(id)
	if err != nil {
		return err
	}
	path, err := Install(ctx, e.dir, m, func(n int64) {
		e.mu.Lock()
		e.message = fmt.Sprintf("%s · %d / %d MiB", m.Label, n>>20, m.Bytes>>20)
		e.mu.Unlock()
	})
	if err != nil {
		return err
	}
	if e.loaded != nil {
		e.loaded.close()
		e.loaded = nil
	}
	candidate, err := load(path)
	if err != nil {
		return err
	}
	err = e.store.SetSpeechModel(ctx, id)
	if err != nil {
		candidate.close()
		return err
	}
	e.mu.Lock()
	e.active = id
	e.loaded = candidate
	e.mu.Unlock()
	return nil
}
func (e *Engine) Select(ctx context.Context, id string) {
	if !e.reserve() {
		return
	}
	go func() {
		err := e.prepare(ctx, id)
		if err != nil {
			e.done(err.Error())
			return
		}
		e.done("Ready for dictation")
	}()
}

type Result struct{ State, Text string }
type Session struct {
	mu     sync.Mutex
	result Result
	finish chan struct{}
	once   sync.Once
	cancel context.CancelFunc
}

func (s *Session) set(state, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result = Result{state, text}
}
func (s *Session) Poll() Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result
}
func (s *Session) Release() {
	s.once.Do(func() {
		close(s.finish)
	})
}
func (s *Session) Cancel() {
	s.cancel()
}
func (e *Engine) Start(parent context.Context) (*Session, error) {
	if !e.reserve() {
		return nil, errors.New("speech is busy")
	}
	ctx, cancel := context.WithCancel(parent)
	s := &Session{result: Result{State: "starting"}, finish: make(chan struct{}), cancel: cancel}
	go func() {
		text, err := e.record(ctx, s)
		e.done("")
		switch {
		case ctx.Err() != nil:
			s.set("cancelled", "")
		case err != nil:
			s.set("error", err.Error())
		default:
			s.set("final", text)
		}
		cancel()
	}()
	return s, nil
}
func (e *Engine) record(ctx context.Context, s *Session) (string, error) {
	mic, err := capture()
	if err != nil {
		return "", err
	}
	defer mic.close()
	s.set("recording", "")
	type response struct {
		text string
		err  error
	}
	output := make(chan response, 1)
	var jobs sync.WaitGroup
	defer jobs.Wait()
	jobs.Add(1)
	go func() {
		defer jobs.Done()
		var err error
		if e.loaded == nil {
			err = e.prepare(ctx, e.active)
		}
		output <- response{err: err}
	}()
	busy := true
	released := false
	finalizing := false
	committing := false
	prefix := ""
	finish := s.finish
	var last int
	decode := func(samples []float32) {
		busy = true
		last = len(samples)
		jobs.Add(1)
		go func() {
			defer jobs.Done()
			text, err := e.loaded.decode(samples)
			output <- response{text, err}
		}()
	}
	tick := time.NewTicker(750 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			mic.stop()
			return "", ctx.Err()
		case <-finish:
			finish = nil
			released = true
			mic.stop()
			s.set("finishing", s.Poll().Text)
		case r := <-output:
			busy = false
			if r.err != nil {
				return "", r.err
			}
			if finalizing {
				return strings.TrimSpace(prefix + " " + r.text), nil
			}
			preview := strings.TrimSpace(prefix + " " + r.text)
			if committing {
				prefix = preview
				committing = false
				last = 0
			}
			if !released {
				s.set("recording", preview)
			}
		case <-tick.C:
		}
		if busy {
			continue
		}
		n := mic.count()
		if n < 0 {
			return "", errors.New("decoder fell behind microphone capture; retry with a smaller model")
		}
		const chunk = 20 * 16000
		switch {
		case n >= chunk:
			committing = true
			decode(mic.snapshot(chunk, true))
		case released:
			finalizing = true
			decode(mic.snapshot(n, true))
		case n >= last+16000:
			decode(mic.snapshot(n, false))
		}
	}
}
