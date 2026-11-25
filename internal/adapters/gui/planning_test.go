//go:build gui

// Checks rendered session navigation, measured header spacing and local turn state.
// Optional screenshot fixtures exercise both sides of a conversation without contacting inference.
package gui

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

type planningTestStore struct {
	ServiceQueries
	planningStore
	saved domain.GlobalRequest
}

func (s *planningTestStore) PlanningUser() string {
	return "test-user"
}

func (s *planningTestStore) SavePlanning(_ context.Context, _ string, _ string, q domain.GlobalRequest) error {
	s.saved = q
	return nil
}

func attachPlanningTestStore(g *Game) *planningTestStore {
	store := &planningTestStore{}
	g.runtimeInfo.Queries = store
	g.planning.sync.user = store.PlanningUser()
	return store
}

func TestPlanningSessionNavigation(t *testing.T) {
	for _, size := range [][2]int{{390, 844}, {390, 424}, {1100, 844}} {
		width := size[0]
		g := horizontalGame()
		g.Layout(width, size[1])
		selectTabThroughStrip(t, g, hud.TabGlobalSession)
		store := attachPlanningTestStore(g)
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
		if g.globalField.GetWidget().Rect.Min.Y <= edit.GetWidget().Rect.Max.Y || g.planning.pane.ViewRect().Dy() < 20 {
			t.Fatal("composer overlaps session toolbar or transcript", size, g.globalField.GetWidget().Rect, edit.GetWidget().Rect, g.planning.pane.ViewRect())
		}
		pane := g.planning.pane
		composer := g.globalField.GetWidget().Rect
		body := g.nav.body.GetWidget().Rect
		if pane.ContentRect().Dy() > pane.ViewRect().Dy() || len(original.Turns) != 0 {
			t.Fatal("expected empty conversation", size, pane.ContentRect(), pane.ViewRect())
		}
		gap := body.Max.Y - max(composer.Max.Y, g.planning.send.GetWidget().Rect.Max.Y)
		if gap > bodyGap {
			t.Fatal("composer above bottom of chat", size, gap)
		}
		pane.ScrollTop = 0
		renderNavigation(g)
		if g.globalField.GetWidget().Rect != composer {
			t.Fatal("scroll moved composer", size)
		}
		editor := g.globalField.GetWidget().CustomData.(*documentSurface)
		editor.insert("first")
		editor.enter(true)
		editor.insert("message")
		if g.globalDraft != "first\nmessage" || store.saved.Turn != "" {
			t.Fatal("Shift+Enter should only insert a newline")
		}
		editor.enter(false)
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if store.saved.Input != "first\nmessage" || store.saved.Session != original.ID || original.Pending() != store.saved.Turn || len(original.Turns) != 1 {
			t.Fatal("first prompt did not reach the planning outbox")
		}
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
	if strings.Contains(path, "project-chat") {
		g.projects.store = layoutRegistry{}
		g.projects.rows = []domain.Project{{ID: "project", Name: "Apparat", Available: true}}
		selectTabThroughStrip(t, g, hud.TabProjects)
		g.detailOpen[hud.TabProjects] = true
		g.selectSection(hud.TabProjects, len(projectEntrySections))
		g.projects.workspace = workspacePanel{id: "project", mode: "conversation", agentThread: true, thread: "session", threadTitle: "Release planning", choices: []domain.WorkspaceChoice{{State: "user", Label: "What needs my attention?"}, {State: "assistant", Label: "The build is waiting for your review."}, {State: "user", Label: "Show the latest result."}, {State: "pending", Label: "Saved to project · awaiting dispatch"}}}
		if strings.Contains(path, "draft") {
			g.openWorkspaceTab("chat", "threads")
		}
	} else if strings.Contains(path, "picker") {
		m := domain.ModelRef{Provider: "ollama", Name: "qwen3.5:4b", Digest: "build-a"}
		g.planning.picker = modelPicker{address: modelAddress{"", g.planning.current.ID}, loaded: true, settings: domain.ModelSettings{CanDefault: true, Session: m, Default: domain.ModelDefault{Model: m}}, options: []domain.ModelOption{{Model: m, Nodes: []string{"trashcan0"}}, {Model: domain.ModelRef{Provider: "ollama", Name: "qwen3.8-flash-next:125b-a6b-q4_K_M", Digest: "build-b"}, Nodes: []string{"trashcan0"}}}}
		g.openModels()
	} else if !strings.Contains(path, "empty") {
		attachPlanningTestStore(g)
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
	}
	g.rebuildUI(g.shell.Snapshot())
	ebiten.SetWindowSize(w, h)
	c := &inventoryCapture{g: g, path: path}
	err := ebiten.RunGame(c)
	if err != nil || c.err != nil {
		t.Fatal(err, c.err)
	}
}

func TestPlanningModelPickerReturnsToDraft(t *testing.T) {
	for _, width := range []int{390, 1200} {
		g := horizontalGame()
		g.Layout(width, 844)
		selectTabThroughStrip(t, g, hud.TabGlobalSession)
		s := g.planning.current
		g.globalDraft = "Draft to keep"
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		p := &g.planning.picker
		m := domain.ModelRef{Provider: "ollama", Name: "qwen:4b", Digest: "test"}
		p.address = modelAddress{"", s.ID}
		p.loaded = true
		p.settings.Session = m
		p.settings.Default.Model = m
		p.options = []domain.ModelOption{{Model: m, Nodes: []string{"worker"}}}
		g.openModels()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		labels := map[string]bool{}
		for _, target := range g.bodyTargets() {
			labels[focusLabel(target.w)] = true
		}
		if !labels["Cancel"] || !labels["Save for this chat"] || labels["Set as global default"] {
			t.Fatal("incorrect member controls", labels)
		}
		g.planningBack()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if g.planning.current != s || g.globalDraft != "Draft to keep" || g.modelLabel() != m.Name {
			t.Fatal("lost originating chat or model")
		}
		g.Close()
	}
}

func TestPickerCatalogUsesStableTextModelIdentities(t *testing.T) {
	now := time.Now()
	v := domain.WorkerInventory{Observed: now.Unix(), Version: "test", Models: []domain.InventoryModel{{Name: "Z", Digest: "same", Text: true}, {Name: "b", Digest: "one", Text: true}, {Name: "image", Digest: "x"}}}
	other := v
	other.Models = []domain.InventoryModel{{Name: "Z", Digest: "same", Text: true, Loaded: true}, {Name: "b", Digest: "two", Text: true}}
	old := v
	old.Observed = now.Add(-time.Hour).Unix()
	catalog := domain.ChatModelCatalog(v, domain.PoolDevice{ID: "a", Host: "first"}, []domain.ClusterDevice{{Host: "second", Inventory: &other}, {Host: "offline", Inventory: &old}}, now)
	if len(catalog) != 3 || catalog[0].Model.Digest != "one" || catalog[1].Model.Digest != "two" || strings.Join(catalog[2].Nodes, ",") != "first,second" {
		t.Fatal(catalog)
	}
}

func TestBubbleActionsAndReturn(t *testing.T) {
	for _, size := range [][2]int{{390, 844}, {1920, 1080}} {
		g := horizontalGame()
		g.Layout(size[0], size[1])
		selectTabThroughStrip(t, g, hud.TabGlobalSession)
		s := g.planning.current
		source := "**Keep**  both spaces\n    and indentation."
		s.Turns = []domain.ChatTurn{{ID: "turn", Input: "question", Reply: source, Steps: []domain.ChatStep{{Output: &domain.ChatOutput{Finish: "stop", Metrics: map[string]float64{"eval_count": 5, "eval_duration": 1e9}}}}}}
		var copied string
		g.runtimeInfo.CopyText = func(s string) {
			copied = s
		}
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		var button *widget.Button
		for _, f := range g.bodyTargets() {
			if strings.HasPrefix(focusLabel(f.w), "**Keep**") {
				button = f.w.(*widget.Button)
			}
			if focusLabel(f.w) == "Copy message" {
				t.Fatal("actions visible before activation")
			}
		}
		if button == nil {
			t.Fatal("missing message")
		}
		button.Click()
		event.ExecuteDeferred()
		tools := g.planning.bubbles.current
		if tools == nil || tools.row.GetWidget().GetVisibility() != widget.Visibility_Show {
			t.Fatal("missing message actions")
		}
		tools.copy()
		if copied != source {
			t.Fatal("copy changed original text")
		}
		g.globalDraft = "draft"
		tools.open()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if g.planning.page != "bubble-info" || len(g.planning.info.steps) != 1 {
			t.Fatal("missing info page")
		}
		g.planningBack()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if g.planning.current != s || g.globalDraft != "draft" {
			t.Fatal("back lost conversation")
		}
		g.Close()
	}
}

func TestPreviewAboveGear(t *testing.T) {
	for _, size := range [][2]int{{390, 844}, {1920, 1080}} {
		g := horizontalGame()
		g.Layout(size[0], size[1])
		selectTabThroughStrip(t, g, hud.TabGlobalSession)
		s := g.planning.current
		s.Turns = []domain.ChatTurn{{ID: "turn", Input: "test", Steps: []domain.ChatStep{{Job: "job"}}}}
		s.acknowledged = true
		g.globalDraft = "next question"
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		p := &g.preview
		if p.body == nil || p.body.GetWidget().GetVisibility() != widget.Visibility_Hide {
			t.Fatal("queued preview should be hidden")
		}
		p.job, p.stage = "job", "Evaluating prompt"
		g.paintChatPreview()
		field := g.globalField
		p.stage, p.text = "Generating reply", "First words of the reply."
		g.paintChatPreview()
		renderNavigation(g)
		if p.body.Label != p.text || p.status.Text().Label != p.stage || p.body.GetWidget().Rect.Max.Y > p.status.GetWidget().Rect.Min.Y {
			t.Fatal("preview did not appear above progress row", size, p.body.GetWidget().Rect, p.status.GetWidget().Rect)
		}
		if g.globalField != field || g.globalDraft != "next question" {
			t.Fatal("preview reset the composer")
		}
		g.Close()
	}
}
