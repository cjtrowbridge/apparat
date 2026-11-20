// Tests the command grammar independently of the GUI, microphone and model.
// Covers exact dispatch, arbitrary names, collisions and suggestion-only results.
package hud

import "testing"

func TestActionQueryGrammar(t *testing.T) {
	list := []Destination{
		{ActionAddress{"open", "project", "projects", "p1"}, "Apparat", "Projects / Apparat"},
		{ActionAddress{"open", "chat", "p2", "Apparat"}, "Apparat", "Projects / Build / Chat / Apparat"},
		{ActionAddress{"run", "task", "p1", "build"}, "Apparat", "Projects / Apparat / Tasks / Build"},
		{ActionAddress{"open", "project", "projects", "p3"}, "Apparats", "Projects / Apparats"},
	}
	for _, test := range []struct {
		input string
		count int
		exact bool
	}{
		{"launch project Apparat!", 1, true}, {"open Apparat project", 1, true},
		{"OPEN APPARAT", 2, true}, {"show thread apparat", 1, true},
		{"open projects / build / chat / apparat", 1, true}, {"open project App", 2, false},
		{"open Appart", 2, false}, {"open nonexistent", 0, false}, {"run Apparat", 0, false},
		{"Apparat", 0, false}, {"", 0, false},
	} {
		got := QueryActions(list, test.input, true)
		if len(got.Choices) != test.count || got.Exact != test.exact {
			t.Errorf("%q: %+v", test.input, got)
		}
	}
	if len(QueryActions(list, "", false).Choices) != len(list) {
		t.Fatal("empty browser must list every action")
	}
}
