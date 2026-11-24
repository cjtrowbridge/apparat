//go:build gui

// Checks rendered session navigation, measured header spacing and local turn state.
// Optional screenshot fixtures exercise both sides of a conversation without contacting inference.
package gui

import (
	"os"
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestPlanningSessionNavigation(t *testing.T) {
	for _, width := range []int{390, 1100} {
		g := horizontalGame()
		g.Layout(width, 844)
		selectTabThroughStrip(t, g, hud.TabGlobalSession)
		original := g.planning.current
		original.Description = strings.Repeat("planning ", 24)
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		targets := g.bodyTargets()
		var back, newButton, edit widget.Focuser
		for _, target := range targets {
			switch focusLabel(target.w) {
			case "Back to sessions":
				back = target.w
			case "New session":
				newButton = target.w
			case "Edit session title":
				edit = target.w
			}
		}
		if back == nil || newButton == nil || edit == nil {
			t.Fatal("missing header controls")
		}
		if newButton.GetWidget().Rect.Min.X-back.GetWidget().Rect.Max.X != 8 || edit.GetWidget().Rect.Max.X > width {
			t.Fatal("header control layout", back.GetWidget().Rect, newButton.GetWidget().Rect, edit.GetWidget().Rect)
		}
		g.globalDraft = "first message"
		g.submitPlanning()
		g.globalDraft = "next draft"
		g.submitPlanning()
		if len(original.Turns) != 1 || g.globalDraft != "next draft" {
			t.Fatal("pending turn lost next draft")
		}
		g.enterPlanningVoice()
		if g.planning.current != original {
			t.Fatal("PTT replaced open conversation")
		}
		g.planningBack()
		g.enterPlanningVoice()
		if g.planning.current == original || g.globalDraft != "" {
			t.Fatal("history PTT did not open a new session")
		}
		g.openPlanning(original)
		if g.globalDraft != "next draft" {
			t.Fatal("draft lost on return")
		}
		g.planningPage("rename")
		g.rebuildUI(g.shell.Snapshot())
		g.planning.title = "Build plan"
		g.updatePlanningControls()
		g.planning.saveTitle.Click()
		event.ExecuteDeferred()
		if original.Description != "Build plan" {
			t.Fatal("rename failed")
		}
		g.Close()
	}
}

func TestCapturePlanning(t *testing.T) {
	path := os.Getenv("APPARAT_PLANNING_PNG")
	if path == "" {
		t.Skip("optional rendering")
	}
	w, h := 1280, 800
	if strings.Contains(path, "phone") {
		w, h = 390, 844
	}
	g := horizontalGame()
	g.Layout(w, h)
	selectTabThroughStrip(t, g, hud.TabGlobalSession)
	s := g.planning.current
	s.Description = "Review the cluster inference inventory and next steps"
	g.globalDraft = "What should I check before adding another model?"
	g.submitPlanning()
	_, err := s.Apply(s.Pending(), "Check the available memory on the worker and the size of the model you want to load.", true)
	if err != nil {
		t.Fatal(err)
	}
	g.globalDraft = "Show me the available workers."
	g.submitPlanning()
	s.acknowledged = true
	g.rebuildUI(g.shell.Snapshot())
	ebiten.SetWindowSize(w, h)
	c := &inventoryCapture{g: g, path: path}
	err = ebiten.RunGame(c)
	if err != nil || c.err != nil {
		t.Fatal(err, c.err)
	}
}
