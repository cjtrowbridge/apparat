//go:build gui

// Checks each game's voice entry, setup controls and saved-session transitions.
// Exercises the public embedded surface with isolated persistent storage.
package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ebe-boilerplate/surface"
)

func TestGameSetupAndResume(t *testing.T) {
	root := t.TempDir()
	g := NewGameWithRuntimeInfo(RuntimeInfo{RuntimePath: root})
	defer g.Close()
	session := g.gallerySession()
	session.Layout(600, 800)
	read := func(t *testing.T) map[string]json.RawMessage {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, "stmpck", "games.json"))
		if err != nil {
			t.Fatal(err)
		}
		var state map[string]json.RawMessage
		err = json.Unmarshal(data, &state)
		if err != nil {
			t.Fatal(err)
		}
		return state
	}
	for id, name := range surface.GameNames() {
		t.Run(name, func(t *testing.T) {
			open := func() {
				g.voiceCommand("open game " + name)
				if session.Selected != id {
					t.Fatalf("selected %d, want %d", session.Selected, id)
				}
			}
			open()
			if session.Plays[id] != 0 {
				t.Fatal("opening setup started a game")
			}
			switch name {
			case "Territory":
				session.ActivateMenu(300, 250)
			case "Merge", "Planets":
				session.ActivateMenu(300, 180)
			case "Snake", "Bounce", "Bloom", "Stack", "Brick":
				session.ActivateMenu(500, 180)
				session.ActivateMenu(75, 770)
			default:
				session.ActivateMenu(300, 180)
			}
			if session.Plays[id] != 1 {
				t.Fatal("setup did not start the game")
			}
			session.Back()
			saved := read(t)
			field := map[string]string{"Snake": "Speed", "Bounce": "Speed", "Bloom": "Speed", "Stack": "Pace", "Brick": "Level"}[name]
			if field != "" {
				var settings map[string]json.RawMessage
				err := json.Unmarshal(saved[name], &settings)
				if err != nil || string(settings[field]) != "2" {
					t.Fatalf("third choice did not set %s for %s", field, name)
				}
			}
			open()
			session.ActivateMenu(450, 180) // Inspect new game settings, then back out.
			session.Back()
			open()
			session.ActivateMenu(150, 180) // Resume the saved game.
			if session.Plays[id] != 2 {
				t.Fatal("expected one start and one resume")
			}
			after := read(t)
			for field, value := range saved {
				if field != "Plays" && string(after[field]) != string(value) {
					t.Fatalf("resume changed %s", field)
				}
			}
			session.Back()
		})
	}
}
