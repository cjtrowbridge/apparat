package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Logger struct {
	path      string
	maxBytes  int64
	retention int
}

type Event struct {
	Time        string         `json:"time"`
	Component   string         `json:"component"`
	Event       string         `json:"event"`
	DeviceID    string         `json:"device_id,omitempty"`
	ProjectID   string         `json:"project_id,omitempty"`
	JobID       string         `json:"job_id,omitempty"`
	TaskID      string         `json:"task_id,omitempty"`
	Correlation string         `json:"correlation_id,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
}

func New(path string, maxBytes int64, retention int) Logger {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	if retention <= 0 {
		retention = 3
	}
	return Logger{path: path, maxBytes: maxBytes, retention: retention}
}

func (logger Logger) Write(event Event) error {
	if event.Time == "" {
		event.Time = time.Now().UTC().Format(time.RFC3339Nano)
	}
	event.Fields = Redact(event.Fields)
	if err := os.MkdirAll(filepath.Dir(logger.path), 0o700); err != nil {
		return err
	}
	if err := logger.rotateIfNeeded(); err != nil {
		return err
	}
	file, err := os.OpenFile(logger.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	return err
}

func Redact(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	out := make(map[string]any, len(fields))
	for key, value := range fields {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "private") || strings.Contains(lower, "passphrase") || strings.Contains(lower, "prompt") || strings.Contains(lower, "model_output") || strings.Contains(lower, "voice") || strings.Contains(lower, "secret") {
			out[key] = "[redacted]"
			continue
		}
		out[key] = value
	}
	return out
}

func (logger Logger) rotateIfNeeded() error {
	info, err := os.Stat(logger.path)
	if err != nil {
		return nil
	}
	if info.Size() < logger.maxBytes {
		return nil
	}
	for i := logger.retention - 1; i >= 1; i-- {
		old := logger.path + "." + strconvItoa(i)
		newer := logger.path + "." + strconvItoa(i+1)
		_ = os.Rename(old, newer)
	}
	return os.Rename(logger.path, logger.path+".1")
}

func strconvItoa(value int) string {
	return fmt.Sprintf("%d", value)
}
