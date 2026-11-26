//go:build gui

// Verifies the New Project focus sequence against the rendered HUD.
// Both the inline editor and the native prompt use the shared controller route.
package gui

import (
	"image"
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

func TestNewProjectFocusSequence(t *testing.T) {
	for _, width := range []int{320, 960} {
		g := NewGame()
		g.projects.store = layoutRegistry{}
		g.serviceProjection.Devices = []domain.ClusterDevice{
			{Host: "workstation", Fingerprint: "one"},
			{Host: "workbench", Fingerprint: "two"},
		}
		g.Layout(width, 720)
		selectTabThroughStrip(t, g, hud.TabProjects)
		g.selectSection(hud.TabProjects, 1)
		renderNavigation(g)
		search, ok := g.ui.GetFocusedWidget().(*widget.TextInput)
		if !ok || g.nav.target.w != search || g.projects.wizard.host != "" {
			t.Fatalf("width %d: new form did not focus search", width)
		}
		for _, query := range []string{"work", "missing", " "} {
			search.SetText(query)
			event.ExecuteDeferred()
			g.rebuildUI(g.shell.Snapshot())
			renderNavigation(g)
			if g.projects.wizard.host != "" || g.ui.GetFocusedWidget() != search {
				t.Fatalf("query %q changed selection or editing focus", query)
			}
		}
		search.SetText("workb")
		event.ExecuteDeferred()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if g.projects.wizard.host != "worker-two" || g.ui.GetFocusedWidget() != nil || focusLabel(g.nav.target.w) != "Create a new repository" {
			t.Fatal("one match did not select the host and hand focus to source")
		}
		g.stepFocus(1)
		if focusLabel(g.nav.target.w) != "Bring in a repository" {
			t.Fatal("Tab did not advance to the next source option")
		}
		g.moveFocus(image.Pt(0, -1))
		if focusLabel(g.nav.target.w) != "Create a new repository" {
			t.Fatal("Up did not return to the first source option")
		}
		g.activateFocus()
		event.ExecuteDeferred()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		if focusLabel(g.nav.target.w) != "✓ Agentic Pipelines starter (default)" {
			t.Fatal("source activation did not focus the template group")
		}
		g.stepFocus(1)
		g.activateFocus()
		event.ExecuteDeferred()
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		_, ok = g.nav.target.w.(*widget.TextInput)
		if g.projects.wizard.mode != "empty" || !ok {
			t.Fatal("template activation did not focus the repository name")
		}
		g.Close()
	}
}

func TestNewProjectNativeSearchInitialFocus(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.runtimeInfo.PromptValue = func(string, string, bool) (string, bool) {
		return "", false
	}
	g.projects.store = layoutRegistry{}
	selectTabThroughStrip(t, g, hud.TabProjects)
	g.selectSection(hud.TabProjects, 1)
	renderNavigation(g)
	if focusLabel(g.nav.target.w) != "Tap to enter…" || g.projects.pending != nil {
		t.Fatal("native search control should be focused and ready for activation")
	}
}

func TestRemoteCloneAcceptsRepositoryAtActivation(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.projects.wizard.host = "worker-example"
	g.projects.wizard.mode = "clone"
	action := g.remoteCloneOption()
	if action == nil {
		t.Fatal("clone must be available before the field receives keyboard input")
	}
	action.run()
	if g.projects.message == "" || g.projects.pending != nil {
		t.Fatal("empty repository requires inline guidance")
	}
}

func TestLockedKeyringCloneChooser(t *testing.T) {
	const oldReason = "Remote enrollment requires an unlocked Linux Secret Service and a private memory-backed staging directory."
	for _, width := range []int{390, 1280} {
		g := NewGame()
		g.Layout(width, 800)
		w := &g.projects.wizard
		w.host, w.mode = "worker-example", "clone"
		p := domain.GitProvider{ID: "github", Name: "GitHub", Installed: true, Accounts: []string{"saved"}}
		for _, legacy := range []bool{false, true} {
			problem := "The worker’s default keyring is locked."
			if legacy {
				problem = oldReason
				p.EnrollmentReason = ""
				p.Reason = problem
			} else {
				p.EnrollmentReason = problem
			}
			w.accountHost, w.providers = w.host, []domain.GitProvider{p}
			pickerText := func() string {
				c := planningColumn()
				g.hostAccountPicker(c, w.host, &w.remoteAccount)
				var out []string
				for _, child := range c.Children() {
					label, ok := child.(*widget.Text)
					if ok {
						out = append(out, label.Label)
					}
				}
				return strings.Join(out, " ")
			}
			before := pickerText()
			if strings.Contains(before, problem) || !strings.Contains(before, "public access") || g.remoteCloneOption() == nil {
				t.Fatal("enrollment interfered with clone", before)
			}
			options := map[string]accountOption{}
			for _, o := range g.accountOptions(w.host, &w.remoteAccount) {
				options[o.id] = o
			}
			_, ok := options["add-github"]
			if ok {
				t.Fatal("secure enrollment offered while locked")
			}
			for _, id := range []string{"public", "github-saved", "enrollment-help-github"} {
				if options[id].run == nil {
					t.Fatal("missing action", id)
				}
			}
			options["enrollment-help-github"].run()
			if !strings.Contains(pickerText(), problem) {
				t.Fatal("missing enrollment explanation")
			}
			options["github-saved"].run()
			if w.remoteAccount != "saved" || w.enrollmentHelp != "" || g.remoteCloneOption() == nil {
				t.Fatal("saved account unavailable")
			}
			options["public"].run()
			if w.remoteAccount != "" {
				t.Fatal("anonymous selection unavailable")
			}
		}
		g.Close()
	}
}
