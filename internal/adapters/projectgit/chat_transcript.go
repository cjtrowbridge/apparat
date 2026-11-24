// Encodes readable owner-hosted chat transcripts with YAML front matter.
// Quoted message lines preserve arbitrary headings, delimiters and trailing newlines.
package projectgit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"gopkg.in/yaml.v3"
)

func chatBytes(s domain.ChatSession) ([]byte, error) {
	header, err := yaml.Marshal(s.ChatHeader)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(header)
	out.WriteString("---\n")
	write := func(role, id, text string) {
		fmt.Fprintf(&out, "\n## %s %s\n\n> %s\n", role, id, strings.ReplaceAll(text, "\n", "\n> "))
	}
	for _, t := range s.Turns {
		write("User", t.ID, t.Input)
		if t.Model.Valid() {
			raw, err := json.Marshal(t.Model)
			if err != nil {
				return nil, err
			}
			write("Model", t.ID, string(raw))
		}
		for _, step := range t.Steps {
			data, err := json.Marshal(step)
			if err != nil {
				return nil, err
			}
			write("Inference", t.ID, string(data))
		}
		if t.Stopped != "" {
			write("Stopped", t.ID, t.Stopped)
		}
		if t.Reply != "" {
			write("Assistant", t.ID, t.Reply)
		}
	}
	return out.Bytes(), nil
}

func parseChat(raw []byte, id string) (s domain.ChatSession, err error) {
	bad := errors.New("invalid chat transcript; retained original file")
	header, body, ok := strings.Cut(string(raw), "\n---\n")
	if !ok || !strings.HasPrefix(header, "---\n") {
		return s, bad
	}
	decoder := yaml.NewDecoder(strings.NewReader(header[4:]))
	decoder.KnownFields(true)
	if decoder.Decode(&s.ChatHeader) != nil || s.Validate() != nil || s.ID != id {
		return s, bad
	}
	seen := map[string]bool{}
	for body != "" {
		if !strings.HasPrefix(body, "\n## ") {
			return s, bad
		}
		label, rest, ok := strings.Cut(body[4:], "\n\n")
		if !ok {
			return s, bad
		}
		role, turn, ok := strings.Cut(label, " ")
		if !ok || !domain.ValidID(turn) {
			return s, bad
		}
		block, next, more := strings.Cut(rest, "\n\n## ")
		if more {
			body = "\n## " + next
		} else {
			if !strings.HasSuffix(block, "\n") {
				return s, bad
			}
			block = strings.TrimSuffix(block, "\n")
			body = ""
		}
		lines := strings.Split(block, "\n")
		for i, line := range lines {
			if !strings.HasPrefix(line, "> ") {
				return s, bad
			}
			lines[i] = line[2:]
		}
		text := strings.Join(lines, "\n")
		if strings.TrimSpace(text) == "" {
			return s, bad
		}
		switch role {
		case "User":
			if seen[turn] || s.Pending() != "" {
				return s, bad
			}
			seen[turn] = true
			s.Turns = append(s.Turns, domain.ChatTurn{ID: turn, Input: text})
		case "Model":
			if s.Pending() != turn {
				return s, bad
			}
			t := &s.Turns[len(s.Turns)-1]
			if t.Model.Valid() || json.Unmarshal([]byte(text), &t.Model) != nil || !t.Model.Valid() {
				return s, bad
			}
		case "Inference":
			if s.Pending() != turn {
				return s, bad
			}
			var step domain.ChatStep
			err = json.Unmarshal([]byte(text), &step)
			if err != nil || !domain.ValidID(step.Job) || step.Route == "" || len(step.Digest) != 64 {
				return s, bad
			}
			t := &s.Turns[len(s.Turns)-1]
			t.Steps = append(t.Steps, step)
		case "Stopped":
			if s.Pending() != turn {
				return s, bad
			}
			s.Turns[len(s.Turns)-1].Stopped = text
		case "Assistant":
			if s.Pending() != turn {
				return s, bad
			}
			s.Turns[len(s.Turns)-1].Reply = text
		default:
			return s, bad
		}
	}
	return s, nil
}
