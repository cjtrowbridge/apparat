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
	Version     int       `yaml:"version"`
	ID          string    `yaml:"session_id"`
	Scope       ChatScope `yaml:"scope"`
	Description string    `yaml:"description"`
	Started     time.Time `yaml:"started_at"`
}

type ChatTurn struct {
	ID, Input, Reply string
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
	if n > 0 && s.Turns[n-1].Reply == "" {
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
	s.Turns = append(s.Turns, ChatTurn{ID: id, Input: text})
	return true, nil
}
