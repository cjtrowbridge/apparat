// Uses the native Ollama chat exchange so the same inference that produces the answer
// also supplies timing and speculative decoding counters. Never issue a diagnostic replay.
package chatprovider

import (
	"bufio"
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
	body, err := json.Marshal(map[string]any{"model": p.model, "messages": messages, "tools": tools, "stream": true, "options": map[string]int{"num_predict": r.MaxOutputTokens}})
	if err != nil {
		return nil, err
	}
	resp, rejected, err := p.open(ctx, body)
	if err != nil || rejected != nil {
		return rejected, err
	}
	defer resp.Body.Close()
	scan := bufio.NewScanner(resp.Body)
	const limit = 4 << 20
	scan.Buffer(make([]byte, 8192), limit)
	var out domain.ChatOutput
	var text strings.Builder
	callBytes := 0
	for scan.Scan() {
		raw := scan.Bytes()
		var chunk struct {
			Message struct {
				nativeMessage
				Thinking string
			}
			Done   bool
			Finish string `json:"done_reason"`
			Error  string
		}
		if json.Unmarshal(raw, &chunk) != nil || chunk.Error != "" || chunk.Message.Role != "assistant" {
			return nil, errors.New("invalid Ollama stream")
		}
		m := chunk.Message
		text.WriteString(m.Content)
		stage := ""
		if m.Thinking != "" {
			stage = "Thinking"
		}
		if m.Content != "" {
			stage = "Generating reply"
		}
		for _, c := range m.Calls {
			stage = "Preparing tool call"
			callBytes += len(c.Function.Name) + len(c.Function.Arguments) + 32
			out.Calls = append(out.Calls, domain.ToolCall{ID: fmt.Sprintf("%s-%d", r.Job, len(out.Calls)), Name: c.Function.Name, Arguments: string(c.Function.Arguments)})
		}
		if text.Len()+callBytes > limit {
			return nil, errors.New("Ollama output exceeds size limit")
		}
		if p.Preview != nil && stage != "" {
			p.Preview(stage, m.Content)
		}
		if chunk.Done {
			out.Text = text.String()
			out.Metrics = nativeMetrics(raw)
			out.Finish = chunk.Finish
			if len(out.Calls) > 0 {
				out.Finish = "tool_calls"
			}
			return &out, nil
		}
	}
	return nil, errors.New("incomplete Ollama stream")
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
