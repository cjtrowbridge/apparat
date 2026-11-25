// Defines owner-hosted sessions shared by project and global callers.
// A turn remains pending until its correlated reply is durably recorded.
package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type ChatScope string

const (
	ProjectChat ChatScope = "project"
	GlobalChat  ChatScope = "global"
)

type ChatHeader struct {
	Model       ModelRef  `yaml:"model,omitempty"`
	Version     int       `yaml:"version"`
	ID          string    `yaml:"session_id"`
	Scope       ChatScope `yaml:"scope"`
	Description string    `yaml:"description"`
	Started     time.Time `yaml:"started_at"`
}

type ChatTurn struct {
	Model            ModelRef
	ID, Input, Reply string
	Stopped          string
	Steps            []ChatStep
}

type ChatSession struct {
	ChatHeader
	Turns []ChatTurn
}

// ChatTime extracts the creation time encoded in a canonical session ID.
func ChatTime(id string) (time.Time, error) {
	if !ValidID(id) {
		return time.Time{}, errors.New("invalid session ID")
	}
	var ms int64
	for _, digit := range id[:10] {
		ms = ms*32 + int64(strings.IndexRune(idAlphabet, digit))
	}
	return time.UnixMilli(ms).UTC(), nil
}

func (s ChatHeader) Validate() error {
	t, err := ChatTime(s.ID)
	if err != nil || s.Version != 1 || !t.Equal(s.Started) || (s.Scope != ProjectChat && s.Scope != GlobalChat) {
		return errors.New("invalid session header")
	}
	if s.Model != (ModelRef{}) && !s.Model.Valid() {
		return errors.New("invalid session model")
	}
	if !chatText(s.Description) || len(s.Description) > 512 || strings.ContainsAny(s.Description, "\r\n") {
		return errors.New("expected a single-line session description of at most 512 bytes")
	}
	return nil
}

func chatText(text string) bool {
	return utf8.ValidString(text) && strings.TrimSpace(text) != "" && !strings.ContainsRune(text, 0)
}

func (s ChatSession) Pending() string {
	n := len(s.Turns)
	if n > 0 && s.Turns[n-1].Reply == "" && s.Turns[n-1].Stopped == "" {
		return s.Turns[n-1].ID
	}
	return ""
}

// Apply records a submission or completion. Identical redelivery is a no-op.
func (s *ChatSession) Apply(id, text string, reply bool) (bool, error) {
	if !ValidID(id) || !chatText(text) {
		return false, errors.New("invalid turn ID or empty/invalid text")
	}
	for i := range s.Turns {
		t := &s.Turns[i]
		if t.ID != id {
			continue
		}
		prior := t.Input
		if reply {
			prior = t.Reply
			if prior == "" {
				if t.Stopped != "" || len(t.Steps) > 0 {
					return false, errors.New("turn is stopped or managed by an inference job")
				}
				t.Reply = text
				return true, nil
			}
		}
		if prior == text {
			return false, nil
		}
		return false, errors.New("turn ID already has different content")
	}
	if reply {
		return false, errors.New("reply refers to an unknown turn")
	}
	if s.Pending() != "" {
		return false, errors.New("session has a pending turn")
	}
	s.Turns = append(s.Turns, ChatTurn{ID: id, Input: text, Model: s.Model})
	return true, nil
}

// PromptTitle fits the opening sentence into a single-line UTF-8 title.
func PromptTitle(prompt string, limit int) string {
	line := strings.Join(strings.Fields(prompt), " ")
	for i, r := range line {
		end := i + utf8.RuneLen(r)
		if end > limit {
			return strings.TrimSpace(line[:i])
		}
		if strings.ContainsRune("。！？", r) || strings.ContainsRune(".!?", r) && (end == len(line) || line[end] == ' ') {
			return line[:end]
		}
	}
	return line
}

// LocalThreadChoice decodes the stable draft ID and title of a document-provider thread.
func LocalThreadChoice(key string) WorkspaceChoice {
	id, title, found := strings.Cut(key, " ")
	if !found || !ValidID(id) {
		title = key
	}
	return WorkspaceChoice{ID: key, Label: title}
}
