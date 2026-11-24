// Defines provider-neutral chat packets, completed model messages and durable step receipts.
// Provider transports and trusted UI callers share these types without GUI dependencies.
package domain

import (
	"context"
	"encoding/json"
)

type ChatAddress struct {
	Project, Session, Turn string
}

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolResult struct {
	CallID  string `json:"toolCallId"`
	Name    string `json:"toolName"`
	Content string `json:"content"`
	IsError bool   `json:"isError"`
}

type ChatOutput struct {
	Finish string     `json:"finish"`
	Text   string     `json:"text,omitempty"`
	Calls  []ToolCall `json:"calls,omitempty"`
}

type ChatStep struct {
	Job     string       `json:"job"`
	Route   string       `json:"route"`
	Digest  string       `json:"packet"`
	Output  *ChatOutput  `json:"output,omitempty"`
	Results []ToolResult `json:"results,omitempty"`
}

type ChatMessage struct {
	Role, Text, CallID string
	Calls              []ToolCall
}

type ChatTool struct {
	Name, Description string
	Parameters        json.RawMessage
}

type ChatRequest struct {
	ChatAddress
	Model              ModelRef `json:"model,omitzero"`
	Job, Route, System string
	Messages           []ChatMessage
	Tools              []ChatTool
	MaxOutputTokens    int
}

// Dispatch accepts one newly claimed job. A nil output means asynchronous delivery.
// The caller can deliver the final output by job ID even after a runtime restart.
type ChatProvider interface {
	Dispatch(context.Context, ChatRequest) (*ChatOutput, error)
}

type ChatProject struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}
