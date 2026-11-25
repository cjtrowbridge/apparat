// Exercises the public terminal path against a loopback Chat Completions fixture.
// Checks real wire encoding, continuation, persisted session discovery and resume.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTerminalConversation(t *testing.T) {
	repo := t.TempDir()
	err := exec.Command("git", "init", repo).Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("APPARAT_RUNTIME_DIR", t.TempDir())
	t.Setenv("APPARAT_CHAT_MODEL", "test-model")
	t.Setenv("APPARAT_CHAT_API_KEY", "test-only")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var packet struct {
			Model    string
			Messages []struct {
				Role, Content string
				CallID        string `json:"tool_call_id"`
			}
			Tools []json.RawMessage
		}
		err := json.NewDecoder(r.Body).Decode(&packet)
		if err != nil || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-only" || packet.Model != "test-model" || len(packet.Tools) != 2 || len(packet.Messages) < 3 || packet.Messages[1].Role != "assistant" || packet.Messages[1].Content != "Ready for prompt." {
			t.Error("invalid request")
			return
		}
		i := calls.Add(1)
		if i == 1 {
			fmt.Fprint(w, `{"choices":[{"finish_reason":"tool_calls","message":{"role":"assistant","tool_calls":[{"id":"c1","type":"function","function":{"name":"projects","arguments":"{}"}}]}}]}`)
			return
		}
		if i == 2 && (len(packet.Messages) != 5 || packet.Messages[4].Role != "tool" || packet.Messages[4].CallID != "c1") {
			t.Error("uncorrelated result")
		}
		fmt.Fprint(w, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"Reply from fixture"}}]}`)
	}))
	defer server.Close()
	t.Setenv("APPARAT_CHAT_ENDPOINT", server.URL)
	var output bytes.Buffer
	err = terminalChat([]string{"--repo", repo}, strings.NewReader("Hello\nlist projects\n/quit\n"), &output)
	if err != nil || !strings.Contains(output.String(), "Reply from fixture") || !strings.Contains(output.String(), `"projects"`) {
		t.Fatal(err, output.String())
	}
	session := strings.Fields(output.String())[1]
	output.Reset()
	err = terminalChat([]string{"--repo", repo, "--list"}, strings.NewReader(""), &output)
	if err != nil || !strings.Contains(output.String(), session) {
		t.Fatal(err, output.String())
	}
	output.Reset()
	err = terminalChat([]string{"--repo", repo, "--session", session}, strings.NewReader("Again\n/quit\n"), &output)
	if err != nil || calls.Load() != 3 || strings.Count(output.String(), "Reply from fixture") != 2 {
		t.Fatal(err, output.String())
	}
}
