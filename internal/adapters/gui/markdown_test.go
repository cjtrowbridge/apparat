//go:build gui

// Tests the actual native draw runs produced by the chat renderer.
// Includes mixed styles, nested containers, alignment, literal text and viewport bounds.
package gui

import (
	"strings"
	"testing"
)

func TestMarkdownBubbleLayout(t *testing.T) {
	source := "Title\n=====\n\n3. **bold *both*** and _italic_\n   - nested\n\n> quote\n\n| Name | Rate |\n|:---|---:|\n| A | ~~slow~~ |\n\n- [x] done\n\n```\n**literal**\n```\n\n\\*escaped\\* [caption](https://example.org) ![diagram](image.png)"
	for _, width := range []int{120, 312, 540} {
		m := layoutMarkdown(source, width, 14)
		runs := map[string]mdRun{}
		for _, r := range m.runs {
			if r.x < 0 || r.x+r.w > float64(width)+.01 {
				t.Fatalf("run outside %d: %+v", width, r)
			}
			if r.fill == nil {
				runs[r.value] = r
			}
		}
		if runs["Title"].style.size <= 14 || !runs["bold"].style.bold || !runs["both"].style.bold || !runs["both"].style.italic || !runs["italic"].style.italic {
			t.Fatal("inline or heading styles", runs)
		}
		if runs["3."].value == "" || runs["nested"].x <= runs["bold"].x || runs["quote"].x == 0 || runs["☑"].value == "" {
			t.Fatal("missing list or block structure", runs)
		}
		if runs["Name"].y != runs["Rate"].y || runs["Rate"].x <= runs["Name"].x || !runs["Name"].style.bold || !runs["slow"].style.strike {
			t.Fatal("table layout", runs)
		}
		if !runs["**literal**"].style.mono || runs["*escaped*"].value == "" || !runs["caption"].style.link || runs["diagram"].value == "" {
			t.Fatal("literal, link or alt text", runs)
		}
	}
	m := layoutMarkdown(strings.Repeat("界", 300), 120, 14)
	w, h := m.size()
	if w > 120 || h < 40 {
		t.Fatal("long token overflow", w, h)
	}
}

func TestMarkdownLineBreaksAndWideTables(t *testing.T) {
	m := layoutMarkdown("a\nb  \nc\n\n~~~lang\n*literal*\n~~~\n\n---\n\n| One | Two | Three | Four |\n|---|---|---|---|\n| 1 | 2 | 3 | 4 |", 100, 14)
	values := map[string][]mdRun{}
	for _, r := range m.runs {
		if r.fill == nil {
			values[r.value] = append(values[r.value], r)
		}
	}
	for _, key := range []string{"a", "b", "c", "*literal*", "Two", "2"} {
		if len(values[key]) == 0 {
			t.Fatal("missing", key)
		}
	}
	if values["a"][0].y != values["b"][0].y || values["c"][0].y <= values["b"][0].y {
		t.Fatal("soft and hard breaks", values)
	}
	if !values["*literal*"][0].style.mono || len(values["Two"]) != 2 || values["2"][0].y <= values["Two"][1].y {
		t.Fatal("stacked field or fenced code", values)
	}
}
