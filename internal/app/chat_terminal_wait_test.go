// Exercises owner-side validation, the deadline and single-delivery tool continuation.
// Tests use an inert command string and never open a terminal or network connection.
package app

import (
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func TestTerminalWaitContract(t *testing.T) {
	call := domain.ToolCall{ID: "request", Name: "terminal_command", Arguments: `{"command":"pwd","timeout_seconds":"1800"}`}
	now := time.Now()
	makeWait := func() *domain.ChatTerminalWait {
		return terminalWait("job", []domain.ToolCall{call}, now)
	}
	wait := makeWait()
	if wait == nil || !wait.Deadline.Equal(now.Add(30*time.Minute)) {
		t.Fatal(wait)
	}
	s := domain.ChatSession{Turns: []domain.ChatTurn{{ID: "turn", Steps: []domain.ChatStep{{Job: "job", Wait: wait}}}}}
	if s.FinishTerminalWait("different", "wrong", false) || !s.FinishTerminalWait("job", "done", false) || s.FinishTerminalWait("job", "again", false) {
		t.Fatal("callback fencing failed")
	}
	for _, args := range []string{`{"command":"pwd"}`, `{"command":"pwd","timeout_seconds":"0"}`, `{"command":"pwd","timeout_seconds":"1801"}`, `{"command":"pwd","command":"id","timeout_seconds":"30"}`} {
		call.Arguments = args
		if makeWait() != nil {
			t.Fatal(args)
		}
	}
	wait.Deadline = now.Add(-time.Second)
	s.Turns[0].Steps[0].Wait = wait
	if !s.FinishTerminalWait("job", "late success", false) || !s.Turns[0].Steps[0].Results[0].IsError {
		t.Fatal("late success escaped deadline")
	}
}
