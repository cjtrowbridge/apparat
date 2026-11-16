// Verifies configures structured logging and redacts sensitive fields.
// Uses package-local contracts and Go standard-library types.
package logging

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestJSONLRedactsSensitiveFields(t *testing.T) {
	path := t.TempDir() + "/apparat.log"
	logger := New(path, 1024, 2)
	{
		err := logger.Write(Event{Component: "test", Event: "redact", Fields: map[string]any{"token": "abc", "ok": "yes", "raw_prompt": "secret"}})
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "abc") || strings.Contains(text, "secret") {
		t.Fatalf("sensitive value leaked: %s", text)
	}
	if !strings.Contains(text, "[redacted]") {
		t.Fatalf("redaction marker missing: %s", text)
	}
}

func TestRotationCreatesBoundedFile(t *testing.T) {
	path := t.TempDir() + "/apparat.log"
	logger := New(path, 10, 2)
	{
		err := logger.Write(Event{Component: "test", Event: "one", Fields: map[string]any{"value": strings.Repeat("x", 20)}})
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		err := logger.Write(Event{Component: "test", Event: "two"})
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		_, err := os.Stat(path + ".1")
		if err != nil {
			t.Fatalf("rotated log missing: %v", err)
		}
	}
}

func TestLastRunResetsAndRedacts(t *testing.T) {
	path := t.TempDir() + "/last_run.log"
	{
		err := os.WriteFile(path, []byte("stale"), 0o600)
		if err != nil {
			t.Fatal(err)
		}
	}
	lastRun, err := StartLastRun(path, map[string]any{"binary": "apparat", "token": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	{
		err := lastRun.Write("info", "database", "ready", "database ready", map[string]any{"private_key": "secret"})
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "stale") || strings.Contains(text, "abc") || strings.Contains(text, "secret") {
		t.Fatalf("last_run leaked stale or sensitive data: %s", text)
	}
	if !strings.Contains(text, "process_start") || !strings.Contains(text, "database") {
		t.Fatalf("last_run missing diagnostics: %s", text)
	}
}

func TestNestedSecretRedaction(t *testing.T) {
	fields := map[string]any{"nested": []any{map[string]any{"Token": "hidden-token", "safe": "visible"}, struct {
		Password string
		Name     string
	}{"hidden-password", "public"}}, "bytes": []byte("hidden-bytes")}
	data, err := json.Marshal(Redact(fields))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "hidden") || !strings.Contains(string(data), "visible") {
		t.Fatal(string(data))
	}
}
