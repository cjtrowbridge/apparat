// Uses the native Ollama chat exchange so the same inference that produces the answer
// also supplies timing and speculative decoding counters. Never issue a diagnostic replay.
package chatprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
)

type nativeFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type nativeCall struct {
	Function nativeFunction `json:"function"`
}
type nativeMessage struct {
	Role     string       `json:"role"`
	Content  string       `json:"content"`
	Calls    []nativeCall `json:"tool_calls,omitempty"`
	ToolName string       `json:"tool_name,omitempty"`
}

func (p *Completions) ollama(ctx context.Context, r domain.ChatRequest) (*domain.ChatOutput, error) {
	if r.Model.Valid() && r.Model.Name != p.model {
		return nil, errors.New("provider model mismatch")
	}
	messages := []nativeMessage{{Role: "system", Content: r.System}}
	names := map[string]string{}
	for _, m := range r.Messages {
		wire := nativeMessage{Role: m.Role, Content: m.Text, ToolName: names[m.CallID]}
		for _, c := range m.Calls {
			args := json.RawMessage(c.Arguments)
			if !json.Valid(args) {
				return nil, errors.New("invalid tool call arguments")
			}
			wire.Calls = append(wire.Calls, nativeCall{nativeFunction{c.Name, args}})
			names[c.ID] = c.Name
		}
		messages = append(messages, wire)
	}
	var tools []map[string]any
	for _, t := range r.Tools {
		tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": t.Name, "description": t.Description, "parameters": t.Parameters}})
	}
	body, err := json.Marshal(map[string]any{"model": p.model, "messages": messages, "tools": tools, "stream": false, "options": map[string]int{"num_predict": r.MaxOutputTokens}})
	if err != nil {
		return nil, err
	}
	raw, rejected, err := p.send(ctx, body)
	if err != nil || rejected != nil {
		return rejected, err
	}
	var reply struct {
		Message nativeMessage
		Done    bool
		Finish  string `json:"done_reason"`
	}
	if json.Unmarshal(raw, &reply) != nil || !reply.Done || reply.Message.Role != "assistant" {
		return nil, errors.New("incomplete Ollama response")
	}
	out := &domain.ChatOutput{Text: reply.Message.Content, Finish: reply.Finish, Metrics: nativeMetrics(raw)}
	for i, c := range reply.Message.Calls {
		out.Calls = append(out.Calls, domain.ToolCall{ID: fmt.Sprintf("%s-%d", r.Job, i), Name: c.Function.Name, Arguments: string(c.Function.Arguments)})
	}
	if len(out.Calls) > 0 {
		out.Finish = "tool_calls"
	}
	return out, nil
}

func nativeMetrics(raw []byte) map[string]float64 {
	var envelope map[string]json.RawMessage
	_ = json.Unmarshal(raw, &envelope)
	metrics := map[string]float64{}
	for key, data := range envelope {
		switch key {
		case "eval_count", "eval_duration", "prompt_eval_count", "prompt_eval_duration", "prompt_eval_cached_count", "total_duration", "load_duration":
		default:
			if !strings.HasPrefix(key, "draft") && !strings.HasPrefix(key, "mtp") && !strings.HasPrefix(key, "specul") {
				continue
			}
		}
		collectMetrics(metrics, key, data, 0)
	}
	return metrics
}
func collectMetrics(out map[string]float64, key string, raw json.RawMessage, depth int) {
	if len(key) > 96 || len(out) >= 64 || len(raw) == 0 {
		return
	}
	if raw[0] == '{' && depth < 2 {
		var nested map[string]json.RawMessage
		if json.Unmarshal(raw, &nested) == nil {
			for k, v := range nested {
				collectMetrics(out, key+"."+k, v, depth+1)
			}
		}
		return
	}
	var n float64
	if raw[0] != 'n' && json.Unmarshal(raw, &n) == nil && n >= 0 {
		out[key] = n
	}
}
