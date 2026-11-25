// Exercises native tool-call correlation and numeric telemetry capture together.
// The simulated response includes private content that must stay outside metrics.
package chatprovider

import (
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNativeToolAndMetrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []nativeMessage
			Options  map[string]int
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || len(req.Messages) != 3 || req.Messages[2].ToolName != "example" || req.Options["num_predict"] != 32 {
			t.Error("lost native tool request")
		}
		_, _ = w.Write([]byte(`{"done":true,"done_reason":"stop","message":{"role":"assistant","content":"","tool_calls":[{"function":{"name":"next","arguments":{"x":2}}}]},"eval_count":5,"prompt_eval_cached_count":0,"speculative":{"accepted":4,"debug":"private"},"_debug_info":{"rendered_template":"private"}}`))
	}))
	defer srv.Close()
	p, err := NewCompletions(srv.URL+"/api/chat", "model", "")
	if err != nil {
		t.Fatal(err)
	}
	r := domain.ChatRequest{Job: "job", MaxOutputTokens: 32, Messages: []domain.ChatMessage{{Role: "assistant", Calls: []domain.ToolCall{{ID: "first", Name: "example", Arguments: "{}"}}}, {Role: "tool", CallID: "first", Text: "result"}}}
	out, err := p.Dispatch(t.Context(), r)
	if err != nil || out.Finish != "tool_calls" || len(out.Calls) != 1 || out.Calls[0].Arguments != `{"x":2}` || len(out.Metrics) != 3 || out.Metrics["speculative.accepted"] != 4 {
		t.Fatal("invalid native result", err)
	}
}
