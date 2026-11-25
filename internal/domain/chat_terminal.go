// Describes a model-requested terminal operation awaiting the originating client.
// Each callback is fenced by a stable model job and a finite owner-clock deadline.
package domain

import "time"

type ChatTerminalWait struct {
	Job, CallID, Command string
	Deadline             time.Time
	Observe              bool
}

func (s ChatSession) TerminalWait() *ChatTerminalWait {
	if s.Pending() == "" {
		return nil
	}
	t := s.Turns[len(s.Turns)-1]
	if len(t.Steps) == 0 {
		return nil
	}
	return t.Steps[len(t.Steps)-1].Wait
}

func (s *ChatSession) FinishTerminalWait(job, content string, failed bool) bool {
	wait := s.TerminalWait()
	if wait == nil || wait.Job != job {
		return false
	}
	if !time.Now().Before(wait.Deadline) {
		content = "Wait expired. Inspect the terminal for its current process state."
		failed = true
	}
	t := &s.Turns[len(s.Turns)-1]
	step := &t.Steps[len(t.Steps)-1]
	name := "terminal_command"
	if wait.Observe {
		name = "terminal_screen"
	}
	step.Results = []ToolResult{{CallID: wait.CallID, Name: name, Content: content, IsError: failed}}
	step.Wait = nil
	return true
}
