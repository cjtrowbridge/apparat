// Translates completed OpenAI-compatible Chat Completions exchanges to the owner contract.
// Configuration and credentials stay with the adapter; transport performs one request per claim.
package chatprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

type Completions struct {
	endpoint, model, key string
	client               http.Client
}

// NewCompletions accepts an explicit full endpoint and model. Plain HTTP is loopback-only.
func NewCompletions(endpoint, model, key string) (*Completions, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.TrimSpace(model) == "" {
		return nil, errors.New("invalid chat endpoint or model")
	}
	ip := net.ParseIP(u.Hostname())
	local := u.Hostname() == "localhost" || ip.IsLoopback()
	if u.Scheme != "https" && !(local && u.Scheme == "http") {
		return nil, errors.New("chat endpoint requires HTTPS or loopback HTTP")
	}
	return &Completions{endpoint: endpoint, model: model, key: key, client: http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}, nil
}

type functionCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}
type message struct {
	Role    string         `json:"role"`
	Content string         `json:"content"`
	CallID  string         `json:"tool_call_id,omitempty"`
	Calls   []functionCall `json:"tool_calls,omitempty"`
}

func (p *Completions) Dispatch(ctx context.Context, r domain.ChatRequest) (*domain.ChatOutput, error) {
	if strings.HasSuffix(p.endpoint, "/api/chat") {
		return p.ollama(ctx, r)
	}
	messages := []message{{Role: "system", Content: r.System}}
	for _, m := range r.Messages {
		wire := message{Role: m.Role, Content: m.Text, CallID: m.CallID}
		for _, c := range m.Calls {
			call := functionCall{ID: c.ID, Type: "function"}
			call.Function.Name = c.Name
			call.Function.Arguments = c.Arguments
			wire.Calls = append(wire.Calls, call)
		}
		messages = append(messages, wire)
	}
	var tools []map[string]any
	for _, t := range r.Tools {
		tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": t.Name, "description": t.Description, "parameters": t.Parameters}})
	}
	if r.Model.Valid() && r.Model.Name != p.model {
		return nil, errors.New("provider is configured for a different chat model")
	}
	body, err := json.Marshal(map[string]any{"model": p.model, "messages": messages, "tools": tools, "max_tokens": r.MaxOutputTokens, "stream": false})
	if err != nil {
		return nil, err
	}
	raw, rejected, err := p.send(ctx, body)
	if err != nil || rejected != nil {
		return rejected, err
	}
	var envelope struct {
		Choices []struct {
			Finish  string  `json:"finish_reason"`
			Message message `json:"message"`
		} `json:"choices"`
	}
	err = json.Unmarshal(raw, &envelope)
	if err != nil || len(envelope.Choices) != 1 || envelope.Choices[0].Message.Role != "assistant" {
		return nil, errors.New("invalid provider response")
	}
	choice := envelope.Choices[0]
	out := domain.ChatOutput{Text: choice.Message.Content, Finish: choice.Finish}
	for _, c := range choice.Message.Calls {
		if c.Type != "function" {
			return nil, errors.New("unsupported provider tool call")
		}
		out.Calls = append(out.Calls, domain.ToolCall{ID: c.ID, Name: c.Function.Name, Arguments: c.Function.Arguments})
	}
	return &out, nil
}

func (p *Completions) send(ctx context.Context, body []byte) ([]byte, *domain.ChatOutput, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.key != "" {
		req.Header.Set("Authorization", "Bearer "+p.key)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, errors.New("provider connection interrupted; reconcile outstanding job")
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	switch resp.StatusCode {
	case 400, 401, 403, 404, 405, 413, 415, 422, 429:
		return nil, &domain.ChatOutput{Finish: "error", Text: fmt.Sprintf("Request rejected (HTTP %d)", resp.StatusCode)}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("provider returned HTTP %d", resp.StatusCode)
	}
	const limit = 4 << 20
	raw, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || len(raw) > limit {
		return nil, nil, errors.New("invalid provider response size")
	}
	return raw, nil, nil
}
