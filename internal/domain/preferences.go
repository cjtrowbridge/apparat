// Preferences and desktop geometry are local to one node.
// Values use logical pixels and stable tab identifiers.
package domain

import "slices"

type Preferences struct {
	Placement string
	Theme     string
	Font      int
	Scale     int
	Landing   string
	Overlay   bool
}

func DefaultPreferences() Preferences {
	return Preferences{Placement: "top", Theme: "Dark", Font: 14, Scale: 100, Landing: "comrades"}
}

func (p Preferences) Valid() bool {
	return slices.Contains([]string{"top", "left_rail", "right_rail", "responsive"}, p.Placement) && slices.Contains([]string{"Dark", "High contrast"}, p.Theme) &&
		slices.Contains([]int{14, 16, 18, 20}, p.Font) &&
		slices.Contains([]int{75, 100, 125, 150}, p.Scale) &&
		slices.Contains([]string{"comrades", "projects", "cluster", "research", "stmpck", "settings"}, p.Landing)
}

type WindowState struct {
	Width, Height, X, Y int
	Maximized           bool
	Monitor             string
}
