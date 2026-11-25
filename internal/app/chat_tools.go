// Constructs stable callable declarations and executes the owner-local read-only catalog query.
// Tool results carry call IDs and project scope is enforced independently of model arguments.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

type chatTool struct {
	name, description string
	fields            []string
	query             func(*Runtime, context.Context, string, string, domain.ChatScope, map[string]string) (string, error)
}

// Read-only tools may be rerun if the owner crashes before the result is committed.
var chatTools = []chatTool{
	{"projects", "List registered projects visible in this session on this owner. Filter by project ID or name; use cursor to continue.", []string{"project", "cursor"}, func(n *Runtime, ctx context.Context, project, session string, scope domain.ChatScope, args map[string]string) (string, error) {
		return n.chatProjectList(ctx, project, session, scope, args["project"], args["cursor"])
	}},
	{"list_chats", "List up to 32 transcript descriptions in this session's owning repository. Continue with the last session ID as cursor.", []string{"cursor"}, func(n *Runtime, ctx context.Context, project, _ string, _ domain.ChatScope, args map[string]string) (string, error) {
		headers, err := n.Chats(ctx, project, args["cursor"])
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(headers)
		return string(data), err
	}},
}

func (t chatTool) declaration() domain.ChatTool {
	properties := map[string]map[string]string{}
	for _, field := range t.fields {
		properties[field] = map[string]string{"type": "string"}
	}
	schema, _ := json.Marshal(map[string]any{"type": "object", "properties": properties, "additionalProperties": false})
	return domain.ChatTool{Name: t.name, Description: t.description, Parameters: schema}
}

const maxChatPacketBytes = 128 * 1024

func chatPacket(a domain.ChatAddress, route, job string, s domain.ChatSession, policy string, legacy bool) (domain.ChatRequest, error) {
	p := domain.ChatRequest{ChatAddress: a, Job: job, Route: route, MaxOutputTokens: 4096}
	if len(s.Turns) > 0 {
		p.Model = s.Turns[len(s.Turns)-1].Model
	}
	p.System = chatSystem(&p, policy, legacy)
	if !legacy {
		p.Messages = []domain.ChatMessage{{Role: "assistant", Text: domain.ChatReadyReply}}
	}

	// Reserve space for the shared envelope and keep a suffix of complete user turns.
	base, err := json.Marshal(p)
	if err != nil {
		return p, err
	}
	remaining := maxChatPacketBytes - len(base)
	var groups [][]domain.ChatMessage
	for i := len(s.Turns) - 1; i >= 0; i-- {
		group := chatMessages(s.Turns[i])
		encoded, err := json.Marshal(group)
		if err != nil {
			return p, err
		}
		if len(encoded) > remaining {
			if len(groups) == 0 {
				return p, errors.New("active turn exceeds provider packet budget")
			}
			break
		}
		groups = append(groups, group)
		remaining -= len(encoded)
	}
	for i := len(groups) - 1; i >= 0; i-- {
		p.Messages = append(p.Messages, groups[i]...)
	}
	return p, nil
}

func chatMessages(t domain.ChatTurn) []domain.ChatMessage {
	var messages []domain.ChatMessage
	messages = append(messages, domain.ChatMessage{Role: "user", Text: t.Input})
	for _, step := range t.Steps {
		out := step.Output
		if out == nil || out.Finish != "tool_calls" {
			continue
		}
		messages = append(messages, domain.ChatMessage{Role: "assistant", Text: out.Text, Calls: out.Calls})
		for _, r := range step.Results {
			body := r.Content
			if r.IsError {
				body = "Tool error: " + body
			}
			messages = append(messages, domain.ChatMessage{Role: "tool", CallID: r.CallID, Text: body})
		}
	}
	if t.Reply != "" {
		messages = append(messages, domain.ChatMessage{Role: "assistant", Text: t.Reply})
	}
	if t.Stopped != "" {
		messages = append(messages, domain.ChatMessage{Role: "assistant", Text: "[Turn stopped: " + t.Stopped + "]"})
	}
	return messages
}

func (n *Runtime) chatToolBatch(ctx context.Context, project, session string, scope domain.ChatScope, calls []domain.ToolCall) ([]domain.ToolResult, error) {
	if len(calls) == 0 || len(calls) > 16 {
		return nil, errors.New("expected one to sixteen tool calls")
	}
	seen := map[string]bool{}
	for _, c := range calls {
		if c.ID == "" || len(c.ID) > 256 || seen[c.ID] {
			return nil, errors.New("invalid or repeated tool call ID")
		}
		seen[c.ID] = true
	}
	var results []domain.ToolResult
	for _, c := range calls {
		r := domain.ToolResult{CallID: c.ID, Name: c.Name, IsError: true, Content: "Unknown tool. Use a declared tool."}
		for _, t := range chatTools {
			if t.name != c.Name {
				continue
			}
			args, valid := t.arguments(c.Arguments)
			if !valid {
				r.Content = "Invalid arguments. Use the declared string fields and a valid continuation ID."
				break
			}
			content, err := t.query(n, ctx, project, session, scope, args)
			if err != nil {
				return nil, err
			}
			r.Content = content
			r.IsError = false
			break
		}
		results = append(results, r)
	}
	return results, nil
}

func (n *Runtime) chatProjectList(ctx context.Context, project, session string, scope domain.ChatScope, selector, cursor string) (string, error) {
	user, global := globalUser(project)
	if global && scope == domain.GlobalChat && n.inference != nil {
		return n.inference.globalProjects(ctx, user, session, selector, cursor)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.state != "ready" {
		return "", domain.ErrUnavailable
	}
	if scope == domain.GlobalChat {
		project = ""
	}
	list, next, err := n.database.ChatCatalog(ctx, project, selector, cursor)
	if err != nil {
		return "", err
	}
	result := struct {
		Projects []domain.ChatProject `json:"projects"`
		Cursor   string               `json:"next_cursor,omitempty"`
	}{list, next}
	data, err := json.Marshal(result)
	return string(data), err
}

func (t chatTool) arguments(raw string) (map[string]string, bool) {
	d := json.NewDecoder(strings.NewReader(raw))
	open, err := d.Token()
	if err != nil || open != json.Delim('{') {
		return nil, false
	}
	args := map[string]string{}
	for d.More() {
		key, err := d.Token()
		field, ok := key.(string)
		_, duplicate := args[field]
		if err != nil || !ok || duplicate || !slices.Contains(t.fields, field) {
			return nil, false
		}
		value, err := d.Token()
		text, ok := value.(string)
		if err != nil || !ok {
			return nil, false
		}
		args[field] = text
	}
	closeToken, err := d.Token()
	if err != nil || closeToken != json.Delim('}') || d.Decode(new(any)) != io.EOF {
		return nil, false
	}
	return args, args["cursor"] == "" || domain.ValidID(args["cursor"])
}
