// Resolves local navigation phrases against feature-provided destinations.
// Addresses identify operations independently of presentation names and query rank.
package hud

import (
	"sort"
	"strings"
	"unicode"
)

type ActionAddress struct {
	Operation, Kind, Owner, Target string
}

type Destination struct {
	Address    ActionAddress
	Name, Path string
}

func CommandWords(s string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}), " ")
}

func navigationPhrase(s string) (string, bool) {
	for _, prefix := range []string{"switch to ", "go to ", "launch ", "open ", "show "} {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimPrefix(s, prefix), true
		}
	}
	return s, false
}

type ActionMatches struct {
	Choices []Destination
	Exact   bool
}

// Voice requires a complete navigation command. Browsing accepts bare searches.
func QueryActions(all []Destination, input string, voice bool) ActionMatches {
	raw := CommandWords(input)
	name, command := navigationPhrase(raw)
	if voice && !command {
		return ActionMatches{}
	}
	filter := ""
	term := name
	for _, k := range []string{"tab", "project", "chat", "thread", "section", "screen", "workspace", "game"} {
		if strings.HasPrefix(name, k+" ") || strings.HasSuffix(name, " "+k) {
			filter = k
			if k == "thread" {
				filter = "chat"
			}
			term = strings.TrimPrefix(name, k+" ")
			if term == name {
				term = strings.TrimSuffix(name, " "+k)
			}
			break
		}
	}
	type match struct {
		item  Destination
		score int
	}
	var ranked []match
	for _, item := range all {
		if command && item.Address.Operation != "open" {
			continue
		}
		target := CommandWords(item.Name)
		path := CommandWords(item.Path)
		if filter != "" && filter != item.Address.Kind && name != path && raw != path {
			continue
		}
		score := 0
		switch {
		case term == target || name == path || raw == path:
			score = 4
		case name == "" && !voice:
			score = 1
		case name != "" && (strings.Contains(target, term) || strings.Contains(path, term)):
			score = 2
		default:
			if nearActionName(term, target) {
				score = 1
			}
		}
		if score > 0 {
			ranked = append(ranked, match{item, score})
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})
	var result ActionMatches
	for _, m := range ranked {
		if voice && result.Exact && m.score < 4 {
			break
		}
		result.Exact = result.Exact || m.score == 4
		result.Choices = append(result.Choices, m.item)
	}
	return result
}

// Allow one insertion, deletion or substitution in names of four or more runes.
func nearActionName(a, b string) bool {
	x, y := []rune(a), []rune(b)
	if len(x) > len(y) {
		x, y = y, x
	}
	if len(x) < 4 || len(y)-len(x) > 1 {
		return false
	}
	i, j, edits := 0, 0, 0
	for i < len(x) && j < len(y) {
		if x[i] != y[j] {
			edits++
			if edits > 1 {
				return false
			}
			if len(x) < len(y) {
				j++
				continue
			}
		}
		i++
		j++
	}
	return edits+len(y)-j <= 1
}
